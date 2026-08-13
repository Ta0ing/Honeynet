package threatintel

import (
	"os"
	"reflect"
	"testing"
	"time"
)

// TestPublisherDatabaseCompatibility is opt-in because the publisher database
// is large, encrypted and intentionally not committed to this repository. CI
// and release smoke tests can point at a freshly downloaded/decrypted .db file.
func TestPublisherDatabaseCompatibility(t *testing.T) {
	path := os.Getenv("HONEYPOT_TEST_THREAT_INTEL_DB")
	if path == "" {
		t.Skip("HONEYPOT_TEST_THREAT_INTEL_DB is not configured")
	}
	database, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if database.Count() < 100_000 || !database.SlowMode() || database.SlowKeySize() != 3 {
		t.Fatalf("publisher database metadata is unexpected: count=%d slow=%v key_size=%d", database.Count(), database.SlowMode(), database.SlowKeySize())
	}
	want := Result{
		Labels:    []string{"Hosting"},
		Behaviors: []string{"Bruteforce", "Brute Force", "Spammers", "Information Gathering", "Command Injection", "SSH Bruteforce", "Port Scanning", "Vulnerability Scanning", "Spider", "Abuse", "Proxy", "Exploit"},
		Level:     3,
	}
	started := time.Now()
	got, found := database.Lookup("62.210.142.161")
	if !found || !reflect.DeepEqual(got, want) {
		t.Fatalf("publisher database lookup = %#v, %v; want %#v, true", got, found, want)
	}
	t.Logf("publisher database first hit lookup took %s", time.Since(started))
	started = time.Now()
	if _, found := database.Lookup("192.0.2.254"); found {
		t.Fatal("unexpected publisher database match for documentation address")
	}
	t.Logf("publisher database first miss lookup took %s", time.Since(started))
	started = time.Now()
	_, _ = database.Lookup("192.0.2.254")
	t.Logf("publisher database cached miss lookup took %s", time.Since(started))
}
