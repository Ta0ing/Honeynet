package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/alerting"
	"github.com/honeynet/honeynet/internal/analytics"
	"github.com/honeynet/honeynet/internal/detection"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ingestAnalyticsStub struct {
	analytics.Store
	events      map[string]analytics.Event
	insertCalls int
	count       uint64
	insertErr   error
}

func newIngestAnalyticsStub() *ingestAnalyticsStub {
	return &ingestAnalyticsStub{events: make(map[string]analytics.Event)}
}

func (s *ingestAnalyticsStub) Get(_ context.Context, eventID string) (analytics.Event, error) {
	event, ok := s.events[eventID]
	if !ok {
		return analytics.Event{}, analytics.ErrNotFound
	}
	return event, nil
}

func (s *ingestAnalyticsStub) InsertEvent(_ context.Context, event analytics.Event) error {
	s.insertCalls++
	if s.insertErr != nil {
		return s.insertErr
	}
	if _, exists := s.events[event.EventID]; !exists {
		s.events[event.EventID] = event
	}
	return nil
}

func TestIngestAnalyticsFailureRejectsOnlyFailedEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	analyticsStub := newIngestAnalyticsStub()
	api, db := newIngestTestAPI(t, analyticsStub)
	nodeID := uuid.NewString()
	firstID, secondID := uuid.NewString(), uuid.NewString()
	analyticsStub.insertErr = errors.New("clickhouse unavailable")
	_, response := ingestBatchRequest(t, api, nodeID, []map[string]any{
		eventPayload(firstID, time.Now().Unix()), eventPayload(secondID, time.Now().Add(time.Second).Unix()),
	})
	data := responseData(t, response)
	if len(data["ack_ids"].([]any)) != 0 || len(data["reject"].([]any)) != 2 {
		t.Fatalf("failed analytics result=%#v", data)
	}
	for _, eventID := range []string{firstID, secondID} {
		var receipt store.EventReceipt
		if err := db.First(&receipt, "event_id = ?", eventID).Error; err != nil || receipt.ProcessedAt != nil || receipt.LastError == "" {
			t.Fatalf("pending receipt %s=%#v err=%v", eventID, receipt, err)
		}
	}
	analyticsStub.insertErr = nil
	_, response = ingestBatchRequest(t, api, nodeID, []map[string]any{
		eventPayload(firstID, time.Now().Add(time.Hour).Unix()), eventPayload(secondID, time.Now().Add(time.Hour+time.Second).Unix()),
	})
	if len(responseData(t, response)["ack_ids"].([]any)) != 2 {
		t.Fatalf("analytics retry did not ACK both: %#v", response)
	}
}

func (s *ingestAnalyticsStub) InsertEvents(ctx context.Context, events []analytics.Event) error {
	for _, event := range events {
		if err := s.InsertEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (s *ingestAnalyticsStub) CountAlertWindow(context.Context, analytics.AlertWindowFilter) (uint64, error) {
	return s.count, nil
}

func newIngestTestAPI(t *testing.T, eventStore analytics.Store) (*API, *gorm.DB) {
	t.Helper()
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared&_busy_timeout=5000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&store.AttackEvent{}, &store.EventReceipt{}, &store.Decoy{}, &store.Alert{},
		&store.AlertRule{}, &store.AlertChannel{}, &store.AlertDelivery{}, &store.IOC{},
	); err != nil {
		t.Fatal(err)
	}
	matcher, err := detection.Compile(nil)
	if err != nil {
		t.Fatal(err)
	}
	api := &API{
		db: db, hub: NewHub(), alerts: alerting.NewDispatcher(db, ""),
		detection: &detectionRuntime{matcher: matcher}, analytics: eventStore,
	}
	return api, db
}

func ingestRequest(t *testing.T, api *API, nodeID string, event map[string]any) (int, map[string]any) {
	return ingestBatchRequest(t, api, nodeID, []map[string]any{event})
}

func ingestBatchRequest(t *testing.T, api *API, nodeID string, events []map[string]any) (int, map[string]any) {
	t.Helper()
	body, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/agent/v1/events", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("agent.node", store.Node{Base: store.Base{ID: nodeID}})
	api.ingestEvents(c)
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return recorder.Code, response
}

func eventPayload(eventID string, ts int64) map[string]any {
	return map[string]any{
		"event_id": eventID, "event_type": "web.request", "service": "http", "ts": ts,
		"src":     map[string]any{"ip": "203.0.113.8", "port": 45678},
		"dst":     map[string]any{"ip": "192.0.2.10", "port": 8080},
		"payload": map[string]any{"path": "/"},
	}
}

func responseData(t *testing.T, response map[string]any) map[string]any {
	t.Helper()
	data, ok := response["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing response data: %#v", response)
	}
	return data
}

func TestIngestRejectsInvalidEventUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newIngestTestAPI(t, nil)
	status, response := ingestRequest(t, api, uuid.NewString(), eventPayload("not-a-uuid", time.Now().Unix()))
	if status != http.StatusOK {
		t.Fatalf("status=%d response=%#v", status, response)
	}
	data := responseData(t, response)
	if len(data["ack_ids"].([]any)) != 0 || len(data["reject"].([]any)) != 1 {
		t.Fatalf("unexpected result: %#v", data)
	}
	assertRejectFatal(t, data, true)
	var receipts int64
	if err := db.Model(&store.EventReceipt{}).Count(&receipts).Error; err != nil || receipts != 0 {
		t.Fatalf("receipts=%d err=%v", receipts, err)
	}
}

func TestIngestRejectsDuplicateEventIDWithinBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _ := newIngestTestAPI(t, nil)
	eventID := uuid.NewString()
	first := eventPayload(eventID, time.Now().Unix())
	second := eventPayload(eventID, time.Now().Add(time.Second).Unix())
	_, response := ingestBatchRequest(t, api, uuid.NewString(), []map[string]any{first, second})
	data := responseData(t, response)
	if len(data["ack_ids"].([]any)) != 1 || len(data["reject"].([]any)) != 1 {
		t.Fatalf("duplicate batch result=%#v", data)
	}
	assertRejectFatal(t, data, false)
}

func TestIngestFatalValidationRejects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _ := newIngestTestAPI(t, nil)
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "missing required field", mutate: func(event map[string]any) { delete(event, "event_type") }},
		{name: "payload null", mutate: func(event map[string]any) { event["payload"] = nil }},
		{name: "payload array", mutate: func(event map[string]any) { event["payload"] = []any{"not", "object"} }},
		{name: "tags object", mutate: func(event map[string]any) { event["tags"] = map[string]any{"not": "an array"} }},
		{name: "tags mixed array", mutate: func(event map[string]any) { event["tags"] = []any{"valid", 3} }},
		// Printable bytes keep the encoded request below the independent 1 MiB
		// batch limit, so these cases exercise the intended per-event boundary.
		{name: "raw packet field too large", mutate: func(event map[string]any) { event["raw_packet"] = strings.Repeat("x", (256<<10)+1) }},
		{name: "raw packet fallback too large", mutate: func(event map[string]any) {
			event["payload"] = map[string]any{"raw_request": strings.Repeat("x", (256<<10)+1)}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := eventPayload(uuid.NewString(), time.Now().Unix())
			test.mutate(event)
			_, response := ingestRequest(t, api, uuid.NewString(), event)
			assertRejectFatal(t, responseData(t, response), true)
		})
	}
}

func TestIngestLegacySideEffectFailureRetriesWithoutDuplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newIngestTestAPI(t, nil)
	nodeID, eventID, decoyID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	event := eventPayload(eventID, time.Now().Unix())
	event["event_type"] = "decoy.file"
	event["decoy_id"] = decoyID

	// Force a transactional MySQL side-effect failure. A missing/removed decoy
	// is intentionally non-fatal so late Agent reports cannot retry forever.
	if err := db.Exec("DROP TABLE iocs").Error; err != nil {
		t.Fatal(err)
	}
	_, first := ingestRequest(t, api, nodeID, event)
	firstData := responseData(t, first)
	if len(firstData["ack_ids"].([]any)) != 0 || len(firstData["reject"].([]any)) != 1 {
		t.Fatalf("first attempt must reject: %#v", firstData)
	}
	var receipt store.EventReceipt
	if err := db.First(&receipt, "event_id = ?", eventID).Error; err != nil {
		t.Fatal(err)
	}
	if receipt.ProcessedAt != nil || receipt.LastError == "" {
		t.Fatalf("failed receipt must remain pending with error: %#v", receipt)
	}
	if err := db.AutoMigrate(&store.IOC{}); err != nil {
		t.Fatal(err)
	}
	decoy := store.Decoy{Base: store.Base{ID: decoyID}, NodeID: nodeID, Name: "document", Type: "file", Config: datatypes.JSON(`{"path":"/tmp/a","mode":"0644"}`), Status: "enabled", ActualStatus: "running"}
	if err := db.Create(&decoy).Error; err != nil {
		t.Fatal(err)
	}
	_, second := ingestRequest(t, api, nodeID, event)
	if len(responseData(t, second)["ack_ids"].([]any)) != 1 {
		t.Fatalf("retry must ACK: %#v", second)
	}
	_, third := ingestRequest(t, api, nodeID, event)
	if len(responseData(t, third)["ack_ids"].([]any)) != 1 {
		t.Fatalf("processed retry must ACK: %#v", third)
	}
	if err := db.First(&decoy, "id = ?", decoyID).Error; err != nil {
		t.Fatal(err)
	}
	if decoy.HitCount != 1 {
		t.Fatalf("hit_count=%d, want 1", decoy.HitCount)
	}
	if err := db.First(&receipt, "event_id = ?", eventID).Error; err != nil || receipt.ProcessedAt == nil || receipt.LastError != "" {
		t.Fatalf("processed receipt=%#v err=%v", receipt, err)
	}
}

