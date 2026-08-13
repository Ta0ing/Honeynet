package analytics

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

func TestNormalizeEvent(t *testing.T) {
	now := time.Date(2026, 8, 12, 8, 0, 0, 123000000, time.FixedZone("CST", 8*3600))
	event, err := NormalizeEvent(Event{EventID: " event-1 ", NodeID: " node-1 ", EventType: "web.request", EventTime: now}, now)
	if err != nil {
		t.Fatal(err)
	}
	if event.EventID != "event-1" || event.NodeID != "node-1" {
		t.Fatalf("identifiers were not trimmed: %#v", event)
	}
	if event.EventTime.Location() != time.UTC || event.IngestedAt.Location() != time.UTC {
		t.Fatal("times were not normalized to UTC")
	}
	if string(event.Payload) != "{}" || string(event.Tags) != "[]" || string(event.Detections) != "[]" {
		t.Fatalf("unexpected JSON defaults: %s %s %s", event.Payload, event.Tags, event.Detections)
	}
	if event.RecordVersion == 0 {
		t.Fatal("record version was not populated")
	}
	retry, err := NormalizeEvent(Event{EventID: "event-1", NodeID: "node-1", EventType: "web.request", EventTime: now}, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if event.RecordVersion != retry.RecordVersion {
		t.Fatalf("retry version changed: %d != %d", event.RecordVersion, retry.RecordVersion)
	}
}

func TestExtractCredentialCleansPayload(t *testing.T) {
	event := Event{EventType: "web.credential", Service: "http", SourceIP: "192.0.2.10", Payload: json.RawMessage(`{"username":" admin ","password":"<script>alert(1)</script>","auth_response":"abcd"}`)}
	credential := ExtractCredential(event)
	if credential == nil {
		t.Fatal("credential was not extracted")
	}
	if credential.Username != "admin" || credential.Password != "" || credential.AuthResponse != "abcd" {
		t.Fatalf("unexpected credential: %#v", credential)
	}
}

func TestAttackEventAdapterRoundTrip(t *testing.T) {
	timestamp := time.Date(2026, 8, 12, 1, 2, 3, 0, time.UTC)
	original := store.AttackEvent{EventID: "event-1", NodeID: "node-1", Service: "ssh", EventType: "ssh.credential", Timestamp: timestamp, SrcIP: "192.0.2.1", SrcPort: 55555, DstIP: "198.51.100.1", DstPort: 22, Payload: datatypes.JSON(`{"username":"root"}`), Tags: datatypes.JSON(`[]`), Detections: datatypes.JSON(`[{"rule_key":"r1"}]`), AgentRuleRevision: 4, ServerRuleRevision: 7, CreatedAt: timestamp}
	converted := ToAttackEvent(FromAttackEvent(original))
	if converted.EventID != original.EventID || converted.SrcPort != original.SrcPort || converted.DstPort != original.DstPort || string(converted.Detections) != string(original.Detections) || converted.AgentRuleRevision != 4 || converted.ServerRuleRevision != 7 {
		t.Fatalf("round trip mismatch: %#v", converted)
	}
}

func TestValidateRangeAndClamp(t *testing.T) {
	now := time.Now()
	if ValidateRange(now, now.Add(-time.Second)) == nil {
		t.Fatal("invalid range accepted")
	}
	if ClampLimit(0, 100, 1000) != 100 || ClampLimit(2000, 100, 1000) != 1000 {
		t.Fatal("unexpected limit clamp")
	}
}

func TestCursorRoundTripIsOpaqueAndMillisecondStable(t *testing.T) {
	want := Cursor{
		EventTime: time.Date(2026, 8, 12, 3, 4, 5, 678901234, time.FixedZone("CST", 8*60*60)),
		EventID:   "event-01K2G7S9AQ",
	}
	token, err := EncodeCursor(want)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(token, want.EventID) || strings.ContainsAny(token, "+/=") {
		t.Fatalf("cursor is not opaque URL-safe data: %q", token)
	}
	got, err := DecodeCursor(token)
	if err != nil {
		t.Fatal(err)
	}
	if got.EventID != want.EventID || !got.EventTime.Equal(want.EventTime.UTC()) {
		t.Fatalf("cursor round trip = %#v, want %#v", got, want)
	}
}

func TestDecodeCursorRejectsMalformedOversizedAndModifiedTokens(t *testing.T) {
	token, err := EncodeCursor(Cursor{EventTime: time.Now().UTC(), EventID: "event-1"})
	if err != nil {
		t.Fatal(err)
	}
	modified := token[:len(token)-1] + map[bool]string{true: "A", false: "B"}[token[len(token)-1] != 'A']
	for _, input := range []string{"", "not+base64", token + " ", modified, strings.Repeat("A", MaxCursorLength+1)} {
		if _, err := DecodeCursor(input); !errors.Is(err, ErrInvalidCursor) {
			t.Fatalf("DecodeCursor(%q) err=%v, want ErrInvalidCursor", input, err)
		}
	}
}
