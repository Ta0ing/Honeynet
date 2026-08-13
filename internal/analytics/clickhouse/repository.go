package clickhouse

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/honeynet/honeynet/internal/analytics"
	"github.com/honeynet/honeynet/internal/store"
)

type Repository struct {
	conn          driver.Conn
	table         string
	database      string
	schemaVersion int
	now           func() time.Time
	mu            sync.RWMutex
	lastWriteAt   time.Time
}

func (r *Repository) tableSQL() string {
	value, err := qualifiedTable(r.database, r.table)
	if err != nil {
		// Identifiers are validated by constructors; this branch is defensive.
		return "`default`.`security_events`"
	}
	return value
}

const eventColumns = `event_id,node_id,pot_id,decoy_id,service,event_type,event_time,ingested_at,src_ip,src_port,dst_ip,dst_port,geo,asn,raw_packet,payload,tags,detections,agent_rule_revision,server_rule_revision,session_id,has_credential,credential_username,credential_password,credential_auth_response,credential_mechanism,record_version`

func (r *Repository) InsertEvent(ctx context.Context, event analytics.Event) error {
	return r.InsertEvents(ctx, []analytics.Event{event})
}

func (r *Repository) InsertAttackEvent(ctx context.Context, event store.AttackEvent) error {
	return r.InsertEvent(ctx, analytics.FromAttackEvent(event))
}

func (r *Repository) InsertAttackEvents(ctx context.Context, events []store.AttackEvent) error {
	converted := make([]analytics.Event, 0, len(events))
	for _, event := range events {
		converted = append(converted, analytics.FromAttackEvent(event))
	}
	return r.InsertEvents(ctx, converted)
}

func (r *Repository) AppendBatch(ctx context.Context, events []analytics.Event) error {
	return r.InsertEvents(ctx, events)
}

