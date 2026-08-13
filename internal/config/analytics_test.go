package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadAnalyticsSeparateConfigAndSecretOverride(t *testing.T) {
	t.Setenv("HONEYPOT_ANALYTICS_CONFIG", "")
	t.Setenv("HONEYPOT_ANALYTICS_ENABLED", "")
	t.Setenv("HONEYPOT_CLICKHOUSE_DSN", "clickhouse://secret@127.0.0.1:9000/honeynet_analytics")
	dir := t.TempDir()
	serverPath := filepath.Join(dir, "server.yaml")
	if err := os.WriteFile(serverPath, []byte("server: {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	data := `analytics:
  enabled: false
  dsn: "clickhouse://file-value"
  database: "honeynet_analytics"
  table: "security_events"
  max_open_conns: 12
  max_idle_conns: 4
  conn_max_lifetime: "30m"
  dial_timeout: "3s"
  read_timeout: "20s"
`
	if err := os.WriteFile(filepath.Join(dir, "analytics.yaml"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadAnalytics(serverPath)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || !strings.Contains(cfg.DSN, "secret") || cfg.MaxOpenConns != 12 || cfg.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("unexpected analytics config: %#v", cfg)
	}
}

func TestLoadAnalyticsMissingDerivedFileIsDisabled(t *testing.T) {
	t.Setenv("HONEYPOT_ANALYTICS_CONFIG", "")
	t.Setenv("HONEYPOT_ANALYTICS_ENABLED", "")
	t.Setenv("HONEYPOT_CLICKHOUSE_DSN", "")
	cfg, err := LoadAnalytics(filepath.Join(t.TempDir(), "server.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled || cfg.Database != "honeynet_analytics" {
		t.Fatalf("unexpected fallback config: %#v", cfg)
	}
}

func TestLoadAnalyticsRejectsUnknownField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "analytics.yaml")
	if err := os.WriteFile(path, []byte("analytics:\n  mystery: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HONEYPOT_ANALYTICS_CONFIG", path)
	if _, err := LoadAnalytics(""); err == nil {
		t.Fatal("expected strict analytics config error")
	}
}
