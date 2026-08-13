package protocol

import (
	"time"

	"github.com/honeynet/honeynet/internal/detection"
)

type Endpoint struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type Event struct {
	EventID    string          `json:"event_id"`
	PotID      string          `json:"pot_id"`
	DecoyID    string          `json:"decoy_id,omitempty"`
	Service    string          `json:"service"`
	EventType  string          `json:"event_type"`
	TS         int64           `json:"ts"`
	Src        Endpoint        `json:"src"`
	Dst        Endpoint        `json:"dst"`
	RawPacket  string          `json:"raw_packet,omitempty"`
	Payload    map[string]any  `json:"payload"`
	Tags       []string        `json:"tags"`
	Detections []detection.Hit `json:"detections,omitempty"`
	// RuleRevision records the Agent ruleset used even when no rule matched.
	RuleRevision int64 `json:"rule_revision,omitempty"`
}

type PotTarget struct {
	ID            string         `json:"id"`
	Service       string         `json:"service"`
	Name          string         `json:"name"`
	Port          int            `json:"port"`
	Config        map[string]any `json:"config"`
	DesiredStatus string         `json:"desired_status"`
	Template      *WebTemplate   `json:"template,omitempty"`
}

type WebTemplate struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	YAML    string `json:"yaml"`
}

type DecoyTarget struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
	Status string         `json:"status"`
}

type DecoyResult struct {
	DecoyID     string     `json:"decoy_id"`
	Status      string     `json:"status"`
	Success     bool       `json:"success"`
	ManagedPath string     `json:"managed_path,omitempty"`
	LastError   string     `json:"error,omitempty"`
	HitCount    int64      `json:"hit_count"`
	LastHitAt   *time.Time `json:"last_hit_at,omitempty"`
}

type DecoyStatus struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	ManagedPath string     `json:"managed_path,omitempty"`
	LastError   string     `json:"last_error,omitempty"`
	HitCount    int64      `json:"hit_count"`
	LastHitAt   *time.Time `json:"last_hit_at,omitempty"`
}

type PotResult struct {
	PotID   string `json:"pot_id"`
	Status  string `json:"status"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type SenseConfig struct {
	Enabled         bool     `json:"enabled"`
	Interface       string   `json:"interface"`
	TCPEnabled      bool     `json:"tcp_enabled"`
	UDPEnabled      bool     `json:"udp_enabled"`
	DistinctPorts   int      `json:"distinct_ports"`
	WindowSeconds   int      `json:"window_seconds"`
	CooldownSeconds int      `json:"cooldown_seconds"`
	ExcludedPorts   []int    `json:"excluded_ports"`
	IgnoredCIDRs    []string `json:"ignored_cidrs"`
}

type SenseStatus struct {
	Enabled         bool       `json:"enabled"`
	ActualStatus    string     `json:"actual_status"`
	Interface       string     `json:"interface"`
	ObservedPackets int64      `json:"observed_packets"`
	Detections      int64      `json:"detections"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	LastDetectionAt *time.Time `json:"last_detection_at,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
}

func NewEvent(eventType string, src, dst Endpoint, payload map[string]any, tags ...string) Event {
	if payload == nil {
		payload = map[string]any{}
	}
	rawPacket, _ := payload["raw_request"].(string)
	return Event{EventType: eventType, TS: time.Now().Unix(), Src: src, Dst: dst, RawPacket: rawPacket, Payload: payload, Tags: tags}
}
