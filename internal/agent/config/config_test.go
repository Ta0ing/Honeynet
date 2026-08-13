package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveEnrollmentPersistsPrivateMaterialWithRestrictedPermissions(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ServerURL: "http://server.example:8080", AgentURL: "https://server.example:8443", NodeID: "node-1",
		RegistrationToken: "one-time-token", StateDir: filepath.Join(dir, "state"), ConfigPath: filepath.Join(dir, "agent.json"),
	}
	if err := cfg.Normalize(); err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(30 * 24 * time.Hour).UTC().Round(time.Second)
	if err := cfg.SaveEnrollment(cfg.AgentURL, []byte("ca"), []byte("cert"), []byte("private-key"), expires, 7*24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if cfg.RegistrationToken != "" || !cfg.HasCertificate() {
		t.Fatalf("enrollment state was not finalized: %#v", cfg)
	}
	info, err := os.Stat(cfg.ClientKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("client key mode is %o, want 600", info.Mode().Perm())
	}
	loaded, err := Load(cfg.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.RegistrationToken != "" || loaded.ClientKeyPath != cfg.ClientKeyPath || !loaded.CertificateExpiry.Equal(expires) {
		t.Fatalf("unexpected saved enrollment: %#v", loaded)
	}
}

func TestNormalizeSupportsIPv6ServerAndAgentURLs(t *testing.T) {
	cfg := Config{
		ServerURL: "http://[2001:4860:4860::8888]:8080/",
		NodeID:    "node-ipv6", ConfigPath: filepath.Join(t.TempDir(), "agent.json"),
	}
	if err := cfg.Normalize(); err != nil {
		t.Fatal(err)
	}
	if cfg.ServerURL != "http://[2001:4860:4860::8888]:8080" {
		t.Fatalf("normalized Server URL = %q", cfg.ServerURL)
	}
	if cfg.AgentURL != "https://[2001:4860:4860::8888]:8443" {
		t.Fatalf("derived Agent URL = %q", cfg.AgentURL)
	}
}

func TestNormalizeRejectsUnbracketedIPv6URL(t *testing.T) {
	cfg := Config{ServerURL: "http://2001:4860:4860::8888:8080", NodeID: "node-ipv6"}
	if err := cfg.Normalize(); err == nil {
		t.Fatal("Normalize accepted an unbracketed IPv6 URL")
	}
}
