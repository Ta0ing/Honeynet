package store

import (
	"path/filepath"
	"testing"

	"github.com/honeynet/honeynet/internal/agent/pots"
)

func TestPotServiceCatalogContract(t *testing.T) {
	seeds := potServiceSeeds()
	if len(seeds) != 111 {
		t.Fatalf("pot service catalog contains %d services, want 111; update the product count deliberately", len(seeds))
	}
	seen := make(map[string]struct{}, len(seeds))
	for _, seed := range seeds {
		if seed.code == "" || seed.name == "" || seed.category == "" || seed.port < 1 || seed.port > 65535 {
			t.Fatalf("invalid pot service catalog entry: %#v", seed)
		}
		if seed.protocol != "tcp" && seed.protocol != "udp" {
			t.Fatalf("service %s has invalid protocol %q", seed.code, seed.protocol)
		}
		if _, exists := seen[seed.code]; exists {
			t.Fatalf("duplicate pot service code %q", seed.code)
		}
		seen[seed.code] = struct{}{}
	}
}

func TestAgentSupportedPotsExistInServiceCatalog(t *testing.T) {
	catalog := make(map[string]struct{}, len(potServiceSeeds()))
	for _, seed := range potServiceSeeds() {
		catalog[seed.code] = struct{}{}
	}
	templateRoot := filepath.Join("..", "..", "honeypot-templates-server", "services")
	supported := pots.SupportedCodesAt(templateRoot)
	if len(supported) != 103 {
		t.Fatalf("Agent service coverage is %d services, want 103 with the supplied resource pack", len(supported))
	}
	for _, code := range supported {
		if _, exists := catalog[code]; !exists {
			t.Errorf("Agent supports %q but the service catalog does not contain it", code)
		}
	}
}

func TestGeneratedWebProfilesAreNotCurrentCatalogServices(t *testing.T) {
	for _, code := range []string{"kingdee", "apache", "phpmyadmin", "edr", "fanwei-oa"} {
		if IsCurrentPotService(code) {
			t.Errorf("retired generated Web profile %q remains current", code)
		}
	}
	for _, code := range []string{"tomcat", "phpadmin", "edr-sangfor", "oa-tongda", "synology-nas"} {
		if !IsCurrentPotService(code) {
			t.Errorf("Web resource Web service %q is missing", code)
		}
	}
}
