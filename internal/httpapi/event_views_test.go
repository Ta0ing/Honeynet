package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/analytics"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPresentAttackEventTargetAddressPriority(t *testing.T) {
	tests := []struct {
		name string
		node eventNodeView
		want string
	}{
		{name: "public address", node: eventNodeView{Name: "杭州节点", IP: "192.168.16.2", PublicIP: "47.96.80.162"}, want: "47.96.80.162"},
		{name: "selected address", node: eventNodeView{Name: "内网节点", IP: "10.3.0.6"}, want: "10.3.0.6"},
		{name: "observed address", node: eventNodeView{}, want: "172.18.0.4"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := presentAttackEvent(store.AttackEvent{NodeID: "node-1", DstIP: "172.18.0.4"}, test.node)
			if view.DisplayDstIP != test.want {
				t.Fatalf("display_dst_ip = %q, want %q", view.DisplayDstIP, test.want)
			}
			if view.DstIP != "172.18.0.4" || view.ObservedDstIP != "172.18.0.4" {
				t.Fatalf("forensic destination was changed: dst=%q observed=%q", view.DstIP, view.ObservedDstIP)
			}
		})
	}
}

func TestPresentAttackEventAddsNodeContext(t *testing.T) {
	view := presentAttackEvent(store.AttackEvent{EventID: "event-1", NodeID: "node-1", DstIP: "10.0.0.8"}, eventNodeView{ID: "node-1", Name: "生产区探针", IP: "10.0.0.8", PublicIP: "203.0.113.8"})
	if view.NodeName != "生产区探针" || view.NodeAddress != "10.0.0.8" || view.NodePublicIP != "203.0.113.8" {
		t.Fatalf("unexpected node context: %#v", view)
	}
	raw, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if json.Unmarshal(raw, &document) != nil || document["event_id"] != "event-1" || document["dst_ip"] != "10.0.0.8" || document["node_name"] != "生产区探针" || document["display_dst_ip"] != "203.0.113.8" {
		t.Fatalf("event view JSON did not retain original and display fields: %s", raw)
	}
}

func TestRedactAttackEventDeeplyMasksCredentialsAndSummarizesRawEvidence(t *testing.T) {
	payload := datatypes.JSON(`{
		"Password":"top-secret",
		"nested":{"AUTH_response":"deadbeef","items":[{"client_token":"bearer-value"}]},
		"headers":{"Authorization":"Basic YWRtaW46cGFzcw==","Cookie":"sid=secret"},
		"raw_request":"POST /login HTTP/1.1\r\n\r\npassword=top-secret",
		"body_base64":"cGFzc3dvcmQ9dG9wLXNlY3JldA==",
		"method":"POST","path":"/login"
	}`)
	original := store.AttackEvent{RawPacket: "raw packet top-secret", Payload: append(datatypes.JSON(nil), payload...)}
	redacted := redactAttackEvent(original)

	encoded, err := json.Marshal(redacted)
	if err != nil {
		t.Fatal(err)
	}
	response := string(encoded)
	for _, secret := range []string{"top-secret", "deadbeef", "bearer-value", "YWRtaW46cGFzcw", "sid=secret", "cGFzc3dvcmQ9"} {
		if strings.Contains(response, secret) {
			t.Fatalf("redacted event leaked %q: %s", secret, response)
		}
	}
	if !strings.Contains(response, maskedEventCredentialValue) || !strings.Contains(response, `"byte_length"`) || !strings.Contains(redacted.RawPacket, "SHA-256") {
		t.Fatalf("response did not retain redaction evidence metadata: %s", response)
	}
	if original.RawPacket != "raw packet top-secret" || string(original.Payload) != string(payload) {
		t.Fatalf("redaction mutated persisted source: raw=%q payload=%s", original.RawPacket, original.Payload)
	}
}

type eventViewAnalyticsStub struct {
	analytics.Store
	event analytics.Event
}

type cursorEventAnalyticsStub struct {
	analytics.Store
	event      analytics.Event
	lastFilter analytics.EventFilter
	calls      int
}

func (stub *cursorEventAnalyticsStub) List(_ context.Context, filter analytics.EventFilter) (analytics.EventPage, error) {
	stub.lastFilter = filter
	stub.calls++
	return analytics.EventPage{
		Items:      []analytics.Event{stub.event},
		NextCursor: &analytics.Cursor{EventTime: stub.event.EventTime, EventID: stub.event.EventID},
		HasMore:    true,
	}, nil
}

