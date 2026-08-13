// Package analytics defines the security-event boundary between Honeynet's
// MySQL control plane and its analytical event store.
package analytics

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrNotFound = errors.New("analytics event not found")

// Event is an immutable security observation. Mutable workflow state such as
// alert acknowledgement deliberately does not belong here and remains in
// MySQL.
type Event struct {
	EventID    string          `json:"event_id"`
	NodeID     string          `json:"node_id"`
	PotID      string          `json:"pot_id"`
	DecoyID    string          `json:"decoy_id,omitempty"`
	Service    string          `json:"service"`
	EventType  string          `json:"event_type"`
	EventTime  time.Time       `json:"ts"`
	IngestedAt time.Time       `json:"created_at"`
	SourceIP   string          `json:"src_ip"`
	SourcePort uint16          `json:"src_port"`
	TargetIP   string          `json:"dst_ip"`
	TargetPort uint16          `json:"dst_port"`
	Geo        string          `json:"geo"`
	ASN        string          `json:"asn"`
	RawPacket  string          `json:"raw_packet,omitempty"`
	Payload    json.RawMessage `json:"payload"`
	Tags       json.RawMessage `json:"tags"`
	// Detections preserves the complete, original Agent/Server detection hit
	// JSON. The rule revisions make the result auditable after rules change.
	Detections         json.RawMessage `json:"detections"`
	AgentRuleRevision  int64           `json:"agent_rule_revision"`
	ServerRuleRevision int64           `json:"server_rule_revision"`
	SessionID          string          `json:"session_id"`
	Credential         *Credential     `json:"credential,omitempty"`
	RecordVersion      uint64          `json:"record_version,omitempty"`
}

type Credential struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	AuthResponse string `json:"auth_response,omitempty"`
	Mechanism    string `json:"mechanism,omitempty"`
}

func (e Event) Validate() error {
	if strings.TrimSpace(e.EventID) == "" {
		return errors.New("event_id is required")
	}
	if strings.TrimSpace(e.NodeID) == "" {
		return errors.New("node_id is required")
	}
	if strings.TrimSpace(e.EventType) == "" {
		return errors.New("event_type is required")
	}
	if e.EventTime.IsZero() {
		return errors.New("event time is required")
	}
	if len(e.Payload) > 0 && !json.Valid(e.Payload) {
		return errors.New("payload must be valid JSON")
	}
	if len(e.Tags) > 0 && !json.Valid(e.Tags) {
		return errors.New("tags must be valid JSON")
	}
	if len(e.Detections) > 0 && !json.Valid(e.Detections) {
		return errors.New("detections must be valid JSON")
	}
	return nil
}

func NormalizeEvent(e Event, now time.Time) (Event, error) {
	if err := e.Validate(); err != nil {
		return Event{}, err
	}
	e.EventID = strings.TrimSpace(e.EventID)
	e.NodeID = strings.TrimSpace(e.NodeID)
	e.PotID = strings.TrimSpace(e.PotID)
	e.DecoyID = strings.TrimSpace(e.DecoyID)
	e.Service = strings.TrimSpace(e.Service)
	e.EventType = strings.TrimSpace(e.EventType)
	e.SourceIP = strings.TrimSpace(e.SourceIP)
	e.TargetIP = strings.TrimSpace(e.TargetIP)
	e.Geo = strings.TrimSpace(e.Geo)
	e.ASN = strings.TrimSpace(e.ASN)
	e.SessionID = strings.TrimSpace(e.SessionID)
	e.EventTime = e.EventTime.UTC()
	if e.IngestedAt.IsZero() {
		e.IngestedAt = now.UTC()
	} else {
		e.IngestedAt = e.IngestedAt.UTC()
	}
	if len(e.Payload) == 0 {
		e.Payload = json.RawMessage("{}")
	}
	if len(e.Tags) == 0 {
		e.Tags = json.RawMessage("[]")
	}
	if len(e.Detections) == 0 {
		e.Detections = json.RawMessage("[]")
	}
	if e.RecordVersion == 0 {
		// A retry of the same immutable event must have the same version even
		// when a Server reconstructs IngestedAt. Explicit non-zero versions are
		// reserved for intentional corrections/replacements.
		e.RecordVersion = StableRecordVersion(e.EventID)
	}
	return e, nil
}