func TestIngestMissingDecoyDoesNotRetryForever(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newIngestTestAPI(t, nil)
	nodeID, eventID := uuid.NewString(), uuid.NewString()
	event := eventPayload(eventID, time.Now().Unix())
	event["event_type"] = "decoy.file"
	event["decoy_id"] = uuid.NewString()
	_, response := ingestRequest(t, api, nodeID, event)
	if len(responseData(t, response)["ack_ids"].([]any)) != 1 {
		t.Fatalf("removed decoy event must ACK: %#v", response)
	}
	var receipt store.EventReceipt
	if err := db.First(&receipt, "event_id = ?", eventID).Error; err != nil || receipt.ProcessedAt == nil {
		t.Fatalf("receipt=%#v err=%v", receipt, err)
	}
}

func TestDetectionAlertAndDeliveryRemainIdempotentOnPendingRetry(t *testing.T) {
	api, db := newIngestTestAPI(t, nil)
	channel := store.AlertChannel{Base: store.NewBase(), Name: "webhook", Type: "webhook", Enabled: true, Config: datatypes.JSON(`{}`)}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatal(err)
	}
	eventID, nodeID := uuid.NewString(), uuid.NewString()
	hit := detection.Hit{RuleID: uuid.NewString(), RuleKey: "stable-rule", Name: "stable alert", Severity: "high", Stage: "server"}
	hits, _ := json.Marshal([]detection.Hit{hit})
	event := store.AttackEvent{EventID: eventID, NodeID: nodeID, EventType: "web.request", Service: "http", Timestamp: time.Now(), Payload: datatypes.JSON(`{}`), Tags: datatypes.JSON(`[]`), Detections: hits, CreatedAt: time.Now()}
	receipt := store.EventReceipt{EventID: eventID, NodeID: nodeID, ReceivedAt: time.Now()}
	if err := db.Create(&receipt).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := api.processEventBusinessEffects(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	// Simulate recovery metadata lag while the alert/delivery rows already
	// exist. Stable alert IDs and the delivery composite unique key must make
	// the resumed transaction a no-op for those rows.
	if err := db.Model(&store.EventReceipt{}).Where("event_id = ?", eventID).Updates(map[string]any{"processed_at": nil}).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := api.processEventBusinessEffects(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	var alerts, deliveries int64
	if err := db.Model(&store.Alert{}).Where("event_id = ?", eventID).Count(&alerts).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&store.AlertDelivery{}).Count(&deliveries).Error; err != nil {
		t.Fatal(err)
	}
	if alerts != 1 || deliveries != 1 {
		t.Fatalf("alerts=%d deliveries=%d, want 1/1", alerts, deliveries)
	}
}

func TestMalformedNetworkDecoyRejectsAndKeepsReceiptPending(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newIngestTestAPI(t, nil)
	nodeID, eventID := uuid.NewString(), uuid.NewString()
	decoy := store.Decoy{Base: store.NewBase(), NodeID: nodeID, Name: "broken", Type: "network", Config: datatypes.JSON(`{"token":"short"}`), Status: "enabled", ActualStatus: "running"}
	if err := db.Create(&decoy).Error; err != nil {
		t.Fatal(err)
	}
	_, response := ingestRequest(t, api, nodeID, eventPayload(eventID, time.Now().Unix()))
	data := responseData(t, response)
	if len(data["ack_ids"].([]any)) != 0 || len(data["reject"].([]any)) != 1 {
		t.Fatalf("malformed network decoy must reject: %#v", data)
	}
	var receipt store.EventReceipt
	if err := db.First(&receipt, "event_id = ?", eventID).Error; err != nil || receipt.ProcessedAt != nil || receipt.LastError == "" {
		t.Fatalf("receipt=%#v err=%v", receipt, err)
	}
}

