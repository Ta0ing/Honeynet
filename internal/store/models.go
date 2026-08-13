package store

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Base struct {
	ID        string         `json:"id" gorm:"type:char(36);primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type User struct {
	Base
	Username          string     `json:"username" gorm:"size:64;uniqueIndex;not null"`
	PasswordHash      string     `json:"-" gorm:"size:255;not null"`
	DisplayName       string     `json:"display_name" gorm:"size:128"`
	Role              string     `json:"role" gorm:"size:32;index;not null"`
	Enabled           bool       `json:"enabled" gorm:"not null;default:true"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	FailedLoginCount  int        `json:"failed_login_count" gorm:"not null;default:0"`
	LastFailedLoginAt *time.Time `json:"last_failed_login_at,omitempty"`
	LockedUntil       *time.Time `json:"locked_until,omitempty" gorm:"index"`
	// TokenVersion is incremented whenever authentication-relevant account
	// state changes. It intentionally never leaves the API, but is embedded in
	// signed JWT claims so logout, password changes, role changes and account
	// disabling revoke already issued tokens without storing bearer tokens.
	TokenVersion uint64 `json:"-" gorm:"not null;default:1"`
}

type AuditLog struct {
	Base
	UserID   string         `json:"user_id" gorm:"type:char(36);index"`
	Username string         `json:"username" gorm:"size:64"`
	Action   string         `json:"action" gorm:"size:16;index"`
	Object   string         `json:"object" gorm:"size:255;index"`
	Detail   datatypes.JSON `json:"detail" gorm:"type:json"`
	IP       string         `json:"ip" gorm:"size:64"`
}

// PlatformOEMSetting is the single, revisioned platform identity document.
// Logos are stored alongside the document so an update is transactional and
// remains consistent when multiple Server instances share the business DB.
// Binary fields are never serialized directly by the management API.
type PlatformOEMSetting struct {
	Base
	SystemName              string `json:"system_name" gorm:"size:128;not null"`
	SystemVersion           string `json:"system_version" gorm:"size:64;not null"`
	Copyright               string `json:"copyright" gorm:"size:512"`
	CustomerServicePhone    string `json:"customer_service_phone" gorm:"size:64"`
	CustomerServiceEmail    string `json:"customer_service_email" gorm:"size:254"`
	OfficialWebsiteURL      string `json:"official_website_url" gorm:"size:1024"`
	ProductDocumentationURL string `json:"product_documentation_url" gorm:"size:1024"`
	SystemLogo              []byte `json:"-" gorm:"type:mediumblob"`
	SystemLogoMIME          string `json:"-" gorm:"size:64"`
	SystemLogoSHA256        string `json:"-" gorm:"size:64"`
	SystemLogoSize          int64  `json:"-" gorm:"not null;default:0"`
	CompanyLogo             []byte `json:"-" gorm:"type:mediumblob"`
	CompanyLogoMIME         string `json:"-" gorm:"size:64"`
	CompanyLogoSHA256       string `json:"-" gorm:"size:64"`
	CompanyLogoSize         int64  `json:"-" gorm:"not null;default:0"`
	Revision                int64  `json:"revision" gorm:"not null;default:1"`
}