func StableRecordVersion(eventID string) uint64 {
	digest := sha256.Sum256([]byte(strings.TrimSpace(eventID)))
	version := binary.BigEndian.Uint64(digest[:8])
	if version == 0 {
		return 1
	}
	return version
}

type Cursor struct {
	EventTime time.Time
	EventID   string
}

const (
	MaxCursorLength      = 512
	cursorVersion        = byte(2)
	cursorChecksumLength = 12
	maxCursorEventIDSize = 256
)

var ErrInvalidCursor = errors.New("invalid analytics cursor")

// EncodeCursor returns a compact URL-safe, versioned token. The checksum is
// deliberately included to reject truncated or accidentally modified cursors;
// cursor values are opaque pagination state, not authorization credentials.
func EncodeCursor(cursor Cursor) (string, error) {
	eventID := strings.TrimSpace(cursor.EventID)
	if cursor.EventTime.IsZero() || eventID == "" || eventID != cursor.EventID || len(eventID) > maxCursorEventIDSize || !utf8.ValidString(eventID) || containsCursorControl(eventID) {
		return "", ErrInvalidCursor
	}
	payload := make([]byte, 1+8+4+2+len(eventID))
	payload[0] = cursorVersion
	eventTime := cursor.EventTime.UTC()
	binary.BigEndian.PutUint64(payload[1:9], uint64(eventTime.Unix()))
	binary.BigEndian.PutUint32(payload[9:13], uint32(eventTime.Nanosecond()))
	binary.BigEndian.PutUint16(payload[13:15], uint16(len(eventID)))
	copy(payload[15:], eventID)
	digest := sha256.Sum256(payload)
	token := base64.RawURLEncoding.EncodeToString(append(payload, digest[:cursorChecksumLength]...))
	if len(token) > MaxCursorLength {
		return "", ErrInvalidCursor
	}
	return token, nil
}

