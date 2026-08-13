package client

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	agentconfig "github.com/honeynet/honeynet/internal/agent/config"
	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func TestSenseConfigPersistence(t *testing.T) {
	dir := t.TempDir()
	client := &Client{cfg: &agentconfig.Config{StateDir: dir}}
	expected := protocol.SenseConfig{Enabled: true, Interface: "eth0", TCPEnabled: true, DistinctPorts: 8, WindowSeconds: 12, CooldownSeconds: 90, ExcludedPorts: []int{22, 443}, IgnoredCIDRs: []string{"10.0.0.0/8"}}
	if err := client.saveSenseConfig(expected); err != nil {
		t.Fatal(err)
	}
	actual, exists, err := client.loadSenseConfig()
	if err != nil || !exists || !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected persisted config: %#v %v %v", actual, exists, err)
	}
	info, err := os.Stat(filepath.Join(dir, "sense-config.json"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("sense config mode is not 0600: %v %v", info, err)
	}
}
