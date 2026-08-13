package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

type credentialAnalyticsStub struct {
	analytics.Store
	page analytics.CredentialPage
}

func (stub credentialAnalyticsStub) ListCredentials(context.Context, analytics.CredentialFilter) (analytics.CredentialPage, error) {
	return stub.page, nil
}

func credentialTestEvent(eventType, service string, payload map[string]any) store.AttackEvent {
	raw, _ := json.Marshal(payload)
	return store.AttackEvent{
		EventID: "event-1", NodeID: "node-1", EventType: eventType, Service: service,
		Timestamp: time.Unix(1_700_000_000, 0), SrcIP: "203.0.113.9", Payload: datatypes.JSON(raw),
	}
}

func TestBuildCredentialResourceFiltersDirtyEvents(t *testing.T) {
	tests := []store.AttackEvent{
		credentialTestEvent("decoy.credential", "decoy", map[string]any{"path": "/tmp/.env", "action": "opened"}),
		credentialTestEvent("telnet.credential", "telnet", map[string]any{"username": "", "password": ""}),
		credentialTestEvent("web.credential", "http", map[string]any{"username": "null", "password": "undefined"}),
		credentialTestEvent("web.credential", "http", map[string]any{"username": "admin\nroot", "password": "secret\x00value"}),
		credentialTestEvent("web.credential", "http", map[string]any{"username": "请输入用户名", "password": "请输入密码"}),
		credentialTestEvent("web.credential", "http", map[string]any{"username": `{"field":"username"}`, "password": "<script>alert(1)</script>"}),
	}
	for _, event := range tests {
		if item, valid := buildCredentialResource(event, nil); valid {
			t.Fatalf("dirty event %s produced resource %#v", event.EventType, item)
		}
	}
}

func TestBuildCredentialResourcePreservesProtocolVariants(t *testing.T) {
	tests := []struct {
		name      string
		event     store.AttackEvent
		username  string
		password  string
		response  string
		mechanism string
		service   string
	}{
		{name: "redis password only", event: credentialTestEvent("redis.credential", "redis", map[string]any{"password": "redis-secret"}), password: "redis-secret", service: "Redis"},
		{name: "mysql auth response", event: credentialTestEvent("mysql.credential", "mysql", map[string]any{"username": "root", "auth_response": "01020304"}), username: "root", response: "01020304", service: "MySQL 数据库"},
		{name: "rdp username", event: credentialTestEvent("rdp.username", "rdp", map[string]any{"username": "administrator"}), username: "administrator", service: "Windows RDP"},
		{name: "smb ntlm", event: credentialTestEvent("smb.authentication", "smb", map[string]any{"username": "backup", "nt_response": "aabb", "mechanism": "NTLM"}), username: "backup", response: "aabb", mechanism: "NTLM", service: "SMB 文件共享"},
		{name: "pop3 digest", event: credentialTestEvent("pop3.authentication", "pop3", map[string]any{"username": "mail", "digest": "deadbeef", "mechanism": "APOP"}), username: "mail", response: "deadbeef", mechanism: "APOP", service: "POP3 邮件服务"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item, valid := buildCredentialResource(test.event, nil)
			if !valid {
				t.Fatal("valid protocol credential was filtered")
			}
			if item.Username != test.username || item.Password != test.password || item.AuthResponse != test.response || (test.mechanism != "" && item.Mechanism != test.mechanism) || item.ServiceName != test.service {
				t.Fatalf("unexpected credential resource: %#v", item)
			}
		})
	}
}

