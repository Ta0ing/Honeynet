package detection

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestMatcherAllPatternsAndMinCount(t *testing.T) {
	rules := []Rule{{ID: "1", Key: "test", Name: "SQL injection", Severity: "high", Patterns: []Pattern{
		{ID: "method", Field: "method", Operator: PatternContains, Value: "GET", MinCount: 1},
		{ID: "payload", Field: "raw", Operator: PatternContains, Value: "../", MinCount: 2},
	}}}
	matcher, err := Compile(rules)
	if err != nil {
		t.Fatal(err)
	}
	hits := matcher.Match(Event{RawPacket: "GET /?a=../../ HTTP/1.1\r\n\r\n", Payload: map[string]any{"method": "GET"}}, "agent")
	if len(hits) != 1 || hits[0].Severity != "high" {
		t.Fatalf("unexpected hits: %#v", hits)
	}
}

func TestImportBuiltinRules(t *testing.T) {
	_, current, _, _ := runtime.Caller(0)
	directory := filepath.Join(filepath.Dir(current), "..", "..", "cve-rules-decrypted", "Yara")
	rules, err := ImportRuleDirectory(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 54 {
		t.Fatalf("expected 54 rule blocks, got %d", len(rules))
	}
	valid := 0
	for _, item := range rules {
		if item.ValidationError == "" {
			valid++
		}
	}
	if valid != 52 {
		t.Fatalf("expected 52 portable rules, got %d", valid)
	}
}
