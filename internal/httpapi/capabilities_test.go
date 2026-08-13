package httpapi

import (
	"encoding/json"
	"testing"

	"github.com/honeynet/honeynet/internal/store"
)

func TestNormalizeCapabilities(t *testing.T) {
	raw := normalizeCapabilities([]string{" pot.SSH ", "pot.ssh", "sense.passive", "invalid capability", "POT.DNS"})
	var got []string
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	want := []string{"pot.dns", "pot.ssh", "sense.passive"}
	if len(got) != len(want) {
		t.Fatalf("normalizeCapabilities() = %v", got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("normalizeCapabilities() = %v, want %v", got, want)
		}
	}
}

func TestNodeSupportsService(t *testing.T) {
	unknown, supported := nodeSupportsService(store.Node{}, "ssh")
	if unknown || supported {
		t.Fatalf("empty capability state = %v, %v", unknown, supported)
	}
	node := store.Node{Capabilities: normalizeCapabilities([]string{"pot.ssh", "decoy.file"})}
	if known, ok := nodeSupportsService(node, "ssh"); !known || !ok {
		t.Fatalf("ssh support = %v, %v", known, ok)
	}
	if known, ok := nodeSupportsService(node, "ftp"); !known || ok {
		t.Fatalf("ftp support = %v, %v", known, ok)
	}
}
