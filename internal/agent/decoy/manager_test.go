package decoy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func fileTarget(id, path string) protocol.DecoyTarget {
	return protocol.DecoyTarget{ID: id, Name: "Quarterly report", Type: "file", Status: "enabled", Config: map[string]any{
		"path": path, "marker": "qa-decoy", "create_parent": true,
	}}
}

func TestManagerCreatesMonitorsAndSafelyRemovesFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "documents", "salary.xlsx")
	events := make(chan protocol.Event, 4)
	manager, err := New(filepath.Join(root, "state"), func(event protocol.Event) { events <- event })
	if err != nil {
		t.Fatal(err)
	}
	target := fileTarget("decoy-1", path)
	result := manager.Apply(context.Background(), []protocol.DecoyTarget{target})
	if len(result) != 1 || !result[0].Success || result[0].Status != "monitoring" {
		t.Fatalf("Apply() = %#v", result)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("touched"), 0644); err != nil {
		t.Fatal(err)
	}
	select {
	case event := <-events:
		if event.DecoyID != "decoy-1" || event.EventType != "decoy.file" {
			t.Fatalf("unexpected event: %#v", event)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("decoy modification did not emit an event")
	}
	manager.Apply(context.Background(), nil)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("modified evidence should be preserved: %v", err)
	}
}

func TestManagerPersistsOwnershipAcrossRestart(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "decoy.txt")
	target := fileTarget("decoy-2", path)
	manager, err := New(filepath.Join(root, "state"), func(protocol.Event) {})
	if err != nil {
		t.Fatal(err)
	}
	if result := manager.Apply(context.Background(), []protocol.DecoyTarget{target}); !result[0].Success {
		t.Fatal(result[0].LastError)
	}
	manager.Shutdown()
	restarted, err := New(filepath.Join(root, "state"), func(protocol.Event) {})
	if err != nil {
		t.Fatal(err)
	}
	if result := restarted.Apply(context.Background(), []protocol.DecoyTarget{target}); !result[0].Success {
		t.Fatalf("restart reconciliation failed: %#v", result)
	}
	restarted.Apply(context.Background(), []protocol.DecoyTarget{{ID: target.ID, Name: target.Name, Type: target.Type, Status: "disabled", Config: target.Config}})
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("unchanged Agent-owned file was not removed: %v", err)
	}
}

func TestManagerDoesNotOverwriteExistingFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "existing.txt")
	if err := os.WriteFile(path, []byte("production"), 0644); err != nil {
		t.Fatal(err)
	}
	manager, _ := New(filepath.Join(root, "state"), func(protocol.Event) {})
	result := manager.Apply(context.Background(), []protocol.DecoyTarget{fileTarget("decoy-3", path)})
	if result[0].Success || result[0].Status != "error" {
		t.Fatalf("existing file was accepted without monitor_existing: %#v", result)
	}
	content, _ := os.ReadFile(path)
	if string(content) != "production" {
		t.Fatalf("existing file was modified: %q", content)
	}
}

func TestNetworkDecoyIsPassive(t *testing.T) {
	manager, _ := New(t.TempDir(), func(protocol.Event) {})
	target := protocol.DecoyTarget{ID: "network-1", Name: "Backup URL", Type: "network", Status: "enabled", Config: map[string]any{"token": "backup-token-2026"}}
	result := manager.Apply(context.Background(), []protocol.DecoyTarget{target})
	if !result[0].Success || result[0].Status != "passive" {
		t.Fatalf("network Apply() = %#v", result)
	}
}

func TestManagerQuarantinesInvalidManifest(t *testing.T) {
	stateDir := t.TempDir()
	manifestPath := filepath.Join(stateDir, "decoys", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte("{not-json"), 0600); err != nil {
		t.Fatal(err)
	}
	manager, err := New(stateDir, func(protocol.Event) {})
	if err != nil {
		t.Fatalf("New() should recover from an invalid manifest: %v", err)
	}
	if len(manager.owned) != 0 {
		t.Fatalf("invalid ownership data must not be trusted: %#v", manager.owned)
	}
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Fatalf("invalid manifest was not quarantined: %v", err)
	}
	quarantined, err := filepath.Glob(manifestPath + ".corrupt-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(quarantined) != 1 {
		t.Fatalf("quarantined manifests = %v, want exactly one", quarantined)
	}
}
