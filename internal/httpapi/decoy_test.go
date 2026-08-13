package httpapi

import (
	"testing"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

func TestNetworkDecoyMatch(t *testing.T) {
	decoy := store.Decoy{Type: "network", Config: datatypes.JSON(`{"token":"backup-db-token-2026","description":"fake DSN"}`)}
	spec, matched, err := networkDecoyMatch(decoy, []byte(`{"path":"/?dsn=backup-db-token-2026"}`))
	if err != nil || !matched || spec.Token != "backup-db-token-2026" {
		t.Fatalf("networkDecoyMatch() = %#v, %v", spec, matched)
	}
	if _, matched, err := networkDecoyMatch(decoy, []byte(`{"path":"/normal"}`)); err != nil || matched {
		t.Fatal("networkDecoyMatch accepted unrelated payload")
	}
}
