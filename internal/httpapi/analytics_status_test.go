package httpapi

import "testing"

func TestPublicAnalyticsVersionDoesNotExposeBackendBanner(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ClickHouse (node-name) server version 25.8.28 revision 54479 (timezone UTC)", "25.8.28"},
		{"vendor server version 1.2.3-beta revision 4", ""},
		{"implementation-specific banner", ""},
		{"server version unsafe/value", ""},
	}
	for _, test := range tests {
		if got := publicAnalyticsVersion(test.input); got != test.want {
			t.Fatalf("publicAnalyticsVersion(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