func TestCredentialResourceSearchTopAndPagination(t *testing.T) {
	items := []credentialResource{
		{EventID: "1", Username: "admin", Password: "one", SourceIP: "203.0.113.1", ServiceName: "SSH 服务"},
		{EventID: "2", Username: "admin", Password: "two", SourceIP: "203.0.113.2", ServiceName: "Redis"},
		{EventID: "3", Username: "root", Password: "one", SourceIP: "203.0.113.3", ServiceName: "MySQL 数据库"},
	}
	if !credentialResourceMatches(items[2], "mysql", false) || credentialResourceMatches(items[0], "not-found", false) {
		t.Fatal("credential search did not use normalized service fields")
	}
	if credentialResourceMatches(items[0], "one", false) || !credentialResourceMatches(items[0], "one", true) {
		t.Fatal("masked credential search exposed a password oracle")
	}
	topUsers := topCredentialValues(items, func(item credentialResource) string { return item.Username })
	if len(topUsers) != 2 || topUsers[0].Value != "admin" || topUsers[0].Count != 2 {
		t.Fatalf("top usernames = %#v", topUsers)
	}
	pageItems := paginateCredentialResources(items, 2, 2)
	if len(pageItems) != 1 || pageItems[0].EventID != "3" {
		t.Fatalf("second page = %#v", pageItems)
	}
}

func credentialResourceTestAPI(t *testing.T) (*API, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&store.AttackEvent{}, &store.PotService{}, &store.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	stub := credentialAnalyticsStub{page: analytics.CredentialPage{
		Items: []analytics.CredentialResource{{EventID: uuid.NewString(), NodeID: uuid.NewString(), EventType: "web.credential", EventTime: time.Now(), SourceIP: "203.0.113.9", Username: "admin", Password: "super-secret", AuthResponse: "deadbeef", Service: "http"}},
		Total: 1, TopUsernames: []analytics.CredentialCount{{Value: "admin", Count: 1}}, TopPasswords: []analytics.CredentialCount{{Value: "super-secret", Count: 1}},
	}}
	return &API{db: db, analytics: stub}, db
}

func callCredentialResources(t *testing.T, api *API, role, query string) (int, map[string]any) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/credential-resources"+query, nil)
	c.Set(userContextKey, AuthUser{ID: uuid.NewString(), Username: role, Role: role})
	api.listCredentialResources(c)
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return recorder.Code, response
}

func TestCredentialResourcesMaskSensitiveValuesByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, _ := credentialResourceTestAPI(t)
	status, response := callCredentialResources(t, api, "admin", "?page_size=20")
	if status != http.StatusOK {
		t.Fatalf("status=%d response=%#v", status, response)
	}
	data := responseData(t, response)
	items := data["items"].([]any)
	item := items[0].(map[string]any)
	if item["password"] != maskedCredentialValue || item["auth_response"] != maskedCredentialValue {
		t.Fatalf("sensitive values leaked by default: %#v", item)
	}
	if data["sensitive_visible"] != false {
		t.Fatalf("sensitive_visible=%#v", data["sensitive_visible"])
	}
}

func TestViewerCannotRevealCredentialValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := credentialResourceTestAPI(t)
	status, response := callCredentialResources(t, api, "viewer", "?include_sensitive=true")
	if status != http.StatusForbidden || response["code"] != "SENSITIVE_CREDENTIALS_FORBIDDEN" {
		t.Fatalf("status=%d response=%#v", status, response)
	}
	var audits int64
	if err := db.Model(&store.AuditLog{}).Count(&audits).Error; err != nil || audits != 0 {
		t.Fatalf("denied reveal created audit=%d err=%v", audits, err)
	}
}

func TestOperatorRevealIsExplicitAndAudited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api, db := credentialResourceTestAPI(t)
	status, response := callCredentialResources(t, api, "operator", "?include_sensitive=true")
	if status != http.StatusOK {
		t.Fatalf("status=%d response=%#v", status, response)
	}
	data := responseData(t, response)
	item := data["items"].([]any)[0].(map[string]any)
	if item["password"] != "super-secret" || item["auth_response"] != "deadbeef" || data["sensitive_visible"] != true {
		t.Fatalf("explicit reveal response=%#v", data)
	}
	var log store.AuditLog
	if err := db.Where("object = ?", "/api/v1/credential-resources:sensitive").First(&log).Error; err != nil {
		t.Fatalf("sensitive reveal audit missing: %v", err)
	}
}