type Node struct {
	Base
	Name                           string           `json:"name" gorm:"size:128;not null"`
	GroupName                      string           `json:"group" gorm:"column:group_name;size:128;index"`
	Status                         string           `json:"status" gorm:"size:32;index;not null;default:offline"`
	Version                        string           `json:"version" gorm:"size:64"`
	IP                             string           `json:"ip" gorm:"size:64"`
	AddressMode                    string           `json:"address_mode" gorm:"size:16;not null;default:auto"`
	PublicIP                       string           `json:"public_ip" gorm:"size:64"`
	PublicIPs                      datatypes.JSON   `json:"public_ips" gorm:"type:json"`
	PrivateIPs                     datatypes.JSON   `json:"private_ips" gorm:"type:json"`
	OS                             string           `json:"os" gorm:"size:64"`
	Arch                           string           `json:"arch" gorm:"size:64"`
	Labels                         datatypes.JSON   `json:"labels" gorm:"type:json"`
	Capabilities                   datatypes.JSON   `json:"capabilities" gorm:"type:json"`
	LastHeartbeatAt                *time.Time       `json:"last_heartbeat_at"`
	QueuedEvents                   int              `json:"queued_events" gorm:"not null;default:0"`
	DetectionRuleRevision          int64            `json:"detection_rule_revision" gorm:"not null;default:0;index"`
	DetectionRuleCount             int              `json:"detection_rule_count" gorm:"not null;default:0"`
	DetectionRuleStatus            string           `json:"detection_rule_status" gorm:"size:32;not null;default:pending;index"`
	DetectionRuleSyncedAt          *time.Time       `json:"detection_rule_synced_at,omitempty"`
	DetectionRuleError             string           `json:"detection_rule_error,omitempty" gorm:"size:1024"`
	RegistrationTokenHash          string           `json:"-" gorm:"size:64"`
	AgentTokenHash                 string           `json:"-" gorm:"size:64"`
	TokenExpiresAt                 *time.Time       `json:"token_expires_at,omitempty"`
	CertificateSerial              string           `json:"certificate_serial,omitempty" gorm:"size:64;index"`
	CertificateIssuedAt            *time.Time       `json:"certificate_issued_at,omitempty"`
	CertificateExpiresAt           *time.Time       `json:"certificate_expires_at,omitempty" gorm:"index"`
	PendingCertificateSerial       string           `json:"-" gorm:"size:64;index"`
	PendingCertificateNotAfter     *time.Time       `json:"-"`
	PendingCertificateActivationAt *time.Time       `json:"-" gorm:"index"`
	Sense                          *NodeSenseConfig `json:"sense,omitempty" gorm:"foreignKey:NodeID"`
}

