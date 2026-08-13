package httpapi

import "testing"

func TestValidateTemplate(t *testing.T) {
	valid := "name: fake-login\nlisten:\n  port: 8080\npages:\n  - path: /login\n    method: GET\n"
	if err := validateTemplate(valid); err != nil {
		t.Fatalf("validateTemplate(valid) error = %v", err)
	}
	invalid := []string{
		"name: missing-pages\nlisten:\n  port: 80\n",
		"name: bad-port\nlisten:\n  port: 70000\npages:\n  - path: /\n",
		"name: bad-path\nlisten:\n  port: 80\npages:\n  - path: relative\n",
	}
	for _, value := range invalid {
		if validateTemplate(value) == nil {
			t.Fatalf("validateTemplate(%q) accepted invalid template", value)
		}
	}
}

func TestValidateFeed(t *testing.T) {
	if !validateFeed("https://intel.example.test/feed.csv", "csv") {
		t.Fatal("validateFeed rejected HTTPS CSV feed")
	}
	if validateFeed("file:///etc/passwd", "csv") {
		t.Fatal("validateFeed accepted local file URL")
	}
	if validateFeed("https://intel.example.test/feed", "unknown") {
		t.Fatal("validateFeed accepted unknown feed type")
	}
}

func TestNormalizeDecoyConfig(t *testing.T) {
	config, err := normalizeDecoyConfig("file", []byte(`{"path":"/var/lib/honeynet/decoys/report.txt","create_parent":true}`))
	if err != nil || len(config) == 0 {
		t.Fatalf("normalizeDecoyConfig(valid) = %s, %v", config, err)
	}
	invalid := []byte(`{"path":"/etc/passwd","command":"cat /etc/shadow"}`)
	if _, err := normalizeDecoyConfig("file", invalid); err == nil {
		t.Fatal("normalizeDecoyConfig accepted an unknown executable field")
	}
}