func TestIngestRejectsCrossNodeEventID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := newIngestTestAPI(t, nil)
	eventID := uuid.NewString()
	firstNode, secondNode := uuid.NewString(), uuid.NewString()
	_, first := ingestRequest(t, api, firstNode, eventPayload(eventID, time.Now().Unix()))
	if len(responseData(t, first)["ack_ids"].([]any)) != 1 {
		t.Fatalf("first ingest failed: %#v", first)
	}
	_, second := ingestRequest(t, api, secondNode, eventPayload(eventID, time.Now().Add(time.Hour).Unix()))
	data := responseData(t, second)
	if len(data["ack_ids"].([]any)) != 0 || len(data["reject"].([]any)) != 1 {
		t.Fatalf("cross-node event must reject: %#v", data)
	}
	assertRejectFatal(t, data, true)
	var event store.AttackEvent
	if err := db.First(&event, "event_id = ?", eventID).Error; err != nil || event.NodeID != firstNode {
		t.Fatalf("canonical event owner changed: %#v err=%v", event, err)
	}
}

func TestIngestAnalyticsRejectsCrossNodeBeforeRewrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	analyticsStub := newIngestAnalyticsStub()
	api, _ := newIngestTestAPI(t, analyticsStub)
	eventID := uuid.NewString()
	firstNode, secondNode := uuid.NewString(), uuid.NewString()
	_, first := ingestRequest(t, api, firstNode, eventPayload(eventID, time.Now().Unix()))
	if len(responseData(t, first)["ack_ids"].([]any)) != 1 || analyticsStub.insertCalls != 1 {
		t.Fatalf("first analytics ingest=%#v insert_calls=%d", first, analyticsStub.insertCalls)
	}
	_, second := ingestRequest(t, api, secondNode, eventPayload(eventID, time.Now().Add(time.Hour).Unix()))
	data := responseData(t, second)
	if len(data["ack_ids"].([]any)) != 0 || len(data["reject"].([]any)) != 1 {
		t.Fatalf("cross-node analytics event must reject: %#v", data)
	}
	assertRejectFatal(t, data, true)
	if analyticsStub.insertCalls != 1 || analyticsStub.events[eventID].NodeID != firstNode {
		t.Fatalf("cross-node retry rewrote analytics: calls=%d event=%#v", analyticsStub.insertCalls, analyticsStub.events[eventID])
	}
}

func assertRejectFatal(t *testing.T, data map[string]any, expected bool) {
	t.Helper()
	rejects, ok := data["reject"].([]any)
	if !ok || len(rejects) != 1 {
		t.Fatalf("rejects=%#v", data["reject"])
	}
	reject, ok := rejects[0].(map[string]any)
	if !ok {
		t.Fatalf("reject=%#v", rejects[0])
	}
	fatal, present := reject["fatal"].(bool)
	if expected && (!present || !fatal) {
		t.Fatalf("fatal reject missing: %#v", reject)
	}
	if !expected && present {
		t.Fatalf("retryable reject exposed fatal=%v: %#v", fatal, reject)
	}
}

