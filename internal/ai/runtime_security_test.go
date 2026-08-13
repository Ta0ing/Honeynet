package ai

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeConfigRequiresHTTPSForRemoteProvider(t *testing.T) {
	base := Config{Enabled: true, APIKey: "secret", Model: "model", Timeout: time.Second}
	for _, value := range []string{"http://api.example/v1", "http://192.0.2.10:8080/v1", "http://[2001:db8::10]:8080/v1"} {
		candidate := base
		candidate.BaseURL = value
		if _, err := NormalizeConfig(candidate); err == nil || !strings.Contains(err.Error(), "HTTPS") {
			t.Fatalf("remote plaintext provider %q returned %v", value, err)
		}
	}
}

func TestNormalizeConfigAllowsHTTPSAndLoopbackHTTP(t *testing.T) {
	base := Config{Enabled: true, APIKey: "secret", Model: "model", Timeout: time.Second}
	for _, value := range []string{"https://api.example/v1", "http://localhost:8080/v1", "http://localhost.:8080/v1", "http://127.0.0.1:8080/v1", "http://[::1]:8080/v1"} {
		candidate := base
		candidate.BaseURL = value
		if _, err := NormalizeConfig(candidate); err != nil {
			t.Fatalf("secure/local provider %q rejected: %v", value, err)
		}
	}
}
