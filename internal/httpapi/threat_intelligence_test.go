package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/config"
	"github.com/honeynet/honeynet/internal/store"
	"github.com/honeynet/honeynet/internal/threatintel"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestThreatIntelligenceStatusAndInvalidQuery(t *testing.T) {
	db := testHTTPDB(t)
	api := &API{cfg: config.Config{}, db: db, threatIntel: mustThreatIntelManager(t, threatintel.Config{})}
	router := testAuthedRouter(t, api)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/intel/database/status", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"enabled":false`) {
		t.Fatalf("status response = %d %s", recorder.Code, recorder.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/intel/database/query?ip=not-an-ip", nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "INVALID_IP") {
		t.Fatalf("query response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestThreatIntelligenceEnrichmentPreservesExistingEventTags(t *testing.T) {
	manager := mustThreatIntelManager(t, threatintel.Config{})
	// This test exercises view behavior with the known-unloaded state. Loaded
	// database matching is covered by the threatintel package's authenticated
	// fixture and the real publisher database compatibility smoke test.
	api := &API{threatIntel: manager}
	event := store.AttackEvent{EventID: "00000000-0000-4000-8000-000000000099", SrcIP: "203.0.113.8", Timestamp: time.Now(), Tags: datatypes.JSON(`["existing"]`)}
	views := api.enrichEventThreatIntelligence([]attackEventView{{AttackEvent: event}})
	var tags []string
	if err := json.Unmarshal(views[0].Tags, &tags); err != nil || len(tags) != 1 || tags[0] != "existing" {
		t.Fatalf("event tags changed without a loaded database: %#v, %v", tags, err)
	}
}

func mustThreatIntelManager(t *testing.T, cfg threatintel.Config) *threatintel.Manager {
	t.Helper()
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = filepath.Join(t.TempDir(), "threat-intelligence.db")
	}
	manager, err := threatintel.NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func testHTTPDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&store.User{}, &store.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func testAuthedRouter(t *testing.T, api *API) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	user := store.User{Base: store.NewBase(), Username: "admin", PasswordHash: "unused", Role: "admin", Enabled: true, TokenVersion: 1}
	if err := api.db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	api.tokens = NewTokenManager("a-test-secret-that-is-long-enough", time.Minute).WithUserStore(api.db)
	token, _, err := api.tokens.Issue(AuthUser{ID: user.ID, Username: user.Username, Role: user.Role, TokenVersion: user.TokenVersion})
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Request.Header.Set("Authorization", "Bearer "+token) })
	group := router.Group("/api/v1")
	group.Use(api.tokens.Middleware())
	group.GET("/intel/database/status", api.threatIntelStatus)
	group.GET("/intel/database/query", api.queryThreatIntel)
	return router
}
