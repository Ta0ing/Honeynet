package httpapi

import (
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/config"
	"github.com/honeynet/honeynet/internal/nodepki"
)

func TestNativeTLSInstallCommandsBootstrapPinnedConsoleCA(t *testing.T) {
	authority, err := nodepki.LoadOrCreate(t.TempDir(), []string{"127.0.0.1"}, 400*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	api := &API{cfg: config.Config{PublicURL: "https://127.0.0.1:8080", AgentPublicURL: "https://127.0.0.1:8443", TLSEnabled: true}, pki: authority}
	commands := api.installCommands("node-id", "one-time-token")
	linux := commands["linux"].(string)
	for _, required := range []string{"base64 -d", "--cacert", "--console-ca", "/download/install-agent.sh"} {
		if !strings.Contains(linux, required) {
			t.Fatalf("native TLS Linux command missing %q: %s", required, linux)
		}
	}
	if strings.Contains(linux, " -k") || strings.Contains(linux, "--insecure") || strings.Contains(linux, "| sudo sh") {
		t.Fatalf("native TLS bootstrap weakened transport or used pipe-to-shell: %s", linux)
	}
	windows := commands["windows"].(string)
	for _, required := range []string{"FromBase64String", "Import-Certificate", "CurrentUser\\Root", "-OutFile $script"} {
		if !strings.Contains(windows, required) {
			t.Fatalf("native TLS Windows command missing %q: %s", required, windows)
		}
	}
	if strings.Contains(windows, "Invoke-Expression") {
		t.Fatalf("Windows command uses Invoke-Expression: %s", windows)
	}
}

func TestExternalTLSInstallCommandsUseSystemTrustWithoutEmbeddingAgentCA(t *testing.T) {
	authority, err := nodepki.LoadOrCreate(t.TempDir(), []string{"127.0.0.1"}, 400*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	api := &API{cfg: config.Config{
		PublicURL: "https://console.example", AgentPublicURL: "https://console.example:8443",
		TLSEnabled: true, TLSCertFile: "/etc/honeynet/console/fullchain.pem", TLSKeyFile: "/etc/honeynet/console/privkey.pem",
	}, pki: authority}
	commands := api.installCommands("node-id", "one-time-token")
	linux := commands["linux"].(string)
	for _, forbidden := range []string{"base64 -d", "--cacert", "--console-ca"} {
		if strings.Contains(linux, forbidden) {
			t.Fatalf("external TLS Linux command contains private Console bootstrap %q: %s", forbidden, linux)
		}
	}
	if !strings.Contains(linux, "curl -fsSL") || !strings.Contains(linux, "/download/install-agent.sh") {
		t.Fatalf("external TLS Linux command does not use system-trusted HTTPS: %s", linux)
	}
	if !strings.Contains(linux, "--ca-sha256") || !strings.Contains(linux, authority.CAFingerprint()) {
		t.Fatalf("external Console TLS command lost Agent gateway CA pinning: %s", linux)
	}
	windows := commands["windows"].(string)
	for _, forbidden := range []string{"FromBase64String", "Import-Certificate", "CurrentUser\\Root"} {
		if strings.Contains(windows, forbidden) {
			t.Fatalf("external TLS Windows command embeds Agent CA via %q: %s", forbidden, windows)
		}
	}
	if !strings.Contains(windows, "Invoke-WebRequest") {
		t.Fatalf("external TLS Windows command does not use system-trusted HTTPS: %s", windows)
	}
	if !strings.Contains(windows, "-CASHA256") || !strings.Contains(windows, authority.CAFingerprint()) {
		t.Fatalf("external Console TLS Windows command lost Agent gateway CA pinning: %s", windows)
	}
}

func TestHTTPConsoleStillProvidesInstallCommands(t *testing.T) {
	authority, err := nodepki.LoadOrCreate(t.TempDir(), []string{"127.0.0.1"}, 400*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	api := &API{cfg: config.Config{
		PublicURL: "http://192.0.2.10:8080", AgentPublicURL: "https://192.0.2.10:8443",
	}, pki: authority}
	commands := api.installCommands("node-id", "one-time-token")
	if available, _ := commands["available"].(bool); !available {
		t.Fatalf("HTTP Console unexpectedly disabled node installation: %#v", commands)
	}
	linux, _ := commands["linux"].(string)
	for _, required := range []string{"http://192.0.2.10:8080", "/download/install-agent.sh", "--ca-sha256", "https://192.0.2.10:8443"} {
		if !strings.Contains(linux, required) {
			t.Fatalf("HTTP Console Linux command missing %q: %s", required, linux)
		}
	}
	for _, forbidden := range []string{"一键安装已禁用", "HTTPS_REQUIRED", "| sudo sh"} {
		if strings.Contains(linux, forbidden) {
			t.Fatalf("HTTP Console Linux command contains forbidden marker %q: %s", forbidden, linux)
		}
	}
	if enabled, _ := commands["console_tls_enabled"].(bool); enabled {
		t.Fatal("HTTP Console incorrectly reported TLS enabled")
	}
}