func (stub eventViewAnalyticsStub) Get(context.Context, string) (analytics.Event, error) {
	return stub.event, nil
}

func (stub eventViewAnalyticsStub) List(context.Context, analytics.EventFilter) (analytics.EventPage, error) {
	return analytics.EventPage{Items: []analytics.Event{stub.event}, Total: 1}, nil
}

func newEventViewTestAPI(t *testing.T, useAnalytics bool) (*API, *gorm.DB, store.AttackEvent) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&store.Node{}, &store.AttackEvent{}, &store.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	node := store.Node{Base: store.NewBase(), Name: "取证节点", IP: "10.0.0.8", PublicIP: "203.0.113.8", Status: "online"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}
	event := store.AttackEvent{
		EventID: uuid.NewString(), NodeID: node.ID, Service: "http", EventType: "web.credential",
		Timestamp: time.Now().UTC(), SrcIP: "198.51.100.9", SrcPort: 54321, DstIP: "10.0.0.8", DstPort: 8080,
		RawPacket: "POST /login HTTP/1.1\r\nAuthorization: Bearer raw-token\r\n\r\npassword=server-secret",
		Payload:   datatypes.JSON(`{"username":"admin","password":"server-secret","body":"password=server-secret","nested":{"Cookie":"session=private"}}`),
		Tags:      datatypes.JSON(`[]`), Detections: datatypes.JSON(`[]`), CreatedAt: time.Now().UTC(),
	}
	api := &API{db: db}
	if useAnalytics {
		api.analytics = eventViewAnalyticsStub{event: analytics.FromAttackEvent(event)}
	} else if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	return api, db, event
}

func callEventViewHandler(t *testing.T, api *API, role, path, query string, list bool) (int, map[string]any) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, path+query, nil)
	c.Set(userContextKey, AuthUser{ID: uuid.NewString(), Username: role, Role: role})
	c.Set("request_id", uuid.NewString())
	if list {
		api.listEvents(c)
	} else {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		c.Params = gin.Params{{Key: "id", Value: parts[len(parts)-1]}}
		api.getEvent(c)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return recorder.Code, response
}

func eventResponseDocument(t *testing.T, response map[string]any, list bool) map[string]any {
	t.Helper()
	data, ok := response["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing response data: %#v", response)
	}
	if !list {
		return data
	}
	items, ok := data["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("missing event list item: %#v", data)
	}
	item, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("invalid event list item: %#v", items[0])
	}
	return item
}

func assertEventDocumentRedacted(t *testing.T, document map[string]any) {
	t.Helper()
	encoded, _ := json.Marshal(document)
	if strings.Contains(string(encoded), "server-secret") || strings.Contains(string(encoded), "raw-token") || strings.Contains(string(encoded), "session=private") {
		t.Fatalf("event API leaked sensitive evidence: %s", encoded)
	}
	if document["sensitive_visible"] != false || document["evidence_redacted"] != true || document["sensitive_redaction_enabled"] != true || document["sensitive_reveal_audited"] != false {
		t.Fatalf("redaction metadata is ambiguous: %#v", document)
	}
}

func assertEventDocumentRawDefault(t *testing.T, document map[string]any) {
	t.Helper()
	encoded, _ := json.Marshal(document)
	if !strings.Contains(string(encoded), "server-secret") || !strings.Contains(string(encoded), "raw-token") || !strings.Contains(string(encoded), "session=private") {
		t.Fatalf("default raw event response lost original evidence: %s", encoded)
	}
	if document["sensitive_visible"] != true || document["evidence_redacted"] != false || document["sensitive_redaction_enabled"] != false || document["sensitive_reveal_audited"] != false {
		t.Fatalf("default raw disclosure metadata is ambiguous: %#v", document)
	}
}

func TestEventHandlersDefaultToRawViewsForMySQLAndClickHouse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, useAnalytics := range []bool{false, true} {
		name := "mysql"
		if useAnalytics {
			name = "clickhouse"
		}
		t.Run(name, func(t *testing.T) {
			api, _, event := newEventViewTestAPI(t, useAnalytics)
			status, response := callEventViewHandler(t, api, "admin", "/api/v1/events", "", true)
			if status != http.StatusOK {
				t.Fatalf("list status=%d response=%#v", status, response)
			}
			assertEventDocumentRawDefault(t, eventResponseDocument(t, response, true))

			status, response = callEventViewHandler(t, api, "admin", "/api/v1/events/"+event.EventID, "", false)
			if status != http.StatusOK {
				t.Fatalf("get status=%d response=%#v", status, response)
			}
			assertEventDocumentRawDefault(t, eventResponseDocument(t, response, false))
		})
	}
}

