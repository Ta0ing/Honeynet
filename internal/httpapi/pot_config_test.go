package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizePotConfigIPv4AndIPv6(t *testing.T) {
	for input, expected := range map[string]string{
		`{}`:                       "",
		`{"bind":"0.0.0.0"}`:       "0.0.0.0",
		`{"bind":"[2001:db8::1]"}`: "2001:db8::1",
		`{"bind":"::"}`:            "::",
	} {
		output, err := normalizePotConfig(json.RawMessage(input))
		if err != nil {
			t.Fatalf("normalize %s: %v", input, err)
		}
		var config map[string]any
		_ = json.Unmarshal(output, &config)
		if got, _ := config["bind"].(string); got != expected {
			t.Fatalf("normalize %s bind=%q want %q", input, got, expected)
		}
	}
}

func TestNormalizePotConfigRejectsInvalidOrOversizedInput(t *testing.T) {
	for _, input := range []json.RawMessage{json.RawMessage(`[]`), json.RawMessage(`{"bind":"all"}`), json.RawMessage(`{"bind":42}`), json.RawMessage(strings.Repeat("x", maxPotConfigBytes+1))} {
		if _, err := normalizePotConfig(input); err == nil {
			t.Fatalf("expected invalid config %q to fail", string(input[:min(len(input), 32)]))
		}
	}
}
