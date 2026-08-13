package agentupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestSignerPersistsAndRejectsTampering(t *testing.T) {
	dir := t.TempDir()
	signer, err := LoadOrCreateSigner(dir)
	if err != nil {
		t.Fatal(err)
	}
	descriptor := Descriptor{Version: "1.2.3", OS: "linux", Arch: "amd64", SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Size: 42}
	signature, err := signer.Sign(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(signer.PublicKey(), signature, descriptor); err != nil {
		t.Fatal(err)
	}
	descriptor.Size++
	if err := Verify(signer.PublicKey(), signature, descriptor); err == nil {
		t.Fatal("tampered descriptor passed signature verification")
	}
	reloaded, err := LoadOrCreateSigner(dir)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.KeyID() != signer.KeyID() {
		t.Fatal("update signing key changed after reload")
	}
}

func TestInstallConfirmAndRollback(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{name: "healthy update", mode: "confirm"},
		{name: "failed startup rolls back", mode: "wrong-version"},
		{name: "health deadline rolls back", mode: "deadline"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			current := filepath.Join(dir, "honeynet-agent")
			oldBinary := []byte("old signed Agent binary")
			newBinary := []byte("new signed Agent binary")
			if err := os.WriteFile(current, oldBinary, 0755); err != nil {
				t.Fatal(err)
			}
			signer, err := LoadOrCreateSigner(filepath.Join(dir, "pki"))
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(newBinary)
			descriptor := Descriptor{Version: "2.0.0", OS: runtime.GOOS, Arch: runtime.GOARCH, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(newBinary))}
			signature, err := signer.Sign(descriptor)
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { _, _ = response.Write(newBinary) }))
			defer server.Close()
			manager := &Manager{StateDir: filepath.Join(dir, "state"), CurrentVersion: "1.0.0", PublicKey: signer.PublicKey(), DownloadBaseURL: server.URL, ExecutablePath: current}
			command := Command{TaskID: "task-1", RolloutID: "rollout-1", Version: descriptor.Version, OS: descriptor.OS, Arch: descriptor.Arch, URL: "/agent/v1/updates/test", SHA256: descriptor.SHA256, Signature: signature, Size: descriptor.Size}
			if err := manager.Install(context.Background(), server.Client(), command); !errors.Is(err, ErrRestartRequired) {
				t.Fatalf("install returned %v", err)
			}
			installed, err := os.ReadFile(current)
			if err != nil || string(installed) != string(newBinary) {
				t.Fatalf("new binary was not installed: %v %q", err, installed)
			}
			if test.mode == "confirm" {
				started := &Manager{StateDir: manager.StateDir, CurrentVersion: "2.0.0", PublicKey: signer.PublicKey(), ExecutablePath: current}
				if err := started.CheckStartup(); err != nil {
					t.Fatal(err)
				}
				if err := started.ConfirmHealthy(); err != nil {
					t.Fatal(err)
				}
				if state := started.Status(); state == nil || state.Status != "healthy" {
					t.Fatalf("unexpected update state: %#v", state)
				}
				return
			}
			if test.mode == "deadline" {
				stalled := &Manager{StateDir: manager.StateDir, CurrentVersion: "2.0.0", PublicKey: signer.PublicKey(), ExecutablePath: current}
				state := stalled.Status()
				state.HealthDeadline = time.Now().Add(-time.Second)
				if err := stalled.saveState(*state); err != nil {
					t.Fatal(err)
				}
				if err := stalled.WaitForHealth(context.Background()); !errors.Is(err, ErrRollbackRequired) {
					t.Fatalf("health watchdog returned %v", err)
				}
			} else {
				failed := &Manager{StateDir: manager.StateDir, CurrentVersion: "1.0.0", PublicKey: signer.PublicKey(), ExecutablePath: current}
				if err := failed.CheckStartup(); !errors.Is(err, ErrRollbackRequired) {
					t.Fatalf("startup returned %v", err)
				}
			}
			restored, err := os.ReadFile(current)
			if err != nil || string(restored) != string(oldBinary) {
				t.Fatalf("old binary was not restored: %v %q", err, restored)
			}
		})
	}
}