func DecodeCursor(token string) (Cursor, error) {
	if token == "" || token != strings.TrimSpace(token) || len(token) > MaxCursorLength {
		return Cursor{}, ErrInvalidCursor
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(decoded) < 1+8+4+2+1+cursorChecksumLength {
		return Cursor{}, ErrInvalidCursor
	}
	payload := decoded[:len(decoded)-cursorChecksumLength]
	checksum := decoded[len(decoded)-cursorChecksumLength:]
	digest := sha256.Sum256(payload)
	if subtle.ConstantTimeCompare(checksum, digest[:cursorChecksumLength]) != 1 || payload[0] != cursorVersion {
		return Cursor{}, ErrInvalidCursor
	}
	nanoseconds := binary.BigEndian.Uint32(payload[9:13])
	if nanoseconds >= uint32(time.Second) {
		return Cursor{}, ErrInvalidCursor
	}
	eventIDLength := int(binary.BigEndian.Uint16(payload[13:15]))
	if eventIDLength < 1 || eventIDLength > maxCursorEventIDSize || len(payload) != 15+eventIDLength {
		return Cursor{}, ErrInvalidCursor
	}
	eventID := string(payload[15:])
	if strings.TrimSpace(eventID) != eventID || !utf8.ValidString(eventID) || containsCursorControl(eventID) {
		return Cursor{}, ErrInvalidCursor
	}
	eventTime := time.Unix(int64(binary.BigEndian.Uint64(payload[1:9])), int64(nanoseconds)).UTC()
	if eventTime.IsZero() || eventTime.Year() < 1970 || eventTime.Year() > 9999 {
		return Cursor{}, ErrInvalidCursor
	}
	return Cursor{EventTime: eventTime, EventID: eventID}, nil
}

func containsCursorControl(value string) bool {
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

type EventFilter struct {
	NodeID        string
	PotID         string
	DecoyID       string
	Service       string
	EventType     string
	EventClass    string
	SourceIP      string
	ExactSourceIP string
	From          time.Time
	To            time.Time
	Cursor        *Cursor
	CursorMode    bool
	// SkipTotal lets interactive page navigation reuse the exact total from
	// the first page instead of repeating an expensive count on every page.
	// Page offsets remain bounded by the HTTP API.
	SkipTotal bool
	Limit     int
	Offset    int
}

type EventPage struct {
	Items      []Event `json:"items"`
	Total      uint64  `json:"total"`
	TotalKnown bool    `json:"total_known"`
	NextCursor *Cursor `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

type DayCount struct {
	Day   string `json:"day"`
	Count uint64 `json:"count"`
}

type AttackerCount struct {
	SourceIP string    `json:"src_ip"`
	Geo      string    `json:"geo"`
	Count    uint64    `json:"count"`
	LastSeen time.Time `json:"last_seen"`
}

type DashboardStats struct {
	Events uint64 `json:"events"`
}

type CredentialResource struct {
	EventID      string    `json:"event_id"`
	NodeID       string    `json:"node_id"`
	EventType    string    `json:"event_type"`
	EventTime    time.Time `json:"ts"`
	SourceIP     string    `json:"src_ip"`
	Geo          string    `json:"geo"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	AuthResponse string    `json:"auth_response"`
	Mechanism    string    `json:"mechanism"`
	Service      string    `json:"service"`
}

type CredentialCount struct {
	Value string `json:"value"`
	Count uint64 `json:"count"`
}

type CredentialFilter struct {
	Keyword          string
	Service          string
	From             time.Time
	To               time.Time
	Limit            int
	Offset           int
	IncludeSensitive bool
}

type CredentialPage struct {
	Items        []CredentialResource `json:"items"`
	Total        uint64               `json:"total"`
	TopUsernames []CredentialCount    `json:"top_usernames"`
	TopPasswords []CredentialCount    `json:"top_passwords"`
}

type NamedCount struct {
	Name  string `json:"name"`
	Count uint64 `json:"count"`
}

type AttackerEvidence struct {
	SourceIP   string       `json:"source_ip"`
	Geo        string       `json:"geo"`
	ASN        string       `json:"asn"`
	Count      uint64       `json:"count"`
	FirstSeen  time.Time    `json:"first_seen"`
	LastSeen   time.Time    `json:"last_seen"`
	Services   []NamedCount `json:"services"`
	EventTypes []NamedCount `json:"event_types"`
	Events     []Event      `json:"events"`
}

type AlertWindowFilter struct {
	From             time.Time
	To               time.Time
	EventTypePattern string
	Service          string
	SourceIP         string
	NodeID           string
}

type Writer interface {
	InsertEvent(context.Context, Event) error
	InsertEvents(context.Context, []Event) error
	AppendBatch(context.Context, []Event) error
}

type Query interface {
	Get(context.Context, string) (Event, error)
	List(context.Context, EventFilter) (EventPage, error)
	Dashboard(context.Context, time.Time, time.Time) (DashboardStats, error)
	Trends(context.Context, time.Time, time.Time) ([]DayCount, error)
	TopAttackers(context.Context, time.Time, time.Time, int) ([]AttackerCount, error)
	CountByNodes(context.Context, []string) (map[string]uint64, error)
	ListCredentials(context.Context, CredentialFilter) (CredentialPage, error)
	AttackerEvidence(context.Context, string, time.Time, time.Time, int) (AttackerEvidence, error)
	CountAlertWindow(context.Context, AlertWindowFilter) (uint64, error)
}

type Store interface {
	Writer
	Query
	Ping(context.Context) error
	Status(context.Context) Status
	Close() error
}

type Status struct {
	Enabled       bool       `json:"enabled"`
	Healthy       bool       `json:"healthy"`
	Driver        string     `json:"driver"`
	Database      string     `json:"database,omitempty"`
	Table         string     `json:"table,omitempty"`
	SchemaVersion int        `json:"schema_version,omitempty"`
	ServerVersion string     `json:"server_version,omitempty"`
	LastWriteAt   *time.Time `json:"last_write_at,omitempty"`
	Error         string     `json:"error,omitempty"`
}

func ValidIP(value string) bool { return net.ParseIP(strings.TrimSpace(value)) != nil }

func ClampLimit(value, fallback, maximum int) int {
	if value <= 0 {
		return fallback
	}
	if value > maximum {
		return maximum
	}
	return value
}

func ValidateRange(from, to time.Time) error {
	if !from.IsZero() && !to.IsZero() && from.After(to) {
		return fmt.Errorf("from must not be after to")
	}
	return nil
}