type NodeSenseConfig struct {
	Base
	NodeID          string         `json:"node_id" gorm:"type:char(36);uniqueIndex;not null"`
	Enabled         bool           `json:"enabled" gorm:"not null"`
	Interface       string         `json:"interface" gorm:"size:64"`
	TCPEnabled      bool           `json:"tcp_enabled" gorm:"not null"`
	UDPEnabled      bool           `json:"udp_enabled" gorm:"not null"`
	DistinctPorts   int            `json:"distinct_ports" gorm:"not null"`
	WindowSeconds   int            `json:"window_seconds" gorm:"not null"`
	CooldownSeconds int            `json:"cooldown_seconds" gorm:"not null"`
	ExcludedPorts   datatypes.JSON `json:"excluded_ports" gorm:"column:excluded_ports;type:json;not null"`
	IgnoredCIDRs    datatypes.JSON `json:"ignored_cidrs" gorm:"column:ignored_cidrs;type:json;not null"`
	ActualStatus    string         `json:"actual_status" gorm:"size:32;index;not null"`
	ObservedPackets int64          `json:"observed_packets" gorm:"not null"`
	Detections      int64          `json:"detections" gorm:"not null"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	LastDetectionAt *time.Time     `json:"last_detection_at,omitempty"`
	LastError       string         `json:"last_error,omitempty" gorm:"size:1024"`
}

type PotService struct {
	Code         string         `json:"code" gorm:"size:64;primaryKey"`
	Name         string         `json:"name" gorm:"size:128;not null"`
	Category     string         `json:"category" gorm:"size:64;index"`
	Protocol     string         `json:"protocol" gorm:"size:32"`
	DefaultPort  int            `json:"default_port"`
	Depth        string         `json:"depth" gorm:"size:32"`
	Description  string         `json:"description" gorm:"size:512"`
	ConfigSchema datatypes.JSON `json:"config_schema" gorm:"type:json"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type PotInstance struct {
	Base
	NodeID        string         `json:"node_id" gorm:"type:char(36);index;not null"`
	ServiceCode   string         `json:"service_code" gorm:"size:64;index;not null"`
	TemplateID    *string        `json:"template_id,omitempty" gorm:"type:char(36);index"`
	Name          string         `json:"name" gorm:"size:128;not null"`
	Port          int            `json:"port" gorm:"index;not null"`
	Config        datatypes.JSON `json:"config" gorm:"type:json"`
	DesiredStatus string         `json:"desired_status" gorm:"size:32;index;not null;default:stopped"`
	ActualStatus  string         `json:"actual_status" gorm:"size:32;index;not null;default:pending"`
	Node          *Node          `json:"node,omitempty" gorm:"foreignKey:NodeID"`
	Service       *PotService    `json:"service,omitempty" gorm:"foreignKey:ServiceCode;references:Code"`
	Template      *PotTemplate   `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
}

type PotTemplate struct {
	Base
	Name          string `json:"name" gorm:"size:128;not null;uniqueIndex"`
	YAML          string `json:"yaml" gorm:"type:longtext;not null"`
	Version       int    `json:"version" gorm:"not null;default:1"`
	CreatedBy     string `json:"created_by" gorm:"type:char(36);index"`
	InstanceCount int64  `json:"instance_count" gorm:"-"`
}

type Decoy struct {
	Base
	NodeID       string         `json:"node_id" gorm:"type:char(36);index;not null"`
	Name         string         `json:"name" gorm:"size:128;not null"`
	Type         string         `json:"type" gorm:"size:32;index;not null"`
	Config       datatypes.JSON `json:"config" gorm:"type:json;not null"`
	Status       string         `json:"status" gorm:"size:32;index;not null;default:enabled"`
	ActualStatus string         `json:"actual_status" gorm:"size:32;index;not null;default:pending"`
	ManagedPath  string         `json:"managed_path,omitempty" gorm:"size:4096"`
	LastError    string         `json:"last_error,omitempty" gorm:"size:1024"`
	DeployedAt   *time.Time     `json:"deployed_at,omitempty"`
	LastHitAt    *time.Time     `json:"last_hit_at,omitempty"`
	HitCount     int64          `json:"hit_count" gorm:"not null;default:0"`
	Node         *Node          `json:"node,omitempty" gorm:"foreignKey:NodeID"`
}

type AttackEvent struct {
	EventID            string         `json:"event_id" gorm:"type:char(36);primaryKey"`
	NodeID             string         `json:"node_id" gorm:"type:char(36);index:idx_event_node_time;not null"`
	PotID              string         `json:"pot_id" gorm:"type:char(36);index"`
	DecoyID            string         `json:"decoy_id,omitempty" gorm:"type:char(36);index"`
	Service            string         `json:"service" gorm:"size:64;index:idx_event_service_time"`
	EventType          string         `json:"event_type" gorm:"size:96;index"`
	Timestamp          time.Time      `json:"ts" gorm:"column:ts;index:idx_event_node_time,priority:2;index:idx_event_service_time,priority:2"`
	SrcIP              string         `json:"src_ip" gorm:"size:64;index:idx_event_src_time"`
	SrcPort            int            `json:"src_port"`
	DstIP              string         `json:"dst_ip" gorm:"size:64"`
	DstPort            int            `json:"dst_port"`
	Geo                string         `json:"geo" gorm:"size:128"`
	ASN                string         `json:"asn" gorm:"size:128"`
	RawPacket          string         `json:"raw_packet,omitempty" gorm:"type:mediumtext"`
	Payload            datatypes.JSON `json:"payload" gorm:"type:json"`
	Tags               datatypes.JSON `json:"tags" gorm:"type:json"`
	Detections         datatypes.JSON `json:"detections" gorm:"type:json"`
	AgentRuleRevision  int64          `json:"agent_rule_revision" gorm:"not null;default:0"`
	ServerRuleRevision int64          `json:"server_rule_revision" gorm:"not null;default:0"`
	SessionID          string         `json:"session_id" gorm:"type:char(36);index"`
	CreatedAt          time.Time      `json:"created_at" gorm:"index:idx_event_src_time,priority:2"`
}

func (AttackEvent) TableName() string { return "events" }

// EventReceipt is control-plane metadata only. The complete immutable event is
// stored in ClickHouse; this small MySQL row makes alert/IOC side effects
// idempotent without turning MySQL back into an event warehouse.
type EventReceipt struct {
	EventID     string     `json:"event_id" gorm:"type:char(36);primaryKey"`
	NodeID      string     `json:"node_id" gorm:"type:char(36);index;not null"`
	ReceivedAt  time.Time  `json:"received_at" gorm:"index;not null"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" gorm:"index"`
	LastError   string     `json:"last_error,omitempty" gorm:"size:1024"`
}

// EventMigrationCheckpoint is the durable MySQL control-plane cursor for the
// one-time legacy events -> ClickHouse migration. It contains no security
// event payload. Completed checkpoints remain as upgrade audit metadata.
type EventMigrationCheckpoint struct {
	Name           string     `json:"name" gorm:"size:128;primaryKey"`
	LastCreatedAt  *time.Time `json:"last_created_at,omitempty" gorm:"index"`
	LastEventID    string     `json:"last_event_id" gorm:"type:char(36)"`
	MigratedEvents int64      `json:"migrated_events" gorm:"not null;default:0"`
	Status         string     `json:"status" gorm:"size:32;index;not null;default:pending"`
	LastError      string     `json:"last_error,omitempty" gorm:"size:1024"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Alert struct {
	Base
	EventID     string     `json:"event_id" gorm:"type:char(36);index"`
	RuleID      string     `json:"rule_id" gorm:"type:char(36);index"`
	Fingerprint string     `json:"fingerprint" gorm:"size:64;uniqueIndex"`
	SilenceKey  string     `json:"silence_key,omitempty" gorm:"size:64;index"`
	Title       string     `json:"title" gorm:"size:255;not null"`
	Level       string     `json:"level" gorm:"size:32;index;not null"`
	Status      string     `json:"status" gorm:"size:32;index;not null;default:new"`
	SourceIP    string     `json:"source_ip" gorm:"size:64;index"`
	NodeID      string     `json:"node_id" gorm:"type:char(36);index"`
	Service     string     `json:"service" gorm:"size:64"`
	Description string     `json:"description" gorm:"size:1024"`
	AckedBy     string     `json:"acked_by" gorm:"type:char(36)"`
	AckedAt     *time.Time `json:"acked_at"`
}

type AlertRule struct {
	Base
	Name          string         `json:"name" gorm:"size:128;not null"`
	Enabled       bool           `json:"enabled" gorm:"default:true"`
	EventType     string         `json:"event_type" gorm:"size:96;index"`
	Level         string         `json:"level" gorm:"size:32"`
	Threshold     int            `json:"threshold"`
	WindowMinute  int            `json:"window_minute"`
	SilenceMinute int            `json:"silence_minute"`
	ChannelIDs    datatypes.JSON `json:"channel_ids" gorm:"type:json"`
}

// DetectionRule matches the content of a single event. AlertRule remains the
// independent aggregation/threshold rule type.
type DetectionRule struct {
	Base
	RuleKey           string         `json:"key" gorm:"size:128;uniqueIndex;not null"`
	Name              string         `json:"name" gorm:"size:255;not null"`
	Description       string         `json:"description" gorm:"type:text"`
	Severity          string         `json:"severity" gorm:"size:32;index;not null"`
	Source            string         `json:"source" gorm:"size:64;index"`
	ExternalID        string         `json:"external_id" gorm:"size:128;index"`
	Enabled           bool           `json:"enabled" gorm:"not null;index"`
	AgentEnabled      bool           `json:"agent_enabled" gorm:"not null"`
	ServerEnabled     bool           `json:"server_enabled" gorm:"not null"`
	Revision          int64          `json:"revision" gorm:"not null"`
	Patterns          datatypes.JSON `json:"patterns" gorm:"type:json;not null"`
	OriginalCondition string         `json:"original_condition" gorm:"type:text"`
	ValidationError   string         `json:"validation_error,omitempty" gorm:"size:1024"`
}

type AIAnalysis struct {
	Base
	TargetType string         `json:"target_type" gorm:"size:32;index:idx_ai_target;not null"`
	TargetID   string         `json:"target_id" gorm:"size:128;index:idx_ai_target;not null"`
	Kind       string         `json:"kind" gorm:"size:32;index;not null"`
	Status     string         `json:"status" gorm:"size:32;index;not null"`
	Provider   string         `json:"provider" gorm:"size:64"`
	Model      string         `json:"model" gorm:"size:128"`
	Summary    string         `json:"summary" gorm:"type:mediumtext"`
	Result     datatypes.JSON `json:"result" gorm:"type:json"`
	Error      string         `json:"error,omitempty" gorm:"size:1024"`
	PromptHash string         `json:"prompt_hash" gorm:"size:64"`
}

// AIHarnessRun is the durable, provider-neutral execution record for one AI
// Agent objective. Evidence is referenced by immutable identifiers/digests;
// raw event payloads are deliberately not persisted in the control plane.
type AIHarnessRun struct {
	Base
	Goal           string         `json:"goal" gorm:"type:text;not null"`
	Kind           string         `json:"kind" gorm:"size:32;index;not null"`
	Status         string         `json:"status" gorm:"size:32;index;not null"`
	Stage          string         `json:"stage" gorm:"size:32;index;not null"`
	CreatedBy      string         `json:"created_by" gorm:"type:char(36);index;not null"`
	TargetRuleID   string         `json:"target_rule_id,omitempty" gorm:"type:char(36);index"`
	EvidenceDigest string         `json:"evidence_digest" gorm:"size:64;index"`
	Evidence       datatypes.JSON `json:"evidence" gorm:"type:json;not null"`
	Trace          datatypes.JSON `json:"trace" gorm:"type:json;not null"`
	Result         datatypes.JSON `json:"result" gorm:"type:json;not null"`
	Error          string         `json:"error,omitempty" gorm:"size:1024"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
}

// DetectionRuleProposal separates model output from executable detection
// rules. A proposal can only reach the rule runtime after deterministic
// evaluation and an explicit administrator approval/publish transition.
type DetectionRuleProposal struct {
	Base
	RunID             string         `json:"run_id" gorm:"type:char(36);index;not null"`
	RuleID            string         `json:"rule_id,omitempty" gorm:"type:char(36);index"`
	Action            string         `json:"action" gorm:"size:16;index;not null"`
	Status            string         `json:"status" gorm:"size:32;index;not null"`
	Title             string         `json:"title" gorm:"size:255;not null"`
	Rationale         string         `json:"rationale" gorm:"type:mediumtext"`
	Candidate         datatypes.JSON `json:"candidate" gorm:"type:json;not null"`
	Baseline          datatypes.JSON `json:"baseline" gorm:"type:json;not null"`
	Evidence          datatypes.JSON `json:"evidence" gorm:"type:json;not null"`
	Evaluation        datatypes.JSON `json:"evaluation" gorm:"type:json;not null"`
	CreatedBy         string         `json:"created_by" gorm:"type:char(36);index;not null"`
	ReviewedBy        string         `json:"reviewed_by,omitempty" gorm:"type:char(36);index"`
	ReviewComment     string         `json:"review_comment,omitempty" gorm:"size:1024"`
	ReviewedAt        *time.Time     `json:"reviewed_at,omitempty"`
	PublishedRuleID   string         `json:"published_rule_id,omitempty" gorm:"type:char(36);index"`
	PublishedRevision int64          `json:"published_revision_text,string,omitempty" gorm:"not null;default:0"`
	PublishedAt       *time.Time     `json:"published_at,omitempty"`
	RolledBackBy      string         `json:"rolled_back_by,omitempty" gorm:"type:char(36);index"`
	RollbackReason    string         `json:"rollback_reason,omitempty" gorm:"size:1024"`
	RolledBackAt      *time.Time     `json:"rolled_back_at,omitempty"`
}

// DetectionRuleFeedback is the human-labelled signal used by later Harness
// runs. It is intentionally append-only through the public API.
type DetectionRuleFeedback struct {
	Base
	ProposalID string `json:"proposal_id,omitempty" gorm:"type:char(36);index"`
	RuleID     string `json:"rule_id,omitempty" gorm:"type:char(36);index"`
	EventID    string `json:"event_id" gorm:"size:128;index;not null"`
	Verdict    string `json:"verdict" gorm:"size:32;index;not null"`
	Comment    string `json:"comment,omitempty" gorm:"size:1024"`
	CreatedBy  string `json:"created_by" gorm:"type:char(36);index;not null"`
}

// AISetting is the single Server-side AI provider configuration. Provider
// credentials are encrypted before persistence and are never serialized by
// the management API.
type AISetting struct {
	Base
	Enabled          bool   `json:"enabled" gorm:"not null"`
	Provider         string `json:"provider" gorm:"size:64;not null"`
	BaseURL          string `json:"base_url" gorm:"size:1024"`
	Model            string `json:"model" gorm:"size:128"`
	TimeoutSeconds   int    `json:"timeout_seconds" gorm:"not null"`
	SendRawPacket    bool   `json:"send_raw_packet" gorm:"not null"`
	APIKeyCiphertext []byte `json:"-" gorm:"type:mediumblob"`
	Revision         int64  `json:"revision" gorm:"not null"`
}

type AlertChannel struct {
	Base
	Name    string         `json:"name" gorm:"size:128;not null"`
	Type    string         `json:"type" gorm:"size:32;index"`
	Enabled bool           `json:"enabled" gorm:"default:true"`
	Config  datatypes.JSON `json:"-" gorm:"type:json"`
}

type AlertDelivery struct {
	Base
	AlertID     string     `json:"alert_id" gorm:"type:char(36);uniqueIndex:idx_alert_channel;index;not null"`
	ChannelID   string     `json:"channel_id" gorm:"type:char(36);uniqueIndex:idx_alert_channel;index;not null"`
	ChannelName string     `json:"channel_name" gorm:"size:128"`
	ChannelType string     `json:"channel_type" gorm:"size:32;index"`
	Status      string     `json:"status" gorm:"size:32;index;not null;default:pending"`
	Attempt     int        `json:"attempt" gorm:"not null;default:0"`
	LastError   string     `json:"last_error,omitempty" gorm:"size:1024"`
	NextAttempt time.Time  `json:"next_attempt" gorm:"index"`
	LastAttempt *time.Time `json:"last_attempt,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
}

type AgentRelease struct {
	Base
	Version string `json:"version" gorm:"size:64;uniqueIndex;not null"`
	Notes   string `json:"notes" gorm:"size:1024"`
	Status  string `json:"status" gorm:"size:32;index;not null;default:active"`
	KeyID   string `json:"key_id" gorm:"size:64;not null"`
}

type AgentBuild struct {
	Base
	ReleaseID string `json:"release_id" gorm:"type:char(36);uniqueIndex:idx_release_platform;index;not null"`
	OS        string `json:"os" gorm:"size:32;uniqueIndex:idx_release_platform;not null"`
	Arch      string `json:"arch" gorm:"size:32;uniqueIndex:idx_release_platform;not null"`
	Filename  string `json:"filename" gorm:"size:255;not null"`
	SHA256    string `json:"sha256" gorm:"size:64;not null"`
	Signature string `json:"signature" gorm:"size:128;not null"`
	Size      int64  `json:"size" gorm:"not null"`
}

type UpgradeRollout struct {
	Base
	Name         string     `json:"name" gorm:"size:128;not null"`
	ReleaseID    string     `json:"release_id" gorm:"type:char(36);index;not null"`
	Version      string     `json:"version" gorm:"size:64;index;not null"`
	Strategy     string     `json:"strategy" gorm:"size:32;not null"`
	CanaryCount  int        `json:"canary_count"`
	BatchSize    int        `json:"batch_size"`
	PauseSeconds int        `json:"pause_seconds"`
	CurrentWave  int        `json:"current_wave" gorm:"not null;default:0"`
	Status       string     `json:"status" gorm:"size:32;index;not null;default:running"`
	CreatedBy    string     `json:"created_by" gorm:"type:char(36);index"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

type UpgradeTask struct {
	Base
	RolloutID        string     `json:"rollout_id" gorm:"type:char(36);uniqueIndex:idx_rollout_node;index;not null"`
	NodeID           string     `json:"node_id" gorm:"type:char(36);uniqueIndex:idx_rollout_node;index;not null"`
	BuildID          string     `json:"build_id" gorm:"type:char(36);index;not null"`
	Wave             int        `json:"wave" gorm:"index;not null"`
	FromVersion      string     `json:"from_version" gorm:"size:64"`
	TargetVersion    string     `json:"target_version" gorm:"size:64;index;not null"`
	Status           string     `json:"status" gorm:"size:32;index;not null;default:pending"`
	Attempt          int        `json:"attempt" gorm:"not null;default:0"`
	LastError        string     `json:"last_error,omitempty" gorm:"size:1024"`
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	ConfirmedVersion string     `json:"confirmed_version" gorm:"size:64"`
}

type IOC struct {
	Base
	Type       string    `json:"type" gorm:"size:32;index:idx_ioc_type_value,unique"`
	Value      string    `json:"value" gorm:"size:512;index:idx_ioc_type_value,unique"`
	Source     string    `json:"source" gorm:"size:128"`
	Confidence int       `json:"confidence"`
	EventID    string    `json:"event_id" gorm:"type:char(36);index"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
}

type IntelFeed struct {
	Base
	Name       string         `json:"name" gorm:"size:128;not null"`
	Type       string         `json:"type" gorm:"size:32;index;not null"`
	URL        string         `json:"url" gorm:"size:1024;not null"`
	Schedule   string         `json:"schedule" gorm:"size:64"`
	Auth       datatypes.JSON `json:"-" gorm:"type:json"`
	Enabled    bool           `json:"enabled" gorm:"not null;default:true"`
	LastSyncAt *time.Time     `json:"last_sync_at"`
}
