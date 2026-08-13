package decoyconfig

import (
	"strings"
	"testing"
)

func TestParseCredentialAndRender(t *testing.T) {
	spec, err := Parse("credential", []byte(`{"path":"/opt/app/.env","username":"backup","password":"fake-secret","create_parent":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Mode != "0600" || !spec.CreateParent {
		t.Fatalf("unexpected normalized spec: %#v", spec)
	}
	content := string(Render("credential", "decoy-1", "Database backup", spec))
	if !strings.Contains(content, "DB_USERNAME=backup") || !strings.Contains(content, "honeynet-decoy-decoy-1") {
		t.Fatalf("unexpected rendered content: %q", content)
	}
}

func TestParseNetworkToken(t *testing.T) {
	spec, err := Parse("network", []byte(`{"token":"db-backup-token-2026","description":"嵌入连接串"}`))
	if err != nil || spec.Token != "db-backup-token-2026" {
		t.Fatalf("Parse(network) = %#v, %v", spec, err)
	}
}

func TestParseAcceptsWindowsPathOnAnyServerPlatform(t *testing.T) {
	spec, err := Parse("file", []byte(`{"path":"C:\\ProgramData\\Honeynet\\decoys\\report.txt"}`))
	if err != nil || spec.Path != `C:\ProgramData\Honeynet\decoys\report.txt` {
		t.Fatalf("Parse(Windows path) = %#v, %v", spec, err)
	}
	if _, err := Parse("file", []byte(`{"path":"C:\\"}`)); err == nil {
		t.Fatal("Parse accepted a Windows volume root")
	}
}

func TestParseRejectsUnsafeConfigs(t *testing.T) {
	tests := map[string]struct {
		kind string
		raw  string
	}{
		"relative path":  {"file", `{"path":"relative.txt"}`},
		"root path":      {"file", `{"path":"/"}`},
		"unsafe mode":    {"file", `{"path":"/tmp/decoy","mode":"0777"}`},
		"unknown field":  {"file", `{"path":"/tmp/decoy","command":"id"}`},
		"missing secret": {"credential", `{"path":"/tmp/.env","username":"admin"}`},
		"short token":    {"network", `{"token":"short"}`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(test.kind, []byte(test.raw)); err == nil {
				t.Fatal("Parse accepted unsafe config")
			}
		})
	}
}
