package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager("a-test-secret-that-is-long-enough", time.Minute)
	want := AuthUser{ID: "user-1", Username: "admin", Role: "admin"}
	raw, _, err := manager.Issue(want)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	got, err := manager.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got != want {
		t.Fatalf("Parse() = %#v, want %#v", got, want)
	}
}

func TestInvalidToken(t *testing.T) {
	manager := NewTokenManager("a-test-secret-that-is-long-enough", time.Minute)
	if _, err := manager.Parse("not-a-token"); err == nil {
		t.Fatal("Parse() accepted invalid token")
	}
}

func authTestDB(t *testing.T) *gorm.DB {
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

func TestTokenManagerRevokesChangedAccountState(t *testing.T) {
	db := authTestDB(t)
	user := store.User{Base: store.NewBase(), Username: "operator", PasswordHash: "unused", Role: "operator", Enabled: true, TokenVersion: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	manager := NewTokenManager("a-test-secret-that-is-long-enough", time.Minute).WithUserStore(db)
	raw, _, err := manager.Issue(AuthUser{ID: user.ID, Username: user.Username, Role: user.Role, TokenVersion: user.TokenVersion})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Parse(raw); err != nil {
		t.Fatalf("fresh token rejected: %v", err)
	}
	if err := db.Model(&user).Updates(map[string]any{"role": "viewer", "token_version": gorm.Expr("token_version + 1")}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Parse(raw); err == nil {
		t.Fatal("token remained valid after role/token version change")
	}
}

func TestLogoutRevokesCurrentToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := authTestDB(t)
	user := store.User{Base: store.NewBase(), Username: "admin", PasswordHash: "unused", Role: "admin", Enabled: true, TokenVersion: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	manager := NewTokenManager("a-test-secret-that-is-long-enough", time.Minute).WithUserStore(db)
	raw, _, err := manager.Issue(AuthUser{ID: user.ID, Username: user.Username, Role: user.Role, TokenVersion: user.TokenVersion})
	if err != nil {
		t.Fatal(err)
	}
	api := &API{db: db, tokens: manager}
	router := gin.New()
	router.POST("/auth/logout", manager.Middleware(), api.audit(), api.logout)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	request.Header.Set("Authorization", "Bearer "+raw)
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("logout status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if _, err := manager.Parse(raw); err == nil {
		t.Fatal("logout did not revoke current token")
	}
	var auditCount int64
	if err := db.Model(&store.AuditLog{}).Where("object = ?", "/auth/logout").Count(&auditCount).Error; err != nil || auditCount != 1 {
		t.Fatalf("logout audits=%d err=%v", auditCount, err)
	}
}

func TestLoginAttemptLimiterLocksAndExpires(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	limiter := newLoginAttemptLimiter()
	limiter.now = func() time.Time { return now }
	for index := 1; index <= loginIPFailureLimit; index++ {
		locked := limiter.failure("203.0.113.5")
		if locked != (index == loginIPFailureLimit) {
			t.Fatalf("failure %d locked=%v", index, locked)
		}
	}
	if limiter.allow("203.0.113.5") {
		t.Fatal("rate limiter allowed locked IP")
	}
	now = now.Add(loginIPLockDuration + time.Second)
	if !limiter.allow("203.0.113.5") {
		t.Fatal("rate limiter did not release expired lock")
	}
}

func TestAccountFailureLockPersists(t *testing.T) {
	db := authTestDB(t)
	user := store.User{Base: store.NewBase(), Username: "viewer", PasswordHash: "unused", Role: "viewer", Enabled: true, TokenVersion: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	api := &API{db: db}
	now := time.Unix(1_700_000_000, 0)
	for index := 1; index <= accountFailureLimit; index++ {
		locked := api.recordFailedLogin(user.ID, now)
		if locked != (index == accountFailureLimit) {
			t.Fatalf("failure %d locked=%v", index, locked)
		}
	}
	if err := db.First(&user, "id = ?", user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if user.FailedLoginCount != accountFailureLimit || user.LockedUntil == nil || !user.LockedUntil.Equal(now.Add(accountLockDuration)) {
		t.Fatalf("persisted lock=%#v", user)
	}
}

func TestAccountFailureCounterResetsOutsideWindow(t *testing.T) {
	db := authTestDB(t)
	user := store.User{Base: store.NewBase(), Username: "operator-2", PasswordHash: "unused", Role: "operator", Enabled: true, TokenVersion: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	api := &API{db: db}
	now := time.Unix(1_700_000_000, 0)
	for index := 0; index < accountFailureLimit-1; index++ {
		if api.recordFailedLogin(user.ID, now) {
			t.Fatal("account locked before threshold")
		}
	}
	if api.recordFailedLogin(user.ID, now.Add(loginFailureWindow+time.Second)) {
		t.Fatal("stale failures were counted in a new window")
	}
	if err := db.First(&user, "id = ?", user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if user.FailedLoginCount != 1 || user.LockedUntil != nil {
		t.Fatalf("failure window was not reset: %#v", user)
	}
}

func TestNodeInstallerRouteHasNoSideEffectingGET(t *testing.T) {
	source, err := os.ReadFile("server.go")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(source, []byte(`authed.GET("/nodes/:id/install"`)) {
		t.Fatal("side-effecting GET node installer route is still registered")
	}
	if !bytes.Contains(source, []byte(`authed.POST("/nodes/:id/install", requireRoles("admin", "operator"), api.nodeInstall)`)) {
		t.Fatal("role-protected POST node installer route is missing")
	}
}

func TestAuditIncludesRequestIDStatusAndChangeSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := authTestDB(t)
	user := store.User{Base: store.NewBase(), Username: "audit-admin", PasswordHash: "unused", Role: "admin", Enabled: true, TokenVersion: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	manager := NewTokenManager("a-test-secret-that-is-long-enough", time.Minute).WithUserStore(db)
	raw, _, err := manager.Issue(AuthUser{ID: user.ID, Username: user.Username, Role: user.Role, TokenVersion: user.TokenVersion})
	if err != nil {
		t.Fatal(err)
	}
	api := &API{db: db, tokens: manager, hub: NewHub()}
	router := gin.New()
	router.POST("/users", manager.Middleware(), api.audit(), func(c *gin.Context) {
		setAuditChange(c, "user", "new-user", nil, map[string]any{"role": "viewer"})
		c.Status(http.StatusCreated)
	})
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("{}"))
	request.Header.Set("Authorization", "Bearer "+raw)
	request.Header.Set("X-Request-ID", "audit-test-request")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated || recorder.Header().Get("X-Request-ID") != "audit-test-request" {
		t.Fatalf("unexpected response status=%d request_id=%q", recorder.Code, recorder.Header().Get("X-Request-ID"))
	}
	var item store.AuditLog
	if err := db.Last(&item).Error; err != nil {
		t.Fatal(err)
	}
	var detail map[string]any
	if err := json.Unmarshal(item.Detail, &detail); err != nil {
		t.Fatal(err)
	}
	if detail["request_id"] != "audit-test-request" || int(detail["status"].(float64)) != http.StatusCreated {
		t.Fatalf("unexpected audit detail: %#v", detail)
	}
	change, ok := detail["change"].(map[string]any)
	if !ok || change["object_type"] != "user" || change["object_id"] != "new-user" {
		t.Fatalf("unexpected change detail: %#v", detail["change"])
	}
}

func TestLastEnabledAdminCannotBeRemoved(t *testing.T) {
	db := authTestDB(t)
	api := &API{db: db}
	admin := store.User{Base: store.NewBase(), Username: "only-admin", PasswordHash: "unused", Role: "admin", Enabled: true, TokenVersion: 1}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if !api.removingLastEnabledAdmin(admin, "viewer", true) || !api.removingLastEnabledAdmin(admin, "admin", false) {
		t.Fatal("last enabled admin removal was not blocked")
	}
	second := store.User{Base: store.NewBase(), Username: "second-admin", PasswordHash: "unused", Role: "admin", Enabled: true, TokenVersion: 1}
	if err := db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}
	if api.removingLastEnabledAdmin(admin, "viewer", true) {
		t.Fatal("admin change remained blocked after another enabled admin existed")
	}
}