func TestIngestCanonicalRetryUsesStoredDetections(t *testing.T) {
	gin.SetMode(gin.TestMode)
	analyticsStub := newIngestAnalyticsStub()
	api, db := newIngestTestAPI(t, analyticsStub)
	nodeID, eventID, missingDecoyID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	hit := detection.Hit{RuleID: uuid.NewString(), RuleKey: "canonical-rule", Name: "canonical", Severity: "high", Stage: "server", Revision: 7}
	hits, _ := json.Marshal([]detection.Hit{hit})
	canonical := store.AttackEvent{
		EventID: eventID, NodeID: nodeID, DecoyID: missingDecoyID, Service: "http", EventType: "decoy.file",
		Timestamp: time.Now(), SrcIP: "203.0.113.9", DstPort: 8080, Payload: datatypes.JSON(`{"path":"/original"}`),
		Tags: datatypes.JSON(`[]`), Detections: hits, ServerRuleRevision: 7, CreatedAt: time.Now().Add(-time.Minute),
	}
	analyticsStub.events[eventID] = analytics.FromAttackEvent(canonical)
	incoming := eventPayload(eventID, time.Now().Add(time.Hour).Unix())
	incoming["raw_packet"] = "changed-by-retry"
	if err := db.Exec("DROP TABLE iocs").Error; err != nil {
		t.Fatal(err)
	}
	_, response := ingestRequest(t, api, nodeID, incoming)
	if len(responseData(t, response)["ack_ids"].([]any)) != 0 {
		t.Fatalf("side-effect failure must reject: %#v", response)
	}
	if err := db.AutoMigrate(&store.IOC{}); err != nil {
		t.Fatal(err)
	}
	decoy := store.Decoy{Base: store.Base{ID: missingDecoyID}, NodeID: nodeID, Name: "canonical", Type: "file", Config: datatypes.JSON(`{"path":"/tmp/c","mode":"0644"}`), Status: "enabled", ActualStatus: "running"}
	if err := db.Create(&decoy).Error; err != nil {
		t.Fatal(err)
	}
	_, response = ingestRequest(t, api, nodeID, incoming)
	if len(responseData(t, response)["ack_ids"].([]any)) != 1 {
		t.Fatalf("canonical retry must ACK: %#v", response)
	}
	var alert store.Alert
	if err := db.First(&alert, "event_id = ?", eventID).Error; err != nil {
		t.Fatal(err)
	}
	if alert.Title != "canonical" || alert.RuleID != hit.RuleID {
		t.Fatalf("alert was not based on canonical detection: %#v", alert)
	}
	if analyticsStub.insertCalls != 0 {
		t.Fatalf("canonical retry rewrote ClickHouse %d times", analyticsStub.insertCalls)
	}
}

func TestThresholdAlertSilenceKeyAndRetryIdempotency(t *testing.T) {
	api, db := newIngestTestAPI(t, nil)
	rule := store.AlertRule{Base: store.NewBase(), Name: "burst", Enabled: true, EventType: "web.*", Level: "high", Threshold: 1, WindowMinute: 5, SilenceMinute: 30, ChannelIDs: datatypes.JSON(`[]`)}
	if err := db.Create(&rule).Error; err != nil {
		t.Fatal(err)
	}
	base := time.Now()
	first := store.AttackEvent{EventID: uuid.NewString(), NodeID: uuid.NewString(), EventType: "web.request", Service: "http", SrcIP: "203.0.113.40", Timestamp: base, CreatedAt: base, Payload: datatypes.JSON(`{}`), Tags: datatypes.JSON(`[]`), Detections: datatypes.JSON(`[]`)}
	second := first
	second.EventID = uuid.NewString()
	second.Timestamp = base.Add(time.Minute)
	second.CreatedAt = second.Timestamp
	if err := db.Create(&first).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}
	alerts, err := api.evaluateAlertRules(context.Background(), db, first)
	if err != nil || len(alerts) != 1 {
		t.Fatalf("first alerts=%d err=%v", len(alerts), err)
	}
	if alerts[0].SilenceKey == "" || alerts[0].Fingerprint == alerts[0].SilenceKey {
		t.Fatalf("invalid alert keys: %#v", alerts[0])
	}
	alerts, err = api.evaluateAlertRules(context.Background(), db, first)
	if err != nil || len(alerts) != 0 {
		t.Fatalf("retry duplicated alert: %d err=%v", len(alerts), err)
	}
	alerts, err = api.evaluateAlertRules(context.Background(), db, second)
	if err != nil || len(alerts) != 0 {
		t.Fatalf("silence window did not suppress second event: %d err=%v", len(alerts), err)
	}
	var count int64
	if err := db.Model(&store.Alert{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("alert count=%d err=%v", count, err)
	}
}

func TestProcessEffectsRequiresOwnedPendingReceipt(t *testing.T) {
	api, _ := newIngestTestAPI(t, nil)
	event := store.AttackEvent{EventID: uuid.NewString(), NodeID: uuid.NewString(), EventType: "web.request", Timestamp: time.Now(), Payload: datatypes.JSON(`{}`), Tags: datatypes.JSON(`[]`), Detections: datatypes.JSON(`[]`)}
	_, _, err := api.processEventBusinessEffects(context.Background(), event)
	if err == nil || err.Error() != "event receipt already processed or missing" {
		t.Fatalf("missing receipt must fail, got %v", err)
	}
}