func TestEventHandlersRedactionPolicyMasksMySQLAndClickHouse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, useAnalytics := range []bool{false, true} {
		name := "mysql"
		if useAnalytics {
			name = "clickhouse"
		}
		t.Run(name, func(t *testing.T) {
			api, _, event := newEventViewTestAPI(t, useAnalytics)
			api.cfg.RedactSensitiveEvents = true
			status, response := callEventViewHandler(t, api, "admin", "/api/v1/events", "", true)
			if status != http.StatusOK {
				t.Fatalf("list status=%d response=%#v", status, response)
			}
			assertEventDocumentRedacted(t, eventResponseDocument(t, response, true))

			status, response = callEventViewHandler(t, api, "admin", "/api/v1/events/"+event.EventID, "", false)
			if status != http.StatusOK {
				t.Fatalf("get status=%d response=%#v", status, response)
			}
			assertEventDocumentRedacted(t, eventResponseDocument(t, response, false))
		})
	}
}

func TestListEventsCursorModeReturnsOpaqueNextCursorAndForwardsKeyset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _, event := newEventViewTestAPI(t, false)
	stub := &cursorEventAnalyticsStub{event: analytics.FromAttackEvent(event)}
	api.analytics = stub

	status, response := callEventViewHandler(t, api, "admin", "/api/v1/events", "?pagination=cursor&page_size=1", true)
	if status != http.StatusOK {
		t.Fatalf("first page status=%d response=%#v", status, response)
	}
	data := response["data"].(map[string]any)
	token, ok := data["next_cursor"].(string)
	if !ok || token == "" || data["total"] != nil || data["has_more"] != true || data["total_known"] != false {
		t.Fatalf("unexpected cursor response: %#v", data)
	}
	decoded, err := analytics.DecodeCursor(token)
	if err != nil || decoded.EventID != event.EventID {
		t.Fatalf("decode next cursor=%#v err=%v", decoded, err)
	}
	if !stub.lastFilter.CursorMode || stub.lastFilter.Cursor != nil || stub.lastFilter.Offset != 0 || stub.lastFilter.Limit != 1 {
		t.Fatalf("first page filter=%#v", stub.lastFilter)
	}

	status, response = callEventViewHandler(t, api, "admin", "/api/v1/events", "?pagination=cursor&page_size=1&cursor="+token, true)
	if status != http.StatusOK {
		t.Fatalf("next page status=%d response=%#v", status, response)
	}
	if stub.lastFilter.Cursor == nil || stub.lastFilter.Cursor.EventID != event.EventID || !stub.lastFilter.Cursor.EventTime.Equal(event.Timestamp) {
		t.Fatalf("next page did not forward decoded cursor: %#v", stub.lastFilter)
	}
}

func TestListEventsRejectsMalformedPaginationBeforeAnalyticsQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _, event := newEventViewTestAPI(t, false)
	stub := &cursorEventAnalyticsStub{event: analytics.FromAttackEvent(event)}
	api.analytics = stub
	status, response := callEventViewHandler(t, api, "admin", "/api/v1/events", "?pagination=cursor&cursor=not%2Bbase64", true)
	if status != http.StatusBadRequest || response["code"] != "INVALID_PAGINATION" || stub.calls != 0 {
		t.Fatalf("status=%d response=%#v calls=%d", status, response, stub.calls)
	}
	if message, _ := response["message"].(string); strings.Contains(message, "游标") || strings.Contains(message, "ClickHouse") {
		t.Fatalf("implementation detail leaked in user message: %q", message)
	}
}

func TestListEventsRejectsDeepLegacyPageBeforeAnalyticsQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _, event := newEventViewTestAPI(t, false)
	stub := &cursorEventAnalyticsStub{event: analytics.FromAttackEvent(event)}
	api.analytics = stub
	status, response := callEventViewHandler(t, api, "admin", "/api/v1/events", "?page=101&page_size=200", true)
	if status != http.StatusBadRequest || response["code"] != "PAGE_TOO_DEEP" || stub.calls != 0 {
		t.Fatalf("status=%d response=%#v calls=%d", status, response, stub.calls)
	}
	if message, _ := response["message"].(string); strings.Contains(message, "游标") || strings.Contains(message, "ClickHouse") {
		t.Fatalf("implementation detail leaked in user message: %q", message)
	}
}

func TestListEventsPageNavigationCanReuseExactTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _, event := newEventViewTestAPI(t, false)
	stub := &cursorEventAnalyticsStub{event: analytics.FromAttackEvent(event)}
	api.analytics = stub

	status, response := callEventViewHandler(t, api, "admin", "/api/v1/events", "?pagination=page&page=2&page_size=10&include_total=false", true)
	if status != http.StatusOK {
		t.Fatalf("status=%d response=%#v", status, response)
	}
	data := response["data"].(map[string]any)
	if stub.lastFilter.CursorMode || !stub.lastFilter.SkipTotal || stub.lastFilter.Offset != 10 || stub.lastFilter.Limit != 10 {
		t.Fatalf("page navigation filter=%#v", stub.lastFilter)
	}
	if data["total"] != nil || data["total_known"] != false || data["page"].(float64) != 2 || data["max_page"].(float64) != maxEventPage {
		t.Fatalf("unexpected page response: %#v", data)
	}
}

func TestListEventsMySQLCursorWalksWithoutTotalCountContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db, newest := newEventViewTestAPI(t, false)
	newest.Timestamp = time.Date(2026, 8, 12, 6, 0, 0, 123000000, time.UTC)
	if err := db.Save(&newest).Error; err != nil {
		t.Fatal(err)
	}
	older := newest
	older.EventID = uuid.NewString()
	older.Timestamp = newest.Timestamp.Add(-time.Second)
	older.CreatedAt = older.Timestamp
	if err := db.Create(&older).Error; err != nil {
		t.Fatal(err)
	}

	status, response := callEventViewHandler(t, api, "admin", "/api/v1/events", "?pagination=cursor&page_size=1", true)
	if status != http.StatusOK {
		t.Fatalf("first page status=%d response=%#v", status, response)
	}
	first := response["data"].(map[string]any)
	firstItems := first["items"].([]any)
	if len(firstItems) != 1 || firstItems[0].(map[string]any)["event_id"] != newest.EventID || first["total"] != nil || first["has_more"] != true {
		t.Fatalf("unexpected first page: %#v", first)
	}
	token := first["next_cursor"].(string)
	status, response = callEventViewHandler(t, api, "admin", "/api/v1/events", "?pagination=cursor&page_size=1&cursor="+token, true)
	if status != http.StatusOK {
		t.Fatalf("second page status=%d response=%#v", status, response)
	}
	second := response["data"].(map[string]any)
	secondItems := second["items"].([]any)
	if len(secondItems) != 1 || secondItems[0].(map[string]any)["event_id"] != older.EventID || second["has_more"] != false || second["next_cursor"] != nil {
		t.Fatalf("unexpected second page: %#v", second)
	}
}

func TestViewerCannotRevealSensitiveEventEvidence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db, event := newEventViewTestAPI(t, false)
	api.cfg.RedactSensitiveEvents = true
	status, response := callEventViewHandler(t, api, "viewer", "/api/v1/events/"+event.EventID, "?include_sensitive=true", false)
	if status != http.StatusForbidden || response["code"] != "SENSITIVE_EVENT_FORBIDDEN" {
		t.Fatalf("status=%d response=%#v", status, response)
	}
	var count int64
	if err := db.Model(&store.AuditLog{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("denied reveal created audit=%d err=%v", count, err)
	}
}

func TestOperatorSensitiveEventRevealIsExplicitAndAuditedOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db, event := newEventViewTestAPI(t, false)
	api.cfg.RedactSensitiveEvents = true
	status, response := callEventViewHandler(t, api, "operator", "/api/v1/events/"+event.EventID, "?include_sensitive=true", false)
	if status != http.StatusOK {
		t.Fatalf("status=%d response=%#v", status, response)
	}
	document := eventResponseDocument(t, response, false)
	encoded, _ := json.Marshal(document)
	if !strings.Contains(string(encoded), "server-secret") || document["sensitive_visible"] != true || document["evidence_redacted"] != false || document["sensitive_redaction_enabled"] != true || document["sensitive_reveal_audited"] != true {
		t.Fatalf("explicit reveal did not return raw evidence: %s", encoded)
	}
	var logs []store.AuditLog
	if err := db.Where("object = ? AND action = ?", "/api/v1/events:sensitive", "READ").Find(&logs).Error; err != nil || len(logs) != 1 {
		t.Fatalf("sensitive reveal audits=%d err=%v", len(logs), err)
	}
	if !strings.Contains(string(logs[0].Detail), event.EventID) || !strings.Contains(string(logs[0].Detail), `"scope":"event"`) {
		t.Fatalf("audit detail does not identify event reveal: %s", logs[0].Detail)
	}
}