func (r *Repository) InsertEvents(ctx context.Context, events []analytics.Event) error {
	if len(events) == 0 {
		return nil
	}
	normalized := make([]analytics.Event, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	for _, input := range events {
		event, err := analytics.NormalizeEvent(input, r.now())
		if err != nil {
			return fmt.Errorf("normalize event %q: %w", input.EventID, err)
		}
		// Duplicate IDs within one request are deterministically collapsed.
		// Retrying the exact same batch receives the same ClickHouse token.
		if _, duplicate := seen[event.EventID]; duplicate {
			continue
		}
		seen[event.EventID] = struct{}{}
		event.Credential = analytics.ExtractCredential(event)
		normalized = append(normalized, event)
	}
	if len(normalized) == 0 {
		return nil
	}
	token := deduplicationToken(normalized)
	insertContext := ch.Context(ctx, ch.WithSettings(ch.Settings{
		"insert_deduplication_token": token,
		"insert_deduplicate":         1,
	}))
	batch, err := r.conn.PrepareBatch(insertContext, "INSERT INTO "+r.tableSQL()+" ("+eventColumns+")")
	if err != nil {
		return fmt.Errorf("prepare clickhouse event batch: %w", err)
	}
	defer batch.Close()
	for _, event := range normalized {
		credential := analytics.Credential{}
		hasCredential := uint8(0)
		if event.Credential != nil {
			credential = *event.Credential
			if credential.Username != "" || credential.Password != "" || credential.AuthResponse != "" {
				hasCredential = 1
			}
		}
		if err := batch.Append(
			event.EventID, event.NodeID, event.PotID, event.DecoyID, event.Service, event.EventType,
			event.EventTime, event.IngestedAt, event.SourceIP, event.SourcePort, event.TargetIP, event.TargetPort,
			event.Geo, event.ASN, event.RawPacket, string(event.Payload), string(event.Tags), string(event.Detections),
			event.AgentRuleRevision, event.ServerRuleRevision, event.SessionID, hasCredential,
			credential.Username, credential.Password, credential.AuthResponse, credential.Mechanism, event.RecordVersion,
		); err != nil {
			return fmt.Errorf("append clickhouse event %q: %w", event.EventID, err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("send clickhouse event batch: %w", err)
	}
	r.mu.Lock()
	r.lastWriteAt = r.now().UTC()
	r.mu.Unlock()
	return nil
}

func deduplicationToken(events []analytics.Event) string {
	hash := sha256.New()
	for _, event := range events {
		_, _ = hash.Write([]byte(event.EventID))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(fmt.Sprintf("%d", event.RecordVersion)))
		_, _ = hash.Write([]byte{0})
	}
	return "honeynet-events-" + hex.EncodeToString(hash.Sum(nil))
}

func (r *Repository) Ping(ctx context.Context) error { return r.conn.Ping(ctx) }

func (r *Repository) Status(ctx context.Context) analytics.Status {
	status := analytics.Status{Enabled: true, Driver: "clickhouse", Database: r.database, Table: r.table}
	r.mu.RLock()
	status.SchemaVersion = r.schemaVersion
	if !r.lastWriteAt.IsZero() {
		lastWrite := r.lastWriteAt
		status.LastWriteAt = &lastWrite
	}
	r.mu.RUnlock()
	if version, err := r.conn.ServerVersion(); err == nil && version != nil {
		status.ServerVersion = version.String()
	}
	if err := r.Ping(ctx); err != nil {
		status.Error = "安全分析引擎连接不可用"
		return status
	}
	if err := r.ValidateSchema(ctx); err != nil {
		status.SchemaVersion = r.currentSchemaVersion()
		status.Error = "安全分析引擎数据结构校验失败"
		return status
	}
	status.SchemaVersion = r.currentSchemaVersion()
	status.Healthy = true
	return status
}

func (r *Repository) Close() error { return r.conn.Close() }

func (r *Repository) Table() string { return r.table }

func (r *Repository) Database() string { return r.database }

func scanEvent(scanner interface{ Scan(...any) error }) (analytics.Event, error) {
	var event analytics.Event
	var payload, tags, detections string
	var hasCredential uint8
	var username, password, response, mechanism string
	err := scanner.Scan(
		&event.EventID, &event.NodeID, &event.PotID, &event.DecoyID, &event.Service, &event.EventType,
		&event.EventTime, &event.IngestedAt, &event.SourceIP, &event.SourcePort, &event.TargetIP, &event.TargetPort,
		&event.Geo, &event.ASN, &event.RawPacket, &payload, &tags, &detections,
		&event.AgentRuleRevision, &event.ServerRuleRevision, &event.SessionID, &hasCredential,
		&username, &password, &response, &mechanism, &event.RecordVersion,
	)
	if err != nil {
		return analytics.Event{}, err
	}
	event.Payload = json.RawMessage(payload)
	event.Tags = json.RawMessage(tags)
	event.Detections = json.RawMessage(detections)
	if hasCredential != 0 {
		event.Credential = &analytics.Credential{Username: username, Password: password, AuthResponse: response, Mechanism: mechanism}
	}
	event.EventTime = event.EventTime.UTC()
	event.IngestedAt = event.IngestedAt.UTC()
	return event, nil
}

func (r *Repository) Get(ctx context.Context, eventID string) (analytics.Event, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return analytics.Event{}, analytics.ErrNotFound
	}
	query := "SELECT " + eventColumns + " FROM " + r.tableSQL() + " FINAL WHERE event_id = ? ORDER BY record_version DESC LIMIT 1"
	event, err := scanEvent(r.conn.QueryRow(ctx, query, eventID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return analytics.Event{}, analytics.ErrNotFound
		}
		return analytics.Event{}, fmt.Errorf("get clickhouse event: %w", err)
	}
	return event, nil
}

func (r *Repository) GetAttackEvent(ctx context.Context, eventID string) (store.AttackEvent, error) {
	event, err := r.Get(ctx, eventID)
	if err != nil {
		return store.AttackEvent{}, err
	}
	return analytics.ToAttackEvent(event), nil
}

func (r *Repository) ListAttackEvents(ctx context.Context, filter analytics.EventFilter) ([]store.AttackEvent, uint64, *analytics.Cursor, error) {
	page, err := r.List(ctx, filter)
	if err != nil {
		return nil, 0, nil, err
	}
	items := make([]store.AttackEvent, 0, len(page.Items))
	for _, event := range page.Items {
		items = append(items, analytics.ToAttackEvent(event))
	}
	return items, page.Total, page.NextCursor, nil
}