func TestOperatorSensitiveListRevealIsAuditedForMySQLAndClickHouse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, useAnalytics := range []bool{false, true} {
		name := "mysql"
		if useAnalytics {
			name = "clickhouse"
		}
		t.Run(name, func(t *testing.T) {
			api, db, _ := newEventViewTestAPI(t, useAnalytics)
			api.cfg.RedactSensitiveEvents = true
			status, response := callEventViewHandler(t, api, "operator", "/api/v1/events", "?include_sensitive=true", true)
			if status != http.StatusOK {
				t.Fatalf("status=%d response=%#v", status, response)
			}
			document := eventResponseDocument(t, response, true)
			encoded, _ := json.Marshal(document)
			if !strings.Contains(string(encoded), "server-secret") || document["sensitive_visible"] != true || document["evidence_redacted"] != false || document["sensitive_redaction_enabled"] != true || document["sensitive_reveal_audited"] != true {
				t.Fatalf("explicit list reveal metadata/evidence=%s", encoded)
			}
			var logs []store.AuditLog
			if err := db.Where("object = ? AND action = ?", "/api/v1/events:sensitive", "READ").Find(&logs).Error; err != nil || len(logs) != 1 {
				t.Fatalf("sensitive list reveal audits=%d err=%v", len(logs), err)
			}
			if !strings.Contains(string(logs[0].Detail), `"scope":"list"`) {
				t.Fatalf("audit detail does not identify list reveal: %s", logs[0].Detail)
			}
		})
	}
}

func TestRealtimeEventViewHonorsEnabledRedactionPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _, event := newEventViewTestAPI(t, false)
	api.cfg.RedactSensitiveEvents = true
	view := api.realtimeEventView(event)
	message, err := json.Marshal(WSMessage{Type: "event.new", Payload: view, TS: time.Now().Unix()})
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"server-secret", "raw-token", "session=private"} {
		if strings.Contains(string(message), secret) {
			t.Fatalf("WebSocket event.new leaked %q: %s", secret, message)
		}
	}
	if !strings.Contains(string(message), `"evidence_redacted":true`) || !strings.Contains(string(message), `"sensitive_visible":false`) || !strings.Contains(string(message), `"sensitive_redaction_enabled":true`) || !strings.Contains(string(message), `"sensitive_reveal_audited":false`) || !strings.Contains(string(message), `"sha256"`) {
		t.Fatalf("WebSocket event.new is not an evidence-friendly redacted view: %s", message)
	}
}

func TestRealtimeEventViewDefaultsToRawEvidenceWithoutRevealAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _, event := newEventViewTestAPI(t, false)
	view := api.realtimeEventView(event)
	message, err := json.Marshal(WSMessage{Type: "event.new", Payload: view, TS: time.Now().Unix()})
	if err != nil {
		t.Fatal(err)
	}
	for _, evidence := range []string{"server-secret", "raw-token", "session=private"} {
		if !strings.Contains(string(message), evidence) {
			t.Fatalf("WebSocket event.new lost default raw evidence %q: %s", evidence, message)
		}
	}
	if !strings.Contains(string(message), `"evidence_redacted":false`) || !strings.Contains(string(message), `"sensitive_visible":true`) || !strings.Contains(string(message), `"sensitive_redaction_enabled":false`) || !strings.Contains(string(message), `"sensitive_reveal_audited":false`) {
		t.Fatalf("WebSocket default raw disclosure metadata is ambiguous: %s", message)
	}
}

func TestDefaultRawPolicyDoesNotRequireIncludeSensitiveOrCreateRevealAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db, event := newEventViewTestAPI(t, false)
	for _, role := range []string{"viewer", "operator", "admin"} {
		status, response := callEventViewHandler(t, api, role, "/api/v1/events/"+event.EventID, "?include_sensitive=true", false)
		if status != http.StatusOK {
			t.Fatalf("role=%s status=%d response=%#v", role, status, response)
		}
		assertEventDocumentRawDefault(t, eventResponseDocument(t, response, false))
	}
	var count int64
	if err := db.Model(&store.AuditLog{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("default raw policy created reveal audits=%d err=%v", count, err)
	}
}
