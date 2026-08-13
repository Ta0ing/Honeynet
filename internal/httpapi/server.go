package httpapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/honeynet/honeynet/internal/agentupdate"
	aimodule "github.com/honeynet/honeynet/internal/ai"
	"github.com/honeynet/honeynet/internal/alerting"
	"github.com/honeynet/honeynet/internal/analytics"
	"github.com/honeynet/honeynet/internal/config"
	"github.com/honeynet/honeynet/internal/decoyconfig"
	"github.com/honeynet/honeynet/internal/detection"
	"github.com/honeynet/honeynet/internal/geoip"
	"github.com/honeynet/honeynet/internal/nodepki"
	"github.com/honeynet/honeynet/internal/store"
	"github.com/honeynet/honeynet/internal/threatintel"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type API struct {
	cfg          config.Config
	db           *gorm.DB
	tokens       *TokenManager
	hub          *Hub
	agents       *AgentGateway
	alerts       *alerting.Dispatcher
	pki          *nodepki.Authority
	updateSigner *agentupdate.Signer
	upgrades     *UpgradeManager
	geoIP        eventGeoResolver
	threatIntel  *threatintel.Manager
	detection    *detectionRuntime
	ai           *aimodule.Runtime
	aiSettings   *aimodule.SettingsStore
	aiAgent      *aimodule.Agent
	analytics    analytics.Store
	analyticsCfg config.AnalyticsConfig
	loginLimiter *loginAttemptLimiter
}

type eventGeoResolver interface {
	Lookup(string) (geoip.Result, error)
}

type Runtime struct {
	Console *gin.Engine
	Agent   *gin.Engine
}

func New(cfg config.Config, db *gorm.DB) *gin.Engine {
	return NewWithContext(context.Background(), cfg, db)
}

func NewWithContext(ctx context.Context, cfg config.Config, db *gorm.DB) *gin.Engine {
	authority, err := nodepki.LoadOrCreate(cfg.PKIDir, cfg.AgentTLSNames, cfg.AgentCertValidity)
	if err != nil {
		panic(fmt.Sprintf("initialize node PKI: %v", err))
	}
	return NewRuntime(ctx, cfg, db, authority).Console
}

func NewRuntime(ctx context.Context, cfg config.Config, db *gorm.DB, authority *nodepki.Authority) Runtime {
	return NewRuntimeWithAnalytics(ctx, cfg, db, authority, nil, config.DefaultAnalyticsConfig())
}

func NewRuntimeWithAnalytics(ctx context.Context, cfg config.Config, db *gorm.DB, authority *nodepki.Authority, eventStore analytics.Store, analyticsCfg config.AnalyticsConfig) Runtime {
	dispatcher := alerting.NewDispatcher(db, cfg.PublicURL)
	dispatcher.Start(ctx)
	geoResolver, err := geoip.Open(cfg.IPIPDBPath, cfg.IPIPLanguage)
	if err != nil {
		panic(fmt.Sprintf("initialize IPIP geolocation: %v", err))
	}
	threatIntelManager, err := threatintel.NewManager(threatintel.Config{
		Enabled: cfg.ThreatIntelEnabled, DatabasePath: cfg.ThreatIntelDBPath,
		DownloadURL: cfg.ThreatIntelDownloadURL, ArchivePassword: cfg.ThreatIntelArchivePassword,
		UpdateInterval: cfg.ThreatIntelUpdateInterval,
	})
	if err != nil {
		panic(fmt.Sprintf("initialize threat intelligence database: %v", err))
	}
	threatIntelManager.Start(ctx)
	signer, err := agentupdate.LoadOrCreateSigner(cfg.PKIDir)
	if err != nil {
		panic(fmt.Sprintf("initialize Agent update signer: %v", err))
	}
	aiSettings, err := aimodule.NewSettingsStore(db, cfg.JWTSecret)
	if err != nil {
		panic(fmt.Sprintf("initialize AI settings store: %v", err))
	}
	storedAISettings, err := aiSettings.LoadOrCreate(ctx, aimodule.Config{Enabled: cfg.AIEnabled, Provider: cfg.AIProvider, BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel, Timeout: cfg.AITimeout, SendRawPacket: cfg.AISendRawPacket})
	if err != nil {
		panic(fmt.Sprintf("load AI settings: %v", err))
	}
	aiRuntime, err := aimodule.NewRuntime(storedAISettings.Config)
	if err != nil {
		panic(fmt.Sprintf("initialize AI runtime: %v", err))
	}
	var eventGeoIP eventGeoResolver
	if geoResolver != nil {
		eventGeoIP = geoResolver
	}
	api := &API{cfg: cfg, db: db, tokens: NewTokenManager(cfg.JWTSecret, cfg.JWTExpires).WithUserStore(db), hub: NewHub(), alerts: dispatcher, pki: authority, updateSigner: signer, geoIP: eventGeoIP, threatIntel: threatIntelManager, detection: &detectionRuntime{}, ai: aiRuntime, aiSettings: aiSettings, analytics: eventStore, analyticsCfg: analyticsCfg, loginLimiter: newLoginAttemptLimiter()}
	api.aiAgent, err = api.newAIAgent()
	if err != nil {
		panic(fmt.Sprintf("initialize AI Agent runtime: %v", err))
	}
	api.agents = NewAgentGateway(db, api.hub)
	api.agents.SetUpdateTrust(signer.PublicKey(), signer.KeyID())
	api.upgrades = NewUpgradeManager(db, cfg, signer, api.agents, api.hub)
	api.agents.SetUpgradeManager(api.upgrades)
	api.upgrades.Start(ctx)
	if report, err := api.importConfiguredDetectionRules(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("detection rule import skipped: %v", err)
		}
	} else if report.Created+report.Updated > 0 {
		log.Printf("detection rules ready: %d created, %d updated, %d pending review", report.Created, report.Updated, report.Pending)
	}
	if err := api.refreshDetectionMatcher(); err != nil {
		panic(fmt.Sprintf("initialize detection matcher: %v", err))
	}
	if geoResolver != nil && eventStore == nil {
		log.Printf("IPIP geolocation enabled: %s (build %s, languages %s)", geoResolver.Path(), geoResolver.BuildTime().Format(time.RFC3339), strings.Join(geoResolver.Languages(), ","))
		go api.backfillEventGeo(ctx)
	}
	return Runtime{Console: api.consoleRouter(), Agent: api.agentRouter()}
}

func (api *API) consoleRouter() *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(gin.Recovery(), requestID(), securityHeaders())
	r.Use(cors.New(cors.Config{AllowOrigins: api.cfg.CORSOrigins, AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID"}, ExposeHeaders: []string{"X-Request-ID"}, MaxAge: 12 * time.Hour}))
	r.GET("/healthz", api.health)
	r.GET("/download/install.sh", api.installShell)
	r.GET("/download/install.ps1", api.installPowerShell)
	r.GET("/download/install-agent.sh", api.installShell)
	r.GET("/download/install-agent.ps1", api.installPowerShell)
	r.GET("/download/templates/:format/sha256", api.downloadTemplateBundleChecksum)
	r.GET("/download/templates/:format", api.downloadTemplateBundle)
	r.GET("/download/agent/:os/:arch/sha256", api.downloadAgentChecksum)
	r.GET("/download/agent/:os/:arch", api.downloadAgent)
	r.GET("/download/agent-guard/:os/:arch/sha256", api.downloadAgentGuardChecksum)
	r.GET("/download/agent-guard/:os/:arch", api.downloadAgentGuard)
	r.GET("/api/v1/ws", func(c *gin.Context) { api.hub.Serve(c, api.tokens) })

	v1 := r.Group("/api/v1")
	v1.POST("/auth/login", api.login)
	// Branding and image assets are deliberately public so the login page can
	// render its identity before a bearer token exists.
	v1.GET("/platform/branding", api.platformOEM)
	v1.GET("/platform/oem", api.platformOEM)
	v1.GET("/platform/assets/:kind", api.servePlatformOEMLogo)
	v1.GET("/platform/oem/logos/:kind", api.servePlatformOEMLogo)
	authed := v1.Group("")
	authed.Use(api.tokens.Middleware(), api.audit())
	authed.GET("/auth/profile", api.profile)
	authed.POST("/auth/logout", api.logout)
	authed.GET("/dashboard/summary", api.dashboardSummary)
	authed.GET("/dashboard/trends", api.dashboardTrends)
	authed.GET("/dashboard/top-attackers", api.dashboardTopAttackers)
	authed.GET("/analytics/status", api.analyticsStatus)
	authed.GET("/platform/settings", requireRoles("admin"), api.platformOEM)
	authed.PUT("/platform/settings", requireRoles("admin"), limitRequestBody(32<<10), api.updatePlatformOEM)
	authed.POST("/platform/assets/:kind", requireRoles("admin"), limitRequestBody(3<<20), api.uploadPlatformOEMLogo)
	authed.DELETE("/platform/assets/:kind", requireRoles("admin"), api.deletePlatformOEMLogo)
	// Compatibility aliases keep early Console clients functional while the
	// canonical settings/assets routes above remain the documented contract.
	authed.GET("/platform/config", requireRoles("admin"), api.platformOEM)
	authed.PUT("/platform/config", requireRoles("admin"), limitRequestBody(32<<10), api.updatePlatformOEM)
	authed.POST("/platform/config/logos/:kind", requireRoles("admin"), limitRequestBody(3<<20), api.uploadPlatformOEMLogo)
	authed.DELETE("/platform/config/logos/:kind", requireRoles("admin"), api.deletePlatformOEMLogo)
	authed.GET("/nodes", api.listNodes)
	authed.GET("/nodes/:id", api.getNode)
	authed.POST("/nodes", requireRoles("admin", "operator"), api.createNode)
	authed.PUT("/nodes/:id", requireRoles("admin", "operator"), api.updateNode)
	authed.DELETE("/nodes/:id", requireRoles("admin"), api.deleteNode)
	authed.GET("/nodes/:id/sense", api.getNodeSense)
	authed.PUT("/nodes/:id/sense", requireRoles("admin", "operator"), api.updateNodeSense)
	authed.POST("/nodes/:id/token", requireRoles("admin", "operator"), api.rotateNodeToken)
	authed.POST("/nodes/:id/install", requireRoles("admin", "operator"), api.nodeInstall)
	authed.GET("/pot-services", api.listPotServices)
	authed.GET("/pots", api.listPots)
	authed.GET("/pots/:id", api.getPot)
	authed.POST("/pots", requireRoles("admin", "operator"), api.createPot)
	authed.PUT("/pots/:id", requireRoles("admin", "operator"), api.updatePot)
	authed.DELETE("/pots/:id", requireRoles("admin", "operator"), api.deletePot)
	authed.POST("/pots/:id/start", requireRoles("admin", "operator"), api.potAction("running"))
	authed.POST("/pots/:id/stop", requireRoles("admin", "operator"), api.potAction("stopped"))
	authed.GET("/pot-templates", api.listPotTemplates)
	authed.POST("/pot-templates", requireRoles("admin", "operator"), api.createPotTemplate)
	authed.PUT("/pot-templates/:id", requireRoles("admin", "operator"), api.updatePotTemplate)
	authed.DELETE("/pot-templates/:id", requireRoles("admin", "operator"), api.deletePotTemplate)
	authed.GET("/decoys", api.listDecoys)
	authed.POST("/decoys", requireRoles("admin", "operator"), api.createDecoy)
	authed.PUT("/decoys/:id", requireRoles("admin", "operator"), api.updateDecoy)
	authed.DELETE("/decoys/:id", requireRoles("admin", "operator"), api.deleteDecoy)
	authed.GET("/events", api.listEvents)
	authed.GET("/events/:id", api.getEvent)
	authed.GET("/credential-resources", api.listCredentialResources)
	authed.GET("/alerts", api.listAlerts)
	authed.PUT("/alerts/:id/ack", requireRoles("admin", "operator"), api.ackAlert)
	authed.GET("/alert-rules", api.listAlertRules)
	authed.POST("/alert-rules", requireRoles("admin"), api.createAlertRule)
	authed.PUT("/alert-rules/:id", requireRoles("admin"), api.updateAlertRule)
	authed.GET("/detection-rules", api.listDetectionRules)
	authed.POST("/detection-rules", requireRoles("admin"), api.createDetectionRule)
	authed.PUT("/detection-rules/:id", requireRoles("admin"), api.updateDetectionRule)
	authed.DELETE("/detection-rules/:id", requireRoles("admin"), api.deleteDetectionRule)
	authed.POST("/detection-rules/:id/test", requireRoles("admin", "operator"), api.testDetectionRule)
	authed.POST("/detection-rules/import-builtin", requireRoles("admin"), api.importDetectionRules)
	authed.GET("/detection/pipeline/status", api.detectionPipelineStatus)
	authed.GET("/ai/status", api.aiStatus)
	authed.GET("/ai/agent/capabilities", api.aiAgentCapabilities)
	authed.GET("/ai/config", requireRoles("admin"), api.getAIConfig)
	authed.PUT("/ai/config", requireRoles("admin"), api.updateAIConfig)
	authed.POST("/ai/config/test", requireRoles("admin"), api.testAIConfig)
	authed.POST("/ai/agent/runs", requireRoles("admin", "operator"), api.runAIAgent)
	authed.GET("/ai/harness/runs", requireRoles("admin", "operator"), api.listAIHarnessRuns)
	authed.POST("/ai/harness/runs", requireRoles("admin", "operator"), api.createAIHarnessRun)
	authed.GET("/ai/rule-proposals", requireRoles("admin", "operator"), api.listDetectionRuleProposals)
	authed.POST("/ai/rule-proposals/:id/review", requireRoles("admin"), api.reviewDetectionRuleProposal)
	authed.POST("/ai/rule-proposals/:id/publish", requireRoles("admin"), api.publishDetectionRuleProposal)
	authed.POST("/ai/rule-proposals/:id/rollback", requireRoles("admin"), api.rollbackDetectionRuleProposal)
	authed.GET("/ai/rule-feedback", requireRoles("admin", "operator"), api.listDetectionRuleFeedback)
	authed.POST("/ai/rule-feedback", requireRoles("admin", "operator"), api.createDetectionRuleFeedback)
	authed.GET("/ai/analyses", requireRoles("admin", "operator"), api.listAIAnalyses)
	authed.POST("/events/:id/ai-analysis", requireRoles("admin", "operator"), api.analyzeEvent)
	authed.POST("/attackers/:ip/ai-profile", requireRoles("admin", "operator"), api.analyzeAttacker)
	authed.GET("/alert-channels", api.listAlertChannels)
	authed.POST("/alert-channels", requireRoles("admin"), api.createAlertChannel)
	authed.PUT("/alert-channels/:id", requireRoles("admin"), api.updateAlertChannel)
	authed.DELETE("/alert-channels/:id", requireRoles("admin"), api.deleteAlertChannel)
	authed.POST("/alert-channels/:id/test", requireRoles("admin"), api.testAlertChannel)
	authed.GET("/alert-deliveries", api.listAlertDeliveries)
	authed.POST("/alert-deliveries/:id/retry", requireRoles("admin", "operator"), api.retryAlertDelivery)
	authed.GET("/intel/iocs", api.listIOCs)
	authed.GET("/intel/feeds", api.listIntelFeeds)
	authed.POST("/intel/feeds", requireRoles("admin"), api.createIntelFeed)
	authed.PUT("/intel/feeds/:id", requireRoles("admin"), api.updateIntelFeed)
	authed.DELETE("/intel/feeds/:id", requireRoles("admin"), api.deleteIntelFeed)
	authed.GET("/intel/database/status", api.threatIntelStatus)
	authed.GET("/intel/database/query", api.queryThreatIntel)
	authed.POST("/intel/database/update", requireRoles("admin"), api.updateThreatIntel)
	authed.GET("/assets", api.listAssets)
	authed.GET("/users", requireRoles("admin"), api.listUsers)
	authed.POST("/users", requireRoles("admin"), api.createUser)
	authed.PUT("/users/:id", requireRoles("admin"), api.updateUser)
	authed.DELETE("/users/:id", requireRoles("admin"), api.deleteUser)
	authed.GET("/audit-logs", requireRoles("admin"), api.listAuditLogs)
	authed.GET("/agent-releases", api.listAgentReleases)
	authed.POST("/agent-releases/scan", requireRoles("admin"), api.scanAgentRelease)
	authed.GET("/upgrade-rollouts", api.listUpgradeRollouts)
	authed.POST("/upgrade-rollouts", requireRoles("admin"), api.createUpgradeRollout)
	authed.POST("/upgrade-rollouts/:id/resume", requireRoles("admin"), api.updateUpgradeRolloutStatus("running"))
	authed.POST("/upgrade-rollouts/:id/pause", requireRoles("admin"), api.updateUpgradeRolloutStatus("paused"))
	authed.POST("/upgrade-rollouts/:id/cancel", requireRoles("admin"), api.updateUpgradeRolloutStatus("cancelled"))

	api.serveFrontend(r)
	return r
}

func (api *API) agentRouter() *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(gin.Recovery(), requestID(), securityHeaders())
	r.GET("/healthz", api.health)
	agent := r.Group("/agent/v1")
	agent.POST("/register", limitRequestBody(64<<10), api.registerAgent)
	agent.POST("/certificates/renew", api.mtlsAuth(), limitRequestBody(4<<10), api.renewAgentCertificate)
	agent.POST("/certificates/activate", api.mtlsAuth(), okEmpty)
	agent.POST("/events:batch", api.mtlsAuth(), api.ingestEvents)
	agent.POST("/heartbeat", api.mtlsAuth(), limitRequestBody(1<<20), api.agentHeartbeat)
	agent.GET("/ws", api.mtlsAuth(), api.agents.Serve)
	agent.GET("/updates/:id", api.mtlsAuth(), api.downloadAgentUpdate)
	return r
}

func limitRequestBody(bytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, bytes)
		c.Next()
	}
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "same-origin")
		c.Next()
	}
}
func ok(c *gin.Context, data any)      { c.JSON(http.StatusOK, gin.H{"data": data}) }
func created(c *gin.Context, data any) { c.JSON(http.StatusCreated, gin.H{"data": data}) }
func okEmpty(c *gin.Context)           { ok(c, gin.H{"success": true}) }
func fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"code": code, "message": message, "request_id": c.GetString("request_id")})
}
func page(c *gin.Context) (int, int) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	s, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if p < 1 {
		p = 1
	}
	if s < 1 {
		s = 20
	}
	if s > 200 {
		s = 200
	}
	return p, s
}
func pageResult(items any, total int64, p, s int) gin.H {
	return gin.H{"items": items, "total": total, "page": p, "page_size": s}
}

func (a *API) health(c *gin.Context) {
	sqlDB, err := a.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		fail(c, 503, "DATABASE_UNAVAILABLE", "数据库连接异常")
		return
	}
	analyticsStatus := a.currentAnalyticsStatus(c.Request.Context())
	if a.analytics != nil && !analyticsStatus["healthy"].(bool) {
		fail(c, 503, "ANALYTICS_UNAVAILABLE", "安全分析引擎连接或数据结构异常")
		return
	}
	ok(c, gin.H{"status": "ok", "time": time.Now(), "version": a.cfg.Version, "ipv6_capable": true, "mysql": gin.H{"healthy": true, "role": "business"}, "analytics": analyticsStatus})
}

func (a *API) currentAnalyticsStatus(ctx context.Context) gin.H {
	if a.analytics == nil {
		return gin.H{"enabled": false, "healthy": false, "driver": "compatibility", "role": "security-analytics", "mode": "compatibility"}
	}
	status := a.analytics.Status(ctx)
	result := gin.H{
		"enabled": status.Enabled, "healthy": status.Healthy, "driver": "security-analytics",
		"role": "security-analytics", "mode": "primary", "database": a.analyticsCfg.Database,
		"table": a.analyticsCfg.Table, "schema_version": status.SchemaVersion,
	}
	if status.Database != "" {
		result["database"] = status.Database
	}
	if status.Table != "" {
		result["table"] = status.Table
	}
	if version := publicAnalyticsVersion(status.ServerVersion); version != "" {
		result["server_version"] = version
	}
	if status.LastWriteAt != nil {
		result["last_write_at"] = status.LastWriteAt
	}
	if status.Error != "" {
		// Repository status errors are intentionally generic and never include a
		// DSN. Preserve the distinction between connectivity and schema drift so
		// operators can take the correct recovery action from the console.
		result["error"] = status.Error
	}
	return result
}

// publicAnalyticsVersion keeps implementation-specific product and host names
// out of the management API while retaining the numeric version operators need
// for compatibility checks. Unknown formats are intentionally omitted rather
// than reflecting an arbitrary backend banner into the console.
func publicAnalyticsVersion(value string) string {
	const marker = "server version "
	lower := strings.ToLower(strings.TrimSpace(value))
	index := strings.Index(lower, marker)
	if index < 0 {
		return ""
	}
	fields := strings.Fields(value[index+len(marker):])
	if len(fields) == 0 {
		return ""
	}
	candidate := strings.Trim(fields[0], " ,;()[]")
	for _, character := range candidate {
		if (character < '0' || character > '9') && character != '.' && character != '-' {
			return ""
		}
	}
	return candidate
}

func (a *API) analyticsStatus(c *gin.Context) {
	result := a.currentAnalyticsStatus(c.Request.Context())
	mysqlStatus := gin.H{"healthy": false, "role": "business", "engine": "mysql"}
	if sqlDB, err := a.db.DB(); err == nil && sqlDB.PingContext(c.Request.Context()) == nil {
		mysqlStatus["healthy"] = true
	}
	result["mysql"] = mysqlStatus
	ok(c, result)
}

func (a *API) serveFrontend(r *gin.Engine) {
	dist, err := filepath.Abs(a.cfg.WebDist)
	if err != nil {
		return
	}
	if info, err := os.Stat(dist); err != nil || !info.IsDir() {
		return
	}
	assets := filepath.Join(dist, "assets")
	if _, err := os.Stat(assets); err == nil {
		r.Static("/assets", assets)
	}
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/agent/") {
			fail(c, 404, "NOT_FOUND", "接口不存在")
			return
		}
		c.File(filepath.Join(dist, "index.html"))
	})
}

func (a *API) audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" || len(requestID) > 128 {
			requestID = uuid.NewString()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions || c.Writer.Status() >= 400 {
			return
		}
		user := currentUser(c)
		if user.ID == "" {
			return
		}
		detail := map[string]any{"request_id": requestID, "status": c.Writer.Status()}
		if value, exists := c.Get("audit_detail"); exists {
			detail["change"] = value
		}
		encoded, _ := json.Marshal(detail)
		log := store.AuditLog{Base: store.NewBase(), UserID: user.ID, Username: user.Username, Action: c.Request.Method, Object: c.FullPath(), IP: c.ClientIP(), Detail: datatypes.JSON(encoded)}
		_ = a.db.Create(&log).Error
	}
}

func setAuditChange(c *gin.Context, objectType, objectID string, before, after any) {
	c.Set("audit_detail", map[string]any{"object_type": objectType, "object_id": objectID, "before": before, "after": after})
}

func auditUserView(item store.User) map[string]any {
	return map[string]any{"username": item.Username, "display_name": item.DisplayName, "role": item.Role, "enabled": item.Enabled}
}

func (a *API) removingLastEnabledAdmin(item store.User, nextRole string, nextEnabled bool) bool {
	if item.Role != "admin" || !item.Enabled || (nextRole == "admin" && nextEnabled) {
		return false
	}
	var others int64
	a.db.Model(&store.User{}).Where("id <> ? AND role = ? AND enabled = ?", item.ID, "admin", true).Count(&others)
	return others == 0
}

func (a *API) login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8<<10)
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "请输入用户名和密码")
		return
	}
	clientIP := c.ClientIP()
	if !a.loginLimiter.allow(clientIP) {
		fail(c, http.StatusTooManyRequests, "LOGIN_RATE_LIMITED", "登录尝试次数过多，请稍后重试")
		return
	}
	username := strings.TrimSpace(req.Username)
	var user store.User
	lookupErr := a.db.Where("username = ?", username).First(&user).Error
	now := time.Now()
	if lookupErr == nil && user.Enabled && user.LockedUntil != nil && now.Before(*user.LockedUntil) {
		a.loginLimiter.failure(clientIP)
		time.Sleep(250 * time.Millisecond)
		fail(c, http.StatusTooManyRequests, "LOGIN_LOCKED", "登录尝试次数过多，请稍后重试")
		return
	}
	passwordHash := dummyPasswordHash
	if lookupErr == nil {
		passwordHash = []byte(user.PasswordHash)
	}
	passwordMatches := bcrypt.CompareHashAndPassword(passwordHash, []byte(req.Password)) == nil
	passwordValid := lookupErr == nil && user.Enabled && passwordMatches
	if !passwordValid {
		accountLocked := false
		if lookupErr == nil && user.Enabled {
			accountLocked = a.recordFailedLogin(user.ID, now)
		}
		ipLocked := a.loginLimiter.failure(clientIP)
		time.Sleep(250 * time.Millisecond)
		if accountLocked || ipLocked {
			fail(c, http.StatusTooManyRequests, "LOGIN_LOCKED", "登录尝试次数过多，请稍后重试")
			return
		}
		fail(c, 401, "LOGIN_FAILED", "用户名或密码错误")
		return
	}
	if err := a.db.Model(&user).Updates(map[string]any{"last_login_at": now, "failed_login_count": 0, "last_failed_login_at": nil, "locked_until": nil}).Error; err != nil {
		fail(c, 500, "LOGIN_STATE_ERROR", "登录失败")
		return
	}
	a.loginLimiter.success(clientIP)
	authUser := AuthUser{ID: user.ID, Username: user.Username, Role: user.Role, TokenVersion: user.TokenVersion}
	token, expiresAt, err := a.tokens.Issue(authUser)
	if err != nil {
		fail(c, 500, "TOKEN_ERROR", "登录失败")
		return
	}
	ok(c, gin.H{"token": token, "expires_at": expiresAt, "user": authUser, "display_name": user.DisplayName})
}

var dummyPasswordHash = func() []byte {
	hash, _ := bcrypt.GenerateFromPassword([]byte("honeynet-invalid-login-placeholder"), 12)
	return hash
}()

func (a *API) recordFailedLogin(userID string, now time.Time) bool {
	locked := false
	_ = a.db.Transaction(func(tx *gorm.DB) error {
		var user store.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ?", userID).Error; err != nil {
			return err
		}
		failures := user.FailedLoginCount
		if user.LastFailedLoginAt == nil || now.Sub(*user.LastFailedLoginAt) >= loginFailureWindow {
			failures = 0
		}
		if user.LockedUntil != nil {
			if now.Before(*user.LockedUntil) {
				locked = true
				return nil
			}
			failures = 0
		}
		failures++
		updates := map[string]any{"failed_login_count": failures, "last_failed_login_at": now, "locked_until": nil}
		if failures >= accountFailureLimit {
			until := now.Add(accountLockDuration)
			updates["locked_until"] = until
			locked = true
		}
		return tx.Model(&user).Updates(updates).Error
	})
	return locked
}

func (a *API) logout(c *gin.Context) {
	user := currentUser(c)
	result := a.db.Model(&store.User{}).Where("id = ?", user.ID).UpdateColumn("token_version", gorm.Expr("token_version + 1"))
	if result.Error != nil || result.RowsAffected != 1 {
		fail(c, 500, "LOGOUT_FAILED", "退出登录失败")
		return
	}
	a.hub.RevokeUser(user.ID)
	okEmpty(c)
}

func (a *API) profile(c *gin.Context) {
	var user store.User
	if a.db.First(&user, "id = ?", currentUser(c).ID).Error != nil {
		fail(c, 404, "USER_NOT_FOUND", "用户不存在")
		return
	}
	ok(c, user)
}

func randomToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}
func (a *API) issueNodeToken(node *store.Node) (string, error) {
	token, hash, err := randomToken()
	if err != nil {
		return "", err
	}
	expires := time.Now().Add(30 * time.Minute)
	needsEnrollment := false
	err = a.db.Transaction(func(tx *gorm.DB) error {
		var current store.Node
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, "id = ?", node.ID).Error; err != nil {
			return err
		}
		now := time.Now()
		hasCurrentCertificate := current.CertificateSerial != "" && current.CertificateExpiresAt != nil && current.CertificateExpiresAt.After(now)
		updates := map[string]any{
			"registration_token_hash": hash,
			"agent_token_hash":        "",
			"token_expires_at":        expires,
		}
		if !hasCurrentCertificate {
			needsEnrollment = true
			updates["certificate_serial"] = ""
			updates["certificate_issued_at"] = nil
			updates["certificate_expires_at"] = nil
			updates["pending_certificate_serial"] = ""
			updates["pending_certificate_not_after"] = nil
			updates["pending_certificate_activation_at"] = nil
			updates["status"] = "registering"
		}
		if err := tx.Model(&current).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.First(&current, "id = ?", current.ID).Error; err != nil {
			return err
		}
		*node = current
		return nil
	})
	if err != nil {
		return "", err
	}
	// Merely viewing a fresh installer or rotating the one-time token must not
	// revoke a healthy Agent. Only a node without a usable certificate is in a
	// true enrollment state.
	if needsEnrollment && a.agents != nil {
		a.agents.Revoke(node.ID)
	}
	return token, nil
}

func (a *API) listNodes(c *gin.Context) {
	p, s := page(c)
	q := a.db.Model(&store.Node{})
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		like := store.LikePattern(keyword)
		q = q.Where("name LIKE ? OR ip LIKE ?", like, like)
	}
	var total int64
	q.Count(&total)
	var items []store.Node
	q.Preload("Sense").Order("created_at DESC").Offset((p - 1) * s).Limit(s).Find(&items)
	ok(c, pageResult(items, total, p, s))
}
func (a *API) getNode(c *gin.Context) {
	var item store.Node
	if a.db.Preload("Sense").First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	var potCount, eventCount int64
	a.db.Model(&store.PotInstance{}).Where("node_id = ?", item.ID).Count(&potCount)
	if a.analytics != nil {
		if counts, err := a.analytics.CountByNodes(c.Request.Context(), []string{item.ID}); err == nil {
			eventCount = int64(counts[item.ID])
		}
	} else {
		a.db.Model(&store.AttackEvent{}).Where("node_id = ?", item.ID).Count(&eventCount)
	}
	ok(c, gin.H{"node": item, "pot_count": potCount, "event_count": eventCount})
}
func (a *API) createNode(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Group       string `json:"group"`
		IP          string `json:"ip"`
		AddressMode string `json:"address_mode"`
		OS          string `json:"os"`
		Arch        string `json:"arch"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "节点名称不能为空")
		return
	}
	addressMode := req.AddressMode
	if strings.TrimSpace(addressMode) == "" && strings.TrimSpace(req.IP) != "" {
		addressMode = nodeAddressCustom
	}
	node := store.Node{Base: store.NewBase(), Name: req.Name, GroupName: req.Group, OS: req.OS, Arch: req.Arch, Status: "offline", Labels: datatypes.JSON("{}"), PublicIPs: encodeNodeIPs(nil), PrivateIPs: encodeNodeIPs(nil)}
	if err := configureNodeAddress(&node, addressMode, req.IP); err != nil {
		fail(c, 400, "INVALID_NODE_ADDRESS", err.Error())
		return
	}
	if a.db.Create(&node).Error != nil {
		fail(c, 500, "CREATE_FAILED", "创建节点失败")
		return
	}
	senseConfig := store.NewNodeSenseConfig(node.ID)
	if a.db.Create(&senseConfig).Error != nil {
		fail(c, 500, "CREATE_FAILED", "初始化节点感知配置失败")
		return
	}
	node.Sense = &senseConfig
	token, err := a.issueNodeToken(&node)
	if err != nil {
		fail(c, 500, "TOKEN_ERROR", "签发节点令牌失败")
		return
	}
	created(c, gin.H{"node": node, "registration_token": token, "expires_at": node.TokenExpiresAt, "agent_url": a.cfg.AgentPublicURL, "ca_sha256": a.pki.CAFingerprint(), "mtls": true, "install_command": a.installCommand(node.ID, token), "install_commands": a.installCommands(node.ID, token)})
}
func (a *API) updateNode(c *gin.Context) {
	var item store.Node
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	var req struct {
		Name        *string         `json:"name"`
		Group       *string         `json:"group"`
		IP          *string         `json:"ip"`
		AddressMode *string         `json:"address_mode"`
		Labels      json.RawMessage `json:"labels"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "请求格式错误")
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			fail(c, 400, "INVALID_ARGUMENT", "节点名称不能为空")
			return
		}
		updates["name"] = name
	}
	if req.Group != nil {
		updates["group_name"] = strings.TrimSpace(*req.Group)
	}
	if req.AddressMode != nil || req.IP != nil {
		mode, selected := item.AddressMode, item.IP
		if req.AddressMode != nil {
			mode = *req.AddressMode
		}
		if req.IP != nil {
			selected = *req.IP
		}
		if err := configureNodeAddress(&item, mode, selected); err != nil {
			fail(c, 400, "INVALID_NODE_ADDRESS", err.Error())
			return
		}
		updates["address_mode"] = item.AddressMode
		updates["ip"] = item.IP
		updates["public_ip"] = item.PublicIP
		updates["public_ips"] = item.PublicIPs
		updates["private_ips"] = item.PrivateIPs
	}
	if len(req.Labels) > 0 {
		updates["labels"] = req.Labels
	}
	if len(updates) > 0 && a.db.Model(&item).Updates(updates).Error != nil {
		fail(c, 500, "UPDATE_FAILED", "更新节点失败")
		return
	}
	a.db.First(&item, "id = ?", item.ID)
	ok(c, item)
}
func (a *API) deleteNode(c *gin.Context) {
	var item store.Node
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	var count int64
	a.db.Model(&store.PotInstance{}).Where("node_id = ?", item.ID).Count(&count)
	if count > 0 {
		fail(c, 409, "NODE_IN_USE", "请先删除该节点上的蜜罐实例")
		return
	}
	a.db.Model(&store.Decoy{}).Where("node_id = ?", item.ID).Count(&count)
	if count > 0 {
		fail(c, 409, "NODE_IN_USE", "请先删除该节点上的蜜饵")
		return
	}
	a.db.Model(&item).Updates(map[string]any{"status": "revoked", "registration_token_hash": "", "agent_token_hash": ""})
	a.agents.Revoke(item.ID)
	a.db.Delete(&store.NodeSenseConfig{}, "node_id = ?", item.ID)
	a.db.Delete(&item)
	okEmpty(c)
}
func (a *API) rotateNodeToken(c *gin.Context) {
	var item store.Node
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	token, err := a.issueNodeToken(&item)
	if err != nil {
		fail(c, 500, "TOKEN_ERROR", "签发节点令牌失败")
		return
	}
	ok(c, gin.H{"node": item, "registration_token": token, "expires_at": item.TokenExpiresAt, "agent_url": a.cfg.AgentPublicURL, "ca_sha256": a.pki.CAFingerprint(), "mtls": true, "install_command": a.installCommand(item.ID, token), "install_commands": a.installCommands(item.ID, token)})
}
func (a *API) nodeInstall(c *gin.Context) {
	var item store.Node
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	token, err := a.issueNodeToken(&item)
	if err != nil {
		fail(c, 500, "TOKEN_ERROR", "签发节点令牌失败")
		return
	}
	ok(c, gin.H{"node": item, "registration_token": token, "agent_url": a.cfg.AgentPublicURL, "ca_sha256": a.pki.CAFingerprint(), "mtls": true, "install_commands": a.installCommands(item.ID, token), "expires_at": item.TokenExpiresAt})
}
func (a *API) installCommand(nodeID, token string) string {
	arguments := fmt.Sprintf("--server %s --agent-url %s --ca-sha256 %s --node-id %s --token %s",
		shellQuote(a.cfg.PublicURL), shellQuote(a.cfg.AgentPublicURL), shellQuote(a.pki.CAFingerprint()), shellQuote(nodeID), shellQuote(token))
	if a.cfg.TLSEnabled && !a.cfg.UsesExternalConsoleCertificate() {
		caBase64 := base64.StdEncoding.EncodeToString(a.pki.CACertificatePEM())
		return fmt.Sprintf("umask 077; CA=$(mktemp /tmp/honeynet-console-ca.XXXXXX); INSTALLER=$(mktemp /tmp/honeynet-install.XXXXXX); cleanup(){ rm -f \"$CA\" \"$INSTALLER\"; }; trap cleanup EXIT HUP INT TERM; printf '%%s' %s | base64 -d >\"$CA\"; curl -fsSL --cacert \"$CA\" %s/download/install-agent.sh -o \"$INSTALLER\"; sudo sh \"$INSTALLER\" --console-ca \"$CA\" %s",
			shellQuote(caBase64), shellQuote(a.cfg.PublicURL), arguments)
	}
	return fmt.Sprintf("umask 077; INSTALLER=$(mktemp /tmp/honeynet-install.XXXXXX); trap 'rm -f \"$INSTALLER\"' EXIT HUP INT TERM; curl -fsSL %s/download/install-agent.sh -o \"$INSTALLER\"; sudo sh \"$INSTALLER\" %s", shellQuote(a.cfg.PublicURL), arguments)
}

func (a *API) installCommands(nodeID, token string) gin.H {
	windowsCommand := fmt.Sprintf("$script=Join-Path $env:TEMP ('honeynet-install-'+[Guid]::NewGuid().ToString('N')+'.ps1'); try { Invoke-WebRequest -UseBasicParsing %s -OutFile $script; . $script; Install-HoneynetAgent -Server %s -AgentURL %s -CASHA256 %s -NodeID %s -Token %s } finally { Remove-Item -Force -ErrorAction SilentlyContinue $script }",
		powerShellQuote(a.cfg.PublicURL+"/download/install-agent.ps1"), powerShellQuote(a.cfg.PublicURL), powerShellQuote(a.cfg.AgentPublicURL), powerShellQuote(a.pki.CAFingerprint()), powerShellQuote(nodeID), powerShellQuote(token))
	if a.cfg.TLSEnabled && !a.cfg.UsesExternalConsoleCertificate() {
		caBase64 := base64.StdEncoding.EncodeToString(a.pki.CACertificateDER())
		windowsCommand = fmt.Sprintf("$caPath=Join-Path $env:TEMP ('honeynet-ca-'+[Guid]::NewGuid().ToString('N')+'.cer'); $script=Join-Path $env:TEMP ('honeynet-install-'+[Guid]::NewGuid().ToString('N')+'.ps1'); $imported=$false; try { [IO.File]::WriteAllBytes($caPath,[Convert]::FromBase64String(%s)); $ca=New-Object Security.Cryptography.X509Certificates.X509Certificate2($caPath); if (-not (Get-ChildItem Cert:\\CurrentUser\\Root | Where-Object Thumbprint -eq $ca.Thumbprint)) { Import-Certificate -FilePath $caPath -CertStoreLocation Cert:\\CurrentUser\\Root | Out-Null; $imported=$true }; Invoke-WebRequest -UseBasicParsing %s -OutFile $script; . $script; Install-HoneynetAgent -Server %s -AgentURL %s -CASHA256 %s -NodeID %s -Token %s } finally { if ($imported -and $ca) { Remove-Item -Force ('Cert:\\CurrentUser\\Root\\'+$ca.Thumbprint) }; Remove-Item -Force -ErrorAction SilentlyContinue $caPath,$script }",
			powerShellQuote(caBase64), powerShellQuote(a.cfg.PublicURL+"/download/install-agent.ps1"), powerShellQuote(a.cfg.PublicURL), powerShellQuote(a.cfg.AgentPublicURL), powerShellQuote(a.pki.CAFingerprint()), powerShellQuote(nodeID), powerShellQuote(token))
	}
	return gin.H{
		"linux":               a.installCommand(nodeID, token),
		"windows":             windowsCommand,
		"available":           true,
		"console_tls_enabled": a.cfg.TLSEnabled,
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func powerShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func (a *API) listPotServices(c *gin.Context) {
	p, s := page(c)
	currentCodes := store.CurrentPotServiceCodes()
	includeRetired := c.Query("include_retired") == "1" || strings.EqualFold(c.Query("include_retired"), "true")
	q := a.db.Model(&store.PotService{})
	if !includeRetired {
		q = q.Where("code IN ?", currentCodes)
	}
	if category := c.Query("category"); category != "" {
		q = q.Where("category = ?", category)
	}
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		like := store.LikePattern(keyword)
		q = q.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	var total int64
	q.Count(&total)
	var items []store.PotService
	q.Order("category, default_port").Offset((p - 1) * s).Limit(s).Find(&items)
	var capabilityNode *store.Node
	if nodeID := strings.TrimSpace(c.Query("node_id")); nodeID != "" {
		var node store.Node
		if a.db.First(&node, "id = ?", nodeID).Error != nil {
			fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
			return
		}
		capabilityNode = &node
	}
	type serviceView struct {
		store.PotService
		Capability    string `json:"capability"`
		SupportStatus string `json:"support_status"`
	}
	views := make([]serviceView, 0, len(items))
	for _, item := range items {
		status := "not_evaluated"
		if capabilityNode != nil {
			known, supported := nodeSupportsService(*capabilityNode, item.Code)
			switch {
			case !known:
				status = "unknown"
			case supported:
				status = "supported"
			default:
				status = "unsupported"
			}
		}
		views = append(views, serviceView{PotService: item, Capability: "pot." + item.Code, SupportStatus: status})
	}
	var categories []string
	categoryQuery := a.db.Model(&store.PotService{})
	if !includeRetired {
		categoryQuery = categoryQuery.Where("code IN ?", currentCodes)
	}
	categoryQuery.Distinct().Order("category").Pluck("category", &categories)
	ok(c, gin.H{"items": views, "total": total, "page": p, "page_size": s, "categories": categories})
}

func requireNodeServiceCapability(c *gin.Context, node store.Node, serviceCode string) bool {
	known, supported := nodeSupportsService(node, serviceCode)
	if !known {
		fail(c, 409, "NODE_CAPABILITIES_UNKNOWN", "节点尚未上报服务能力，请先安装或升级 Agent 并等待其上线")
		return false
	}
	if !supported {
		fail(c, 409, "SERVICE_UNSUPPORTED", "当前节点 Agent 不支持该蜜罐服务")
		return false
	}
	return true
}

func (a *API) listPots(c *gin.Context) {
	p, s := page(c)
	q := a.db.Model(&store.PotInstance{})
	if id := c.Query("node_id"); id != "" {
		q = q.Where("node_id = ?", id)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("desired_status = ?", status)
	}
	var total int64
	q.Count(&total)
	var items []store.PotInstance
	q.Preload("Node").Preload("Service").Preload("Template").Order("created_at DESC").Offset((p - 1) * s).Limit(s).Find(&items)
	ok(c, pageResult(items, total, p, s))
}

func (a *API) getPot(c *gin.Context) {
	var item store.PotInstance
	if err := a.db.Preload("Node").Preload("Service").Preload("Template").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "POT_NOT_FOUND", "蜜罐不存在")
			return
		}
		fail(c, http.StatusInternalServerError, "QUERY_FAILED", "查询蜜罐失败")
		return
	}
	ok(c, item)
}

type potRequest struct {
	NodeID        string          `json:"node_id" binding:"required"`
	ServiceCode   string          `json:"service_code" binding:"required"`
	Name          string          `json:"name" binding:"required"`
	Port          int             `json:"port"`
	Config        json.RawMessage `json:"config"`
	TemplateID    string          `json:"template_id"`
	DesiredStatus string          `json:"desired_status"`
}

func (a *API) createPot(c *gin.Context) {
	var req potRequest
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "节点、服务和名称不能为空")
		return
	}
	if !store.IsCurrentPotService(req.ServiceCode) {
		fail(c, 404, "SERVICE_NOT_FOUND", "蜜罐服务不存在")
		return
	}
	var svc store.PotService
	if a.db.First(&svc, "code = ?", req.ServiceCode).Error != nil {
		fail(c, 404, "SERVICE_NOT_FOUND", "蜜罐服务不存在")
		return
	}
	var node store.Node
	if a.db.First(&node, "id = ?", req.NodeID).Error != nil {
		fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	if !requireNodeServiceCapability(c, node, svc.Code) {
		return
	}
	if req.Port == 0 {
		req.Port = svc.DefaultPort
	}
	if req.Port < 1 || req.Port > 65535 {
		fail(c, 400, "INVALID_PORT", "蜜罐端口必须在 1-65535 范围内")
		return
	}
	var templateID *string
	if svc.Code == "web-template" {
		req.TemplateID = strings.TrimSpace(req.TemplateID)
		if req.TemplateID == "" {
			fail(c, 400, "TEMPLATE_REQUIRED", "自定义 Web 蜜罐必须选择模板")
			return
		}
		var template store.PotTemplate
		if a.db.First(&template, "id = ?", req.TemplateID).Error != nil {
			fail(c, 404, "TEMPLATE_NOT_FOUND", "模板不存在")
			return
		}
		templateID = &template.ID
	} else if strings.TrimSpace(req.TemplateID) != "" {
		fail(c, 400, "TEMPLATE_NOT_SUPPORTED", "只有自定义 Web 蜜罐可以绑定模板")
		return
	}
	var exists int64
	a.db.Model(&store.PotInstance{}).Where("node_id = ? AND port = ?", req.NodeID, req.Port).Count(&exists)
	if exists > 0 {
		fail(c, 409, "PORT_CONFLICT", "该节点端口已被其他蜜罐占用")
		return
	}
	normalizedConfig, err := normalizePotConfig(req.Config)
	if err != nil {
		fail(c, 400, "INVALID_CONFIG", "蜜罐配置必须为有效对象，bind 只能填写 IPv4 或 IPv6 地址，且配置不能超过 32 KiB")
		return
	}
	req.Config = normalizedConfig
	desiredStatus := strings.ToLower(strings.TrimSpace(req.DesiredStatus))
	if desiredStatus == "" {
		desiredStatus = "stopped"
	}
	if desiredStatus != "running" && desiredStatus != "stopped" {
		fail(c, 400, "INVALID_STATUS", "蜜罐目标状态只能是 running 或 stopped")
		return
	}
	item := store.PotInstance{Base: store.NewBase(), NodeID: req.NodeID, ServiceCode: req.ServiceCode, TemplateID: templateID, Name: req.Name, Port: req.Port, Config: datatypes.JSON(req.Config), DesiredStatus: desiredStatus, ActualStatus: "pending"}
	if a.db.Create(&item).Error != nil {
		fail(c, 500, "CREATE_FAILED", "创建蜜罐失败")
		return
	}
	a.db.Preload("Node").Preload("Service").Preload("Template").First(&item, "id = ?", item.ID)
	_ = a.agents.SendPotApply(item.NodeID)
	created(c, item)
}
func (a *API) updatePot(c *gin.Context) {
	var item store.PotInstance
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "POT_NOT_FOUND", "蜜罐不存在")
		return
	}
	var req struct {
		Name       *string         `json:"name"`
		Port       *int            `json:"port"`
		Config     json.RawMessage `json:"config"`
		TemplateID *string         `json:"template_id"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "请求格式错误")
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			fail(c, 400, "INVALID_NAME", "蜜罐名称不能为空")
			return
		}
		updates["name"] = name
	}
	if req.Port != nil {
		if *req.Port < 1 || *req.Port > 65535 {
			fail(c, 400, "INVALID_PORT", "蜜罐端口必须在 1-65535 范围内")
			return
		}
		var count int64
		a.db.Model(&store.PotInstance{}).Where("node_id = ? AND port = ? AND id <> ?", item.NodeID, *req.Port, item.ID).Count(&count)
		if count > 0 {
			fail(c, 409, "PORT_CONFLICT", "该节点端口已被其他蜜罐占用")
			return
		}
		updates["port"] = *req.Port
	}
	if len(req.Config) > 0 {
		normalizedConfig, err := normalizePotConfig(req.Config)
		if err != nil {
			fail(c, 400, "INVALID_CONFIG", "蜜罐配置必须为有效对象，bind 只能填写 IPv4 或 IPv6 地址，且配置不能超过 32 KiB")
			return
		}
		updates["config"] = normalizedConfig
	}
	if req.TemplateID != nil {
		if item.ServiceCode != "web-template" {
			fail(c, 400, "TEMPLATE_NOT_SUPPORTED", "只有自定义 Web 蜜罐可以绑定模板")
			return
		}
		templateID := strings.TrimSpace(*req.TemplateID)
		if templateID == "" {
			fail(c, 400, "TEMPLATE_REQUIRED", "自定义 Web 蜜罐必须选择模板")
			return
		}
		var template store.PotTemplate
		if a.db.First(&template, "id = ?", templateID).Error != nil {
			fail(c, 404, "TEMPLATE_NOT_FOUND", "模板不存在")
			return
		}
		updates["template_id"] = template.ID
	}
	if len(updates) > 0 {
		updates["actual_status"] = "pending"
		if err := a.db.Model(&item).Updates(updates).Error; err != nil {
			fail(c, 500, "UPDATE_FAILED", "更新蜜罐失败")
			return
		}
		_ = a.agents.SendPotApply(item.NodeID)
	}
	if err := a.db.Preload("Node").Preload("Service").Preload("Template").First(&item, "id = ?", item.ID).Error; err != nil {
		fail(c, 500, "QUERY_FAILED", "查询蜜罐失败")
		return
	}
	ok(c, item)
}
func (a *API) deletePot(c *gin.Context) {
	var item store.PotInstance
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "POT_NOT_FOUND", "蜜罐不存在")
		return
	}
	if a.db.Delete(&item).Error != nil {
		fail(c, 500, "DELETE_FAILED", "删除蜜罐失败")
		return
	}
	_ = a.agents.SendPotApply(item.NodeID)
	okEmpty(c)
}
func (a *API) potAction(status string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var item store.PotInstance
		if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
			fail(c, 404, "POT_NOT_FOUND", "蜜罐不存在")
			return
		}
		if status == "running" {
			if !store.IsCurrentPotService(item.ServiceCode) {
				fail(c, 409, "SERVICE_RETIRED", "该蜜罐服务已退出当前版本，请删除旧实例后选择受支持服务")
				return
			}
			var node store.Node
			if a.db.First(&node, "id = ?", item.NodeID).Error != nil {
				fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
				return
			}
			if !requireNodeServiceCapability(c, node, item.ServiceCode) {
				return
			}
		}
		a.db.Model(&item).Updates(map[string]any{"desired_status": status, "actual_status": "pending"})
		item.DesiredStatus = status
		item.ActualStatus = "pending"
		a.hub.Broadcast("pot.desired", item)
		_ = a.agents.SendPotApply(item.NodeID)
		ok(c, item)
	}
}

func (a *API) listEvents(c *gin.Context) {
	disclosure, permitted := a.eventDisclosureForRequest(c)
	if !permitted {
		fail(c, http.StatusForbidden, "SENSITIVE_EVENT_FORBIDDEN", "当前账号无权查看事件敏感原文")
		return
	}
	pagination, err := parseEventPagination(c.Request.URL.Query())
	if err != nil {
		if errors.Is(err, errEventPageTooDeep) {
			fail(c, http.StatusBadRequest, "PAGE_TOO_DEEP", "攻击事件最多支持查询前 100 页，请缩小时间范围或增加筛选条件")
		} else {
			fail(c, http.StatusBadRequest, "INVALID_PAGINATION", "攻击事件分页参数无效或相互冲突")
		}
		return
	}
	p, s := pagination.Page, pagination.PageSize
	from, to, err := parseEventTimeRange(c.Query("from"), c.Query("to"))
	if err != nil {
		fail(c, 400, "INVALID_TIME_RANGE", "事件时间范围必须使用 RFC3339 格式，且开始时间不能晚于结束时间")
		return
	}
	if a.analytics != nil {
		result, err := a.analytics.List(c.Request.Context(), analytics.EventFilter{
			NodeID: c.Query("node_id"), PotID: c.Query("pot_id"), Service: c.Query("service"),
			EventType: c.Query("event_type"), EventClass: c.Query("event_class"), DecoyID: c.Query("decoy_id"),
			SourceIP: c.Query("src_ip"), From: from, To: to, Cursor: pagination.Cursor, CursorMode: pagination.CursorMode,
			SkipTotal: !pagination.IncludeTotal, Limit: s, Offset: (p - 1) * s,
		})
		if err != nil {
			fail(c, 503, "ANALYTICS_UNAVAILABLE", "安全事件分析服务暂不可用")
			return
		}
		items := make([]store.AttackEvent, 0, len(result.Items))
		for _, event := range result.Items {
			item := analytics.ToAttackEvent(event)
			normalizeEventDetections(&item)
			items = append(items, item)
		}
		views := a.enrichEventThreatIntelligence(loadEventViewsWithDisclosure(a.db, items, disclosure))
		if disclosure.explicitReveal {
			if err := a.auditEventReveal(c, "", "list"); err != nil {
				fail(c, http.StatusServiceUnavailable, "AUDIT_UNAVAILABLE", "敏感原文查看审计暂不可用，请稍后重试")
				return
			}
		}
		if pagination.CursorMode {
			nextCursor, encodeErr := encodeEventCursor(result.NextCursor)
			if encodeErr != nil {
				fail(c, http.StatusServiceUnavailable, "EVENT_PAGINATION_FAILED", "安全事件分页状态生成失败")
				return
			}
			ok(c, gin.H{"items": views, "total": nil, "total_known": false, "page_size": s, "pagination": "cursor", "next_cursor": nextCursor, "has_more": result.HasMore})
		} else {
			response := gin.H{"items": views, "total": nil, "total_known": false, "page": p, "page_size": s, "pagination": "page", "max_page": maxEventPage, "has_more": result.HasMore}
			if pagination.IncludeTotal {
				response["total"] = int64(result.Total)
				response["total_known"] = true
			}
			ok(c, response)
		}
		return
	}
	q := a.db.Model(&store.AttackEvent{})
	if value := c.Query("node_id"); value != "" {
		q = q.Where("node_id = ?", value)
	}
	if value := c.Query("pot_id"); value != "" {
		q = q.Where("pot_id = ?", value)
	}
	if value := c.Query("service"); value != "" {
		q = q.Where("service = ?", value)
	}
	if value := c.Query("event_type"); value != "" {
		q = q.Where("event_type = ?", value)
	}
	if value := c.Query("event_class"); value != "" {
		switch value {
		case "credential":
			q = q.Where("event_type LIKE ?", "%.credential")
		case "web":
			q = q.Where("event_type LIKE ?", "web.%")
		case "decoy":
			q = q.Where("event_type LIKE ?", "decoy.%")
		}
	}
	if value := c.Query("decoy_id"); value != "" {
		q = q.Where("decoy_id = ?", value)
	}
	if value := c.Query("src_ip"); value != "" {
		q = q.Where("src_ip LIKE ?", store.LikePattern(value))
	}
	if !from.IsZero() {
		q = q.Where("ts >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("ts <= ?", to)
	}
	if pagination.CursorMode {
		if pagination.Cursor != nil {
			q = q.Where("ts < ? OR (ts = ? AND event_id < ?)", pagination.Cursor.EventTime, pagination.Cursor.EventTime, pagination.Cursor.EventID)
		}
		var items []store.AttackEvent
		if err := q.Order("ts DESC, event_id DESC").Limit(s + 1).Find(&items).Error; err != nil {
			fail(c, http.StatusServiceUnavailable, "EVENT_QUERY_FAILED", "安全事件查询暂不可用")
			return
		}
		hasMore := len(items) > s
		if hasMore {
			items = items[:s]
		}
		for index := range items {
			normalizeEventDetections(&items[index])
		}
		views := a.enrichEventThreatIntelligence(loadEventViewsWithDisclosure(a.db, items, disclosure))
		if disclosure.explicitReveal {
			if err := a.auditEventReveal(c, "", "list"); err != nil {
				fail(c, http.StatusServiceUnavailable, "AUDIT_UNAVAILABLE", "敏感原文查看审计暂不可用，请稍后重试")
				return
			}
		}
		var next *analytics.Cursor
		if hasMore && len(items) > 0 {
			last := items[len(items)-1]
			next = &analytics.Cursor{EventTime: last.Timestamp, EventID: last.EventID}
		}
		nextCursor, encodeErr := encodeEventCursor(next)
		if encodeErr != nil {
			fail(c, http.StatusServiceUnavailable, "EVENT_PAGINATION_FAILED", "安全事件分页状态生成失败")
			return
		}
		ok(c, gin.H{"items": views, "total": nil, "total_known": false, "page_size": s, "pagination": "cursor", "next_cursor": nextCursor, "has_more": hasMore})
		return
	}
	var total int64
	if pagination.IncludeTotal {
		if err := q.Count(&total).Error; err != nil {
			fail(c, http.StatusServiceUnavailable, "EVENT_QUERY_FAILED", "安全事件总数查询暂不可用")
			return
		}
	}
	queryLimit := s
	if !pagination.IncludeTotal {
		queryLimit++
	}
	var items []store.AttackEvent
	if err := q.Order("ts DESC, event_id DESC").Offset((p - 1) * s).Limit(queryLimit).Find(&items).Error; err != nil {
		fail(c, http.StatusServiceUnavailable, "EVENT_QUERY_FAILED", "安全事件查询暂不可用")
		return
	}
	hasMore := int64(p*s) < total
	if !pagination.IncludeTotal {
		hasMore = len(items) > s
		if hasMore {
			items = items[:s]
		}
	}
	for index := range items {
		normalizeEventDetections(&items[index])
	}
	views := a.enrichEventThreatIntelligence(loadEventViewsWithDisclosure(a.db, items, disclosure))
	if disclosure.explicitReveal {
		if err := a.auditEventReveal(c, "", "list"); err != nil {
			fail(c, http.StatusServiceUnavailable, "AUDIT_UNAVAILABLE", "敏感原文查看审计暂不可用，请稍后重试")
			return
		}
	}
	response := gin.H{"items": views, "total": nil, "total_known": false, "page": p, "page_size": s, "pagination": "page", "max_page": maxEventPage, "has_more": hasMore}
	if pagination.IncludeTotal {
		response["total"] = total
		response["total_known"] = true
	}
	ok(c, response)
}
func (a *API) getEvent(c *gin.Context) {
	disclosure, permitted := a.eventDisclosureForRequest(c)
	if !permitted {
		fail(c, http.StatusForbidden, "SENSITIVE_EVENT_FORBIDDEN", "当前账号无权查看事件敏感原文")
		return
	}
	var item store.AttackEvent
	if a.analytics != nil {
		event, err := a.analytics.Get(c.Request.Context(), c.Param("id"))
		if err != nil {
			if errors.Is(err, analytics.ErrNotFound) {
				fail(c, 404, "EVENT_NOT_FOUND", "事件不存在")
			} else {
				fail(c, 503, "ANALYTICS_UNAVAILABLE", "安全事件分析服务暂不可用")
			}
			return
		}
		item = analytics.ToAttackEvent(event)
	} else if a.db.First(&item, "event_id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "EVENT_NOT_FOUND", "事件不存在")
		return
	}
	normalizeEventDetections(&item)
	views := a.enrichEventThreatIntelligence(loadEventViewsWithDisclosure(a.db, []store.AttackEvent{item}, disclosure))
	if len(views) == 0 {
		fail(c, 404, "EVENT_NOT_FOUND", "事件不存在")
		return
	}
	if disclosure.explicitReveal {
		if err := a.auditEventReveal(c, item.EventID, "event"); err != nil {
			fail(c, http.StatusServiceUnavailable, "AUDIT_UNAVAILABLE", "敏感原文查看审计暂不可用，请稍后重试")
			return
		}
	}
	ok(c, views[0])
}
func (a *API) listAlerts(c *gin.Context) {
	p, s := page(c)
	q := a.db.Model(&store.Alert{})
	if value := c.Query("status"); value != "" {
		q = q.Where("status = ?", value)
	}
	if value := c.Query("level"); value != "" {
		q = q.Where("level = ?", value)
	}
	var total int64
	q.Count(&total)
	var items []store.Alert
	q.Order("created_at DESC").Offset((p - 1) * s).Limit(s).Find(&items)
	ok(c, pageResult(items, total, p, s))
}
func (a *API) ackAlert(c *gin.Context) {
	var item store.Alert
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "ALERT_NOT_FOUND", "告警不存在")
		return
	}
	now := time.Now()
	user := currentUser(c)
	a.db.Model(&item).Updates(map[string]any{"status": "acked", "acked_by": user.ID, "acked_at": now})
	a.db.First(&item, "id = ?", item.ID)
	ok(c, item)
}

func (a *API) dashboardSummary(c *gin.Context) {
	var nodes, online, pots, running, events, newAlerts int64
	a.db.Model(&store.Node{}).Count(&nodes)
	a.db.Model(&store.Node{}).Where("status = ?", "online").Count(&online)
	a.db.Model(&store.PotInstance{}).Count(&pots)
	a.db.Model(&store.PotInstance{}).Where("actual_status = ?", "running").Count(&running)
	since := time.Now().Add(-24 * time.Hour)
	if a.analytics != nil {
		if stats, err := a.analytics.Dashboard(c.Request.Context(), since, time.Now()); err == nil {
			events = int64(stats.Events)
		}
	} else {
		a.db.Model(&store.AttackEvent{}).Where("ts >= ?", since).Count(&events)
	}
	a.db.Model(&store.Alert{}).Where("status = ?", "new").Count(&newAlerts)
	analyticsStatus := a.currentAnalyticsStatus(c.Request.Context())
	ok(c, gin.H{"version": a.cfg.Version, "server_healthy": true, "ipv6_capable": true, "nodes": nodes, "online_nodes": online, "pots": pots, "running_pots": running, "events_24h": events, "new_alerts": newAlerts, "geoip_enabled": a.geoIP != nil, "analytics": analyticsStatus})
}
func (a *API) dashboardTrends(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days < 1 || days > 30 {
		days = 7
	}
	if a.analytics != nil {
		rows, err := a.analytics.Trends(c.Request.Context(), time.Now().AddDate(0, 0, -days+1).Truncate(24*time.Hour), time.Now())
		if err != nil {
			fail(c, 503, "ANALYTICS_UNAVAILABLE", "安全分析服务的趋势查询暂不可用")
			return
		}
		result := make([]gin.H, 0, len(rows))
		for _, row := range rows {
			result = append(result, gin.H{"day": row.Day, "count": row.Count})
		}
		ok(c, result)
		return
	}
	type row struct {
		Day   string `json:"day"`
		Count int64  `json:"count"`
	}
	var rows []row
	a.db.Model(&store.AttackEvent{}).Select("DATE(ts) AS day, COUNT(*) AS count").Where("ts >= ?", time.Now().AddDate(0, 0, -days+1).Truncate(24*time.Hour)).Group("DATE(ts)").Order("day").Scan(&rows)
	ok(c, rows)
}
func (a *API) dashboardTopAttackers(c *gin.Context) {
	if a.analytics != nil {
		rows, err := a.analytics.TopAttackers(c.Request.Context(), time.Now().Add(-7*24*time.Hour), time.Now(), 10)
		if err != nil {
			fail(c, 503, "ANALYTICS_UNAVAILABLE", "安全分析服务的攻击来源查询暂不可用")
			return
		}
		ok(c, rows)
		return
	}
	type row struct {
		SrcIP    string    `json:"src_ip"`
		Geo      string    `json:"geo"`
		Count    int64     `json:"count"`
		LastSeen time.Time `json:"last_seen"`
	}
	var rows []row
	a.db.Model(&store.AttackEvent{}).Select("src_ip, MAX(geo) AS geo, COUNT(*) AS count, MAX(ts) AS last_seen").Where("ts >= ? AND src_ip <> ''", time.Now().Add(-7*24*time.Hour)).Group("src_ip").Order("count DESC").Limit(10).Scan(&rows)
	ok(c, rows)
}

func (a *API) listAlertRules(c *gin.Context) {
	var items []store.AlertRule
	a.db.Order("created_at DESC").Find(&items)
	ok(c, items)
}

type alertRuleRequest struct {
	Name          *string   `json:"name"`
	Enabled       *bool     `json:"enabled"`
	EventType     *string   `json:"event_type"`
	Level         *string   `json:"level"`
	Threshold     *int      `json:"threshold"`
	WindowMinute  *int      `json:"window_minute"`
	SilenceMinute *int      `json:"silence_minute"`
	ChannelIDs    *[]string `json:"channel_ids"`
}

func (a *API) createAlertRule(c *gin.Context) {
	var req alertRuleRequest
	if c.ShouldBindJSON(&req) != nil || req.Name == nil || req.EventType == nil {
		fail(c, 400, "INVALID_ARGUMENT", "规则名称和事件匹配不能为空")
		return
	}
	item := store.AlertRule{
		Base: store.NewBase(), Name: strings.TrimSpace(*req.Name), EventType: strings.TrimSpace(*req.EventType),
		Enabled: true, Level: "medium", Threshold: 1, WindowMinute: 1, ChannelIDs: datatypes.JSON("[]"),
	}
	applyAlertRuleRequest(&item, req)
	if err := a.validateAlertRule(item); err != nil {
		fail(c, 400, "INVALID_RULE", err.Error())
		return
	}
	if a.db.Create(&item).Error != nil {
		fail(c, 500, "CREATE_FAILED", "创建告警规则失败")
		return
	}
	created(c, item)
}
func (a *API) updateAlertRule(c *gin.Context) {
	var item store.AlertRule
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "RULE_NOT_FOUND", "规则不存在")
		return
	}
	var req alertRuleRequest
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "请求格式错误")
		return
	}
	applyAlertRuleRequest(&item, req)
	if err := a.validateAlertRule(item); err != nil {
		fail(c, 400, "INVALID_RULE", err.Error())
		return
	}
	updates := map[string]any{
		"name": item.Name, "enabled": item.Enabled, "event_type": item.EventType, "level": item.Level,
		"threshold": item.Threshold, "window_minute": item.WindowMinute, "silence_minute": item.SilenceMinute, "channel_ids": item.ChannelIDs,
	}
	if err := a.db.Model(&store.AlertRule{}).Where("id = ?", item.ID).Updates(updates).Error; err != nil {
		fail(c, 500, "UPDATE_FAILED", "更新告警规则失败")
		return
	}
	a.db.First(&item, "id = ?", item.ID)
	ok(c, item)
}

func applyAlertRuleRequest(item *store.AlertRule, req alertRuleRequest) {
	if req.Name != nil {
		item.Name = strings.TrimSpace(*req.Name)
	}
	if req.Enabled != nil {
		item.Enabled = *req.Enabled
	}
	if req.EventType != nil {
		item.EventType = strings.TrimSpace(*req.EventType)
	}
	if req.Level != nil {
		item.Level = strings.TrimSpace(*req.Level)
	}
	if req.Threshold != nil {
		item.Threshold = *req.Threshold
	}
	if req.WindowMinute != nil {
		item.WindowMinute = *req.WindowMinute
	}
	if req.SilenceMinute != nil {
		item.SilenceMinute = *req.SilenceMinute
	}
	if req.ChannelIDs != nil {
		data, _ := json.Marshal(*req.ChannelIDs)
		item.ChannelIDs = datatypes.JSON(data)
	}
}

func (a *API) validateAlertRule(item store.AlertRule) error {
	if item.Name == "" || item.EventType == "" {
		return errors.New("规则名称和事件匹配不能为空")
	}
	if _, err := path.Match(item.EventType, "validation.event"); err != nil {
		return errors.New("事件匹配表达式无效")
	}
	validLevel := item.Level == "info" || item.Level == "low" || item.Level == "medium" || item.Level == "high" || item.Level == "critical"
	if !validLevel {
		return errors.New("告警级别无效")
	}
	if item.Threshold < 1 || item.WindowMinute < 1 || item.SilenceMinute < 0 {
		return errors.New("阈值和时间窗必须大于零，静默时间不能为负数")
	}
	var channelIDs []string
	if json.Unmarshal(item.ChannelIDs, &channelIDs) != nil {
		return errors.New("投递通道格式无效")
	}
	if len(channelIDs) > 0 {
		unique := map[string]struct{}{}
		for _, id := range channelIDs {
			unique[id] = struct{}{}
		}
		var count int64
		a.db.Model(&store.AlertChannel{}).Where("id IN ?", channelIDs).Count(&count)
		if count != int64(len(unique)) {
			return errors.New("包含不存在的告警通道")
		}
	}
	return nil
}
func (a *API) listAlertChannels(c *gin.Context) {
	var items []store.AlertChannel
	a.db.Order("created_at DESC").Find(&items)
	views := make([]gin.H, 0, len(items))
	for _, item := range items {
		views = append(views, alertChannelView(item))
	}
	ok(c, views)
}
func (a *API) createAlertChannel(c *gin.Context) {
	var req struct {
		Name    string          `json:"name"`
		Type    string          `json:"type"`
		Enabled *bool           `json:"enabled"`
		Config  json.RawMessage `json:"config"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Type) == "" {
		fail(c, 400, "INVALID_ARGUMENT", "通道名称和类型不能为空")
		return
	}
	merged, err := alerting.MergeConfig(req.Type, nil, req.Config)
	if err != nil {
		fail(c, 400, "INVALID_CHANNEL_CONFIG", err.Error())
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	item := store.AlertChannel{Base: store.NewBase(), Name: strings.TrimSpace(req.Name), Type: strings.TrimSpace(req.Type), Enabled: enabled, Config: datatypes.JSON(merged)}
	if err := alerting.ValidateChannel(item); err != nil {
		fail(c, 400, "INVALID_CHANNEL_CONFIG", err.Error())
		return
	}
	if a.db.Create(&item).Error != nil {
		fail(c, 500, "CREATE_FAILED", "创建告警通道失败")
		return
	}
	created(c, alertChannelView(item))
}
func (a *API) updateAlertChannel(c *gin.Context) {
	var item store.AlertChannel
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "CHANNEL_NOT_FOUND", "通道不存在")
		return
	}
	var req struct {
		Name    *string         `json:"name"`
		Type    *string         `json:"type"`
		Enabled *bool           `json:"enabled"`
		Config  json.RawMessage `json:"config"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "请求格式错误")
		return
	}
	candidate := item
	if req.Name != nil {
		candidate.Name = strings.TrimSpace(*req.Name)
	}
	if req.Type != nil {
		candidate.Type = strings.TrimSpace(*req.Type)
	}
	if req.Enabled != nil {
		candidate.Enabled = *req.Enabled
	}
	if candidate.Name == "" || candidate.Type == "" {
		fail(c, 400, "INVALID_ARGUMENT", "通道名称和类型不能为空")
		return
	}
	if len(req.Config) > 0 {
		existing := json.RawMessage(item.Config)
		if candidate.Type != item.Type {
			existing = nil
		}
		merged, err := alerting.MergeConfig(candidate.Type, existing, req.Config)
		if err != nil {
			fail(c, 400, "INVALID_CHANNEL_CONFIG", err.Error())
			return
		}
		candidate.Config = datatypes.JSON(merged)
	}
	if err := alerting.ValidateChannel(candidate); err != nil {
		fail(c, 400, "INVALID_CHANNEL_CONFIG", err.Error())
		return
	}
	a.db.Model(&item).Updates(map[string]any{"name": candidate.Name, "type": candidate.Type, "enabled": candidate.Enabled, "config": candidate.Config})
	a.db.First(&item, "id = ?", item.ID)
	ok(c, alertChannelView(item))
}
func (a *API) deleteAlertChannel(c *gin.Context) {
	result := a.db.Delete(&store.AlertChannel{}, "id = ?", c.Param("id"))
	if result.RowsAffected == 0 {
		fail(c, 404, "CHANNEL_NOT_FOUND", "通道不存在")
		return
	}
	okEmpty(c)
}
func (a *API) testAlertChannel(c *gin.Context) {
	var item store.AlertChannel
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "CHANNEL_NOT_FOUND", "通道不存在")
		return
	}
	if err := a.alerts.Test(c.Request.Context(), item); err != nil {
		fail(c, 502, "CHANNEL_TEST_FAILED", "告警投递失败："+err.Error())
		return
	}
	ok(c, gin.H{"success": true, "message": "测试告警已成功投递"})
}
func (a *API) listAlertDeliveries(c *gin.Context) {
	p, s := page(c)
	query := a.db.Model(&store.AlertDelivery{})
	if value := c.Query("alert_id"); value != "" {
		query = query.Where("alert_id = ?", value)
	}
	if value := c.Query("status"); value != "" {
		query = query.Where("status = ?", value)
	}
	var total int64
	query.Count(&total)
	var items []store.AlertDelivery
	query.Order("created_at DESC").Offset((p - 1) * s).Limit(s).Find(&items)
	ok(c, pageResult(items, total, p, s))
}
func (a *API) retryAlertDelivery(c *gin.Context) {
	if err := a.alerts.Retry(c.Param("id")); err != nil {
		fail(c, 404, "DELIVERY_NOT_FOUND", "投递记录不存在")
		return
	}
	okEmpty(c)
}
func alertChannelView(item store.AlertChannel) gin.H {
	return gin.H{
		"id": item.ID, "name": item.Name, "type": item.Type, "enabled": item.Enabled,
		"config":     json.RawMessage(alerting.RedactedConfig(item.Type, json.RawMessage(item.Config))),
		"created_at": item.CreatedAt, "updated_at": item.UpdatedAt,
	}
}
func (a *API) listIOCs(c *gin.Context) {
	p, s := page(c)
	q := a.db.Model(&store.IOC{})
	if t := c.Query("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	var total int64
	q.Count(&total)
	var items []store.IOC
	q.Order("last_seen DESC").Offset((p - 1) * s).Limit(s).Find(&items)
	ok(c, pageResult(items, total, p, s))
}

func (a *API) listAssets(c *gin.Context) {
	type row struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		IP         string `json:"ip"`
		OS         string `json:"os"`
		Arch       string `json:"arch"`
		Status     string `json:"status"`
		PotCount   int64  `json:"pot_count"`
		OpenPorts  string `json:"open_ports"`
		EventCount int64  `json:"event_count"`
	}
	var rows []row
	a.db.Table("nodes n").Select("n.id,n.name,n.ip,n.os,n.arch,n.status,COUNT(DISTINCT p.id) AS pot_count,COALESCE(GROUP_CONCAT(DISTINCT p.port ORDER BY p.port), '') AS open_ports,0 AS event_count").Joins("LEFT JOIN pot_instances p ON p.node_id=n.id AND p.deleted_at IS NULL").Where("n.deleted_at IS NULL").Group("n.id,n.name,n.ip,n.os,n.arch,n.status").Order("n.created_at DESC").Scan(&rows)
	if a.analytics != nil && len(rows) > 0 {
		ids := make([]string, 0, len(rows))
		for _, item := range rows {
			ids = append(ids, item.ID)
		}
		if counts, err := a.analytics.CountByNodes(c.Request.Context(), ids); err == nil {
			for index := range rows {
				rows[index].EventCount = int64(counts[rows[index].ID])
			}
		}
	} else if a.analytics == nil {
		for index := range rows {
			a.db.Model(&store.AttackEvent{}).Where("node_id = ?", rows[index].ID).Count(&rows[index].EventCount)
		}
	}
	ok(c, rows)
}

func (a *API) listUsers(c *gin.Context) {
	var items []store.User
	a.db.Order("created_at DESC").Find(&items)
	ok(c, items)
}
func (a *API) createUser(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "用户名为 1-64 个字符，密码为 8-72 字节")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Username) > 64 || len(req.Password) < 8 || len(req.Password) > 72 {
		fail(c, 400, "INVALID_ARGUMENT", "用户名为 1-64 个字符，密码为 8-72 字节")
		return
	}
	if req.Role != "admin" && req.Role != "operator" && req.Role != "viewer" {
		fail(c, 400, "INVALID_ROLE", "角色无效")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		fail(c, 500, "PASSWORD_HASH_FAILED", "密码处理失败")
		return
	}
	item := store.User{Base: store.NewBase(), Username: req.Username, DisplayName: req.DisplayName, Role: req.Role, PasswordHash: string(hash), Enabled: true, TokenVersion: 1}
	if err := a.db.Create(&item).Error; err != nil {
		fail(c, 409, "USERNAME_EXISTS", "用户名已存在")
		return
	}
	setAuditChange(c, "user", item.ID, nil, auditUserView(item))
	created(c, item)
}
func (a *API) updateUser(c *gin.Context) {
	var item store.User
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "USER_NOT_FOUND", "用户不存在")
		return
	}
	var req struct {
		DisplayName *string `json:"display_name"`
		Role        *string `json:"role"`
		Enabled     *bool   `json:"enabled"`
		Password    string  `json:"password"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "请求格式错误")
		return
	}
	before := auditUserView(item)
	updates := map[string]any{}
	revokeTokens := false
	nextRole, nextEnabled := item.Role, item.Enabled
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Role != nil {
		if *req.Role != "admin" && *req.Role != "operator" && *req.Role != "viewer" {
			fail(c, 400, "INVALID_ROLE", "角色无效")
			return
		}
		updates["role"] = *req.Role
		nextRole = *req.Role
		revokeTokens = revokeTokens || *req.Role != item.Role
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
		nextEnabled = *req.Enabled
		revokeTokens = revokeTokens || *req.Enabled != item.Enabled
	}
	if req.Password != "" {
		if len(req.Password) < 8 || len(req.Password) > 72 {
			fail(c, 400, "INVALID_PASSWORD", "密码必须为 8-72 字节")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			fail(c, 500, "PASSWORD_HASH_FAILED", "密码处理失败")
			return
		}
		updates["password_hash"] = string(hash)
		revokeTokens = true
	}
	if a.removingLastEnabledAdmin(item, nextRole, nextEnabled) {
		fail(c, 409, "LAST_ADMIN_REQUIRED", "系统必须至少保留一个启用的管理员账号")
		return
	}
	if revokeTokens {
		updates["token_version"] = gorm.Expr("token_version + 1")
	}
	if err := a.db.Model(&item).Updates(updates).Error; err != nil {
		fail(c, 500, "UPDATE_FAILED", "更新用户失败")
		return
	}
	if revokeTokens {
		a.hub.RevokeUser(item.ID)
	}
	a.db.First(&item, "id = ?", item.ID)
	setAuditChange(c, "user", item.ID, before, auditUserView(item))
	ok(c, item)
}
func (a *API) deleteUser(c *gin.Context) {
	if c.Param("id") == currentUser(c).ID {
		fail(c, 409, "CANNOT_DELETE_SELF", "不能删除当前登录账号")
		return
	}
	var item store.User
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "USER_NOT_FOUND", "用户不存在")
		return
	}
	if a.removingLastEnabledAdmin(item, "", false) {
		fail(c, 409, "LAST_ADMIN_REQUIRED", "系统必须至少保留一个启用的管理员账号")
		return
	}
	result := a.db.Delete(&item)
	if result.RowsAffected == 0 {
		fail(c, 404, "USER_NOT_FOUND", "用户不存在")
		return
	}
	a.hub.RevokeUser(c.Param("id"))
	setAuditChange(c, "user", item.ID, auditUserView(item), nil)
	okEmpty(c)
}
func (a *API) listAuditLogs(c *gin.Context) {
	p, s := page(c)
	var total int64
	a.db.Model(&store.AuditLog{}).Count(&total)
	var items []store.AuditLog
	a.db.Order("created_at DESC").Offset((p - 1) * s).Limit(s).Find(&items)
	ok(c, pageResult(items, total, p, s))
}

func (a *API) registerAgent(c *gin.Context) {
	nodeID := c.GetHeader("X-Node-ID")
	token := c.GetHeader("X-Node-Token")
	if nodeID == "" || token == "" || c.Request.TLS == nil {
		fail(c, 401, "REGISTRATION_FAILED", "节点注册信息无效")
		return
	}
	var req struct {
		Version string   `json:"version"`
		IP      string   `json:"ip"`
		IPs     []string `json:"ips"`
		OS      string   `json:"os"`
		Arch    string   `json:"arch"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "节点信息格式错误")
		return
	}
	if len(req.Version) > 64 || len(req.OS) > 64 || len(req.Arch) > 64 || len(req.IPs) > 128 {
		fail(c, 400, "INVALID_ARGUMENT", "节点信息字段超过允许范围")
		return
	}
	reportedIPs := req.IPs
	if req.IP != "" {
		reportedIPs = append(append([]string(nil), reportedIPs...), req.IP)
	}
	observedIP := requestRemoteIP(c.Request)
	var node store.Node
	var enrollment nodepki.Enrollment
	err := a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&node, "id = ?", nodeID).Error; err != nil || node.Status == "revoked" {
			return gorm.ErrRecordNotFound
		}
		sum := sha256.Sum256([]byte(token))
		expected, _ := hex.DecodeString(node.RegistrationTokenHash)
		if len(expected) != len(sum) || subtle.ConstantTimeCompare(expected, sum[:]) != 1 || node.TokenExpiresAt == nil || node.TokenExpiresAt.Before(time.Now()) {
			return errInvalidRegistrationToken
		}
		issued, err := a.pki.IssueNode(node.ID, node.Name)
		if err != nil {
			return err
		}
		now := time.Now()
		hasCurrentCertificate := node.CertificateSerial != "" && node.CertificateExpiresAt != nil && node.CertificateExpiresAt.After(now)
		mergeNodeAddressReport(&node, observedIP, reportedIPs)
		updates := map[string]any{
			"registration_token_hash": "", "agent_token_hash": "", "token_expires_at": nil,
			"version": req.Version, "ip": node.IP, "address_mode": node.AddressMode, "public_ip": node.PublicIP, "public_ips": node.PublicIPs, "private_ips": node.PrivateIPs, "os": req.OS, "arch": req.Arch,
		}
		if hasCurrentCertificate {
			// A reinstall is a staged certificate rotation. Keep the currently
			// running Agent valid until the replacement proves possession of its
			// key on the mTLS gateway.
			updates["pending_certificate_serial"] = issued.Serial
			updates["pending_certificate_not_after"] = issued.ExpiresAt
			updates["pending_certificate_activation_at"] = now.Add(15 * time.Minute)
		} else {
			updates["certificate_serial"] = issued.Serial
			updates["certificate_issued_at"] = now
			updates["certificate_expires_at"] = issued.ExpiresAt
			updates["pending_certificate_serial"] = ""
			updates["pending_certificate_not_after"] = nil
			updates["pending_certificate_activation_at"] = nil
			updates["status"] = "registered"
		}
		if err := tx.Model(&node).Updates(updates).Error; err != nil {
			return err
		}
		enrollment = issued
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		fail(c, 401, "REGISTRATION_FAILED", "节点注册信息无效")
		return
	}
	if errors.Is(err, errInvalidRegistrationToken) {
		fail(c, 401, "REGISTRATION_FAILED", "注册令牌无效、已使用或已过期")
		return
	}
	if err != nil {
		fail(c, 500, "REGISTRATION_FAILED", "节点注册失败")
		return
	}
	// Existing nodes keep serving with their current certificate while the new
	// certificate is staged. Broadcast the persisted status instead of briefly
	// making an online node appear to fall back to "registered" in the console.
	if err := a.db.Select("status").First(&node, "id = ?", node.ID).Error; err != nil {
		fail(c, 500, "REGISTRATION_FAILED", "节点注册状态读取失败")
		return
	}
	a.hub.Broadcast("node.status", gin.H{"id": node.ID, "status": node.Status})
	ok(c, enrollmentResponse(node.ID, a.cfg.AgentPublicURL, a.cfg.AgentRenewBefore, a.updateSigner.PublicKey(), a.updateSigner.KeyID(), enrollment))
}

var errInvalidRegistrationToken = errors.New("invalid registration token")

func enrollmentResponse(nodeID, agentURL string, renewBefore time.Duration, updatePublicKey, updateKeyID string, enrollment nodepki.Enrollment) gin.H {
	return gin.H{
		"node_id": nodeID, "agent_url": agentURL,
		"ca_certificate": string(enrollment.CACertificate), "client_certificate": string(enrollment.CertificatePEM), "client_key": string(enrollment.PrivateKeyPEM),
		"certificate_serial": enrollment.Serial, "certificate_expires_at": enrollment.ExpiresAt, "renew_before_seconds": int64(renewBefore.Seconds()),
		"update_public_key": updatePublicKey, "update_key_id": updateKeyID,
		"heartbeat_interval": 30,
	}
}

func (a *API) mtlsAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 || len(c.Request.TLS.VerifiedChains) == 0 {
			fail(c, 401, "NODE_CERTIFICATE_REQUIRED", "节点客户端证书缺失或无效")
			c.Abort()
			return
		}
		certificate := c.Request.TLS.PeerCertificates[0]
		nodeID := nodepki.CertificateNodeID(certificate)
		var node store.Node
		if nodeID == "" || a.db.First(&node, "id = ?", nodeID).Error != nil || node.Status == "revoked" {
			fail(c, 401, "NODE_UNAUTHORIZED", "节点认证失败")
			c.Abort()
			return
		}
		now := time.Now()
		presentedSerial := nodepki.CertificateSerial(certificate)
		current := node.CertificateSerial != "" && strings.EqualFold(node.CertificateSerial, presentedSerial) && node.CertificateExpiresAt != nil && node.CertificateExpiresAt.After(now)
		pending := node.PendingCertificateSerial != "" && strings.EqualFold(node.PendingCertificateSerial, presentedSerial) && node.PendingCertificateActivationAt != nil && node.PendingCertificateActivationAt.After(now) && node.PendingCertificateNotAfter != nil && node.PendingCertificateNotAfter.After(now)
		if pending {
			result := a.db.Model(&store.Node{}).Where("id = ? AND pending_certificate_serial = ?", node.ID, node.PendingCertificateSerial).Updates(map[string]any{
				"certificate_serial": node.PendingCertificateSerial, "certificate_issued_at": now, "certificate_expires_at": node.PendingCertificateNotAfter,
				"pending_certificate_serial": "", "pending_certificate_not_after": nil, "pending_certificate_activation_at": nil,
			})
			if result.Error != nil || result.RowsAffected != 1 {
				fail(c, 409, "CERTIFICATE_CHANGED", "节点证书激活冲突，请重试")
				c.Abort()
				return
			}
			node.CertificateSerial = presentedSerial
			node.CertificateIssuedAt = &now
			node.CertificateExpiresAt = node.PendingCertificateNotAfter
			current = true
		}
		if !current {
			fail(c, 401, "NODE_CERTIFICATE_REVOKED", "节点证书已过期、已轮换或已吊销")
			c.Abort()
			return
		}
		c.Set("agent.node", node)
		c.Next()
	}
}

func (a *API) renewAgentCertificate(c *gin.Context) {
	node := c.MustGet("agent.node").(store.Node)
	enrollment, err := a.pki.IssueNode(node.ID, node.Name)
	if err != nil {
		fail(c, 500, "CERTIFICATE_ERROR", "节点证书续期失败")
		return
	}
	now := time.Now()
	activationDeadline := now.Add(15 * time.Minute)
	result := a.db.Model(&store.Node{}).
		Where("id = ? AND certificate_serial = ?", node.ID, node.CertificateSerial).
		Updates(map[string]any{"pending_certificate_serial": enrollment.Serial, "pending_certificate_not_after": enrollment.ExpiresAt, "pending_certificate_activation_at": activationDeadline})
	if result.Error != nil || result.RowsAffected != 1 {
		fail(c, 409, "CERTIFICATE_CHANGED", "节点证书已被其他操作更新，请重新注册")
		return
	}
	ok(c, enrollmentResponse(node.ID, a.cfg.AgentPublicURL, a.cfg.AgentRenewBefore, a.updateSigner.PublicKey(), a.updateSigner.KeyID(), enrollment))
}

type eventInput struct {
	EventID   string `json:"event_id"`
	PotID     string `json:"pot_id"`
	DecoyID   string `json:"decoy_id"`
	Service   string `json:"service"`
	EventType string `json:"event_type"`
	TS        int64  `json:"ts"`
	Src       struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	} `json:"src"`
	Dst struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	} `json:"dst"`
	Geo          string          `json:"geo"`
	ASN          string          `json:"asn"`
	RawPacket    string          `json:"raw_packet"`
	Payload      json.RawMessage `json:"payload"`
	Tags         json.RawMessage `json:"tags"`
	Detections   []detection.Hit `json:"detections"`
	RuleRevision int64           `json:"rule_revision"`
}

// eventRejection extends the original reject response without breaking older
// Agents. Fatal is omitted for retryable failures and batch-local duplicates;
// a newer Agent may move only fatal rejections to its local dead-letter file.
type eventRejection struct {
	EventID string `json:"event_id"`
	Reason  string `json:"reason"`
	Fatal   bool   `json:"fatal,omitempty"`
}

func rejectEvent(eventID, reason string, fatal bool) eventRejection {
	return eventRejection{EventID: eventID, Reason: reason, Fatal: fatal}
}

func (a *API) ingestEvents(c *gin.Context) {
	var inputs []eventInput
	compressed := http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	var reader io.Reader = compressed
	if strings.EqualFold(c.GetHeader("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(compressed)
		if err != nil {
			fail(c, 400, "INVALID_GZIP", "事件批次压缩格式无效")
			return
		}
		defer gzipReader.Close()
		reader = io.LimitReader(gzipReader, 2<<20)
	}
	if json.NewDecoder(reader).Decode(&inputs) != nil || len(inputs) == 0 || len(inputs) > 100 {
		fail(c, 400, "INVALID_BATCH", "事件批次必须包含 1-100 条记录")
		return
	}
	node := c.MustGet("agent.node").(store.Node)
	acks := []string{}
	rejects := []eventRejection{}
	accepted := make([]store.AttackEvent, 0, len(inputs))
	batchEventIDs := make(map[string]struct{}, len(inputs))
	for _, in := range inputs {
		if parsed, err := uuid.Parse(in.EventID); err != nil || parsed.String() != in.EventID {
			rejects = append(rejects, rejectEvent(in.EventID, "event_id must be a canonical UUID", true))
			continue
		}
		if _, duplicate := batchEventIDs[in.EventID]; duplicate {
			rejects = append(rejects, rejectEvent(in.EventID, "duplicate event_id in batch", false))
			continue
		}
		batchEventIDs[in.EventID] = struct{}{}
		if in.EventType == "" || in.TS == 0 {
			rejects = append(rejects, rejectEvent(in.EventID, "event_type and ts are required", true))
			continue
		}
		eventTime := time.Unix(in.TS, 0)
		receivedAt := time.Now()
		// Agent clocks may drift, but an extreme timestamp would either bypass
		// ClickHouse retention or be deleted before responders can investigate.
		// Reject it as a permanent bad record while preserving ordinary drift.
		if eventTime.Before(receivedAt.AddDate(-1, 0, -1)) || eventTime.After(receivedAt.Add(24*time.Hour)) {
			rejects = append(rejects, rejectEvent(in.EventID, "event timestamp is outside the accepted retention window", true))
			continue
		}
		if len(in.Payload) == 0 {
			in.Payload = json.RawMessage("{}")
		}
		if len(in.Tags) == 0 {
			in.Tags = json.RawMessage("[]")
		}
		payloadMap := map[string]any{}
		if json.Unmarshal(in.Payload, &payloadMap) != nil || payloadMap == nil {
			rejects = append(rejects, rejectEvent(in.EventID, "payload must be a JSON object", true))
			continue
		}
		var eventTags []string
		if json.Unmarshal(in.Tags, &eventTags) != nil {
			rejects = append(rejects, rejectEvent(in.EventID, "tags must be a JSON string array", true))
			continue
		}
		if in.RawPacket == "" {
			in.RawPacket, _ = payloadMap["raw_request"].(string)
		}
		if len(in.RawPacket) > 256<<10 {
			rejects = append(rejects, rejectEvent(in.EventID, "raw_packet exceeds 256 KiB", true))
			continue
		}
		serverHits, serverRuleRevision := a.matchServerDetectionsWithRevision(detection.Event{EventType: in.EventType, Service: in.Service, RawPacket: in.RawPacket, Payload: payloadMap})
		allHits := mergeDetectionHits(serverHits, in.Detections)
		detectionData, _ := json.Marshal(allHits)
		if len(allHits) > 0 {
			eventTags = appendUniqueStrings(eventTags, "detected", "server-reviewed")
		}
		if a.threatIntel != nil {
			if intelligence, found := a.threatIntel.Lookup(in.Src.IP); found {
				eventTags = appendUniqueStrings(eventTags, threatintel.EventTags(intelligence)...)
			}
		}
		in.Tags, _ = json.Marshal(eventTags)
		location, asn := a.eventLocation(in.Src.IP, in.Geo, in.ASN)
		item := store.AttackEvent{EventID: in.EventID, NodeID: node.ID, PotID: in.PotID, DecoyID: in.DecoyID, Service: in.Service, EventType: in.EventType, Timestamp: eventTime, SrcIP: in.Src.IP, SrcPort: in.Src.Port, DstIP: in.Dst.IP, DstPort: in.Dst.Port, Geo: location, ASN: asn, RawPacket: in.RawPacket, Payload: datatypes.JSON(in.Payload), Tags: datatypes.JSON(in.Tags), Detections: datatypes.JSON(detectionData), AgentRuleRevision: in.RuleRevision, ServerRuleRevision: serverRuleRevision, CreatedAt: receivedAt}
		accepted = append(accepted, item)
	}
	type pendingEvent struct {
		event       store.AttackEvent
		needsInsert bool
	}
	pending := make([]pendingEvent, 0, len(accepted))
	for _, submitted := range accepted {
		item := submitted
		receipt := store.EventReceipt{EventID: item.EventID, NodeID: node.ID, ReceivedAt: time.Now()}
		receiptResult := a.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&receipt)
		if receiptResult.Error != nil {
			rejects = append(rejects, rejectEvent(item.EventID, "control-plane receipt error", false))
			continue
		}
		var existingReceipt store.EventReceipt
		if err := a.db.First(&existingReceipt, "event_id = ?", item.EventID).Error; err != nil {
			rejects = append(rejects, rejectEvent(item.EventID, "control-plane receipt query error", false))
			continue
		}
		if existingReceipt.NodeID != node.ID {
			rejects = append(rejects, rejectEvent(item.EventID, "event_id belongs to another node", true))
			continue
		}
		if existingReceipt.ProcessedAt != nil {
			acks = append(acks, item.EventID)
			continue
		}
		canonicalFound := false
		if a.analytics != nil {
			canonical, err := a.analytics.Get(c.Request.Context(), item.EventID)
			switch {
			case err == nil:
				item = analytics.ToAttackEvent(canonical)
				canonicalFound = true
			case !errors.Is(err, analytics.ErrNotFound):
				fail(c, http.StatusServiceUnavailable, "ANALYTICS_UNAVAILABLE", "安全事件分析服务查询失败，Agent 将自动重试")
				return
			}
			if item.NodeID != node.ID {
				if receiptResult.RowsAffected > 0 {
					a.db.Where("event_id = ? AND node_id = ? AND processed_at IS NULL", item.EventID, node.ID).Delete(&store.EventReceipt{})
				}
				rejects = append(rejects, rejectEvent(submitted.EventID, "event_id belongs to another node", true))
				continue
			}
		} else {
			eventResult := a.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&item)
			if eventResult.Error != nil {
				rejects = append(rejects, rejectEvent(item.EventID, "database error", false))
				continue
			}
			var canonical store.AttackEvent
			if err := a.db.First(&canonical, "event_id = ?", item.EventID).Error; err != nil {
				rejects = append(rejects, rejectEvent(item.EventID, "database event query error", false))
				continue
			}
			if canonical.NodeID != node.ID {
				if receiptResult.RowsAffected > 0 {
					a.db.Where("event_id = ? AND node_id = ? AND processed_at IS NULL", item.EventID, node.ID).Delete(&store.EventReceipt{})
				}
				rejects = append(rejects, rejectEvent(item.EventID, "event_id belongs to another node", true))
				continue
			}
			item = canonical
		}
		pending = append(pending, pendingEvent{event: item, needsInsert: a.analytics != nil && !canonicalFound})
	}
	for _, pendingItem := range pending {
		item := pendingItem.event
		if a.analytics != nil && pendingItem.needsInsert {
			if err := a.analytics.InsertEvent(c.Request.Context(), analytics.FromAttackEvent(item)); err != nil {
				receiptUpdate := a.db.Model(&store.EventReceipt{}).Where("event_id = ? AND node_id = ? AND processed_at IS NULL", item.EventID, node.ID).Update("last_error", "ClickHouse event write failed")
				if receiptUpdate.Error != nil || receiptUpdate.RowsAffected != 1 {
					fail(c, http.StatusServiceUnavailable, "RECEIPT_UNAVAILABLE", "事件处理状态写入失败，Agent 将自动重试")
					return
				}
				rejects = append(rejects, rejectEvent(item.EventID, "analytics event write failed", false))
				continue
			}
		}
		alerts, relatedEvents, processErr := a.processEventBusinessEffects(c.Request.Context(), item)
		if processErr != nil {
			receiptUpdate := a.db.Model(&store.EventReceipt{}).Where("event_id = ? AND node_id = ? AND processed_at IS NULL", item.EventID, node.ID).Update("last_error", truncateEventProcessingError(processErr.Error()))
			if receiptUpdate.Error != nil || receiptUpdate.RowsAffected != 1 {
				fail(c, http.StatusServiceUnavailable, "RECEIPT_UNAVAILABLE", "事件处理状态写入失败，Agent 将自动重试")
				return
			}
			rejects = append(rejects, rejectEvent(item.EventID, "business side effects failed", false))
			continue
		}
		acks = append(acks, item.EventID)
		a.alerts.Notify()
		a.hub.Broadcast("event.new", a.realtimeEventView(item))
		for _, alert := range alerts {
			a.hub.Broadcast("alert.new", alert)
		}
		for _, related := range relatedEvents {
			a.hub.Broadcast("event.new", a.realtimeEventView(related))
		}
	}
	ok(c, gin.H{"ack_ids": acks, "reject": rejects})
}

func (a *API) eventLocation(address, fallbackGeo, fallbackASN string) (string, string) {
	location, asn := strings.TrimSpace(fallbackGeo), strings.TrimSpace(fallbackASN)
	if a.geoIP == nil || strings.TrimSpace(address) == "" {
		return location, asn
	}
	resolved, err := a.geoIP.Lookup(address)
	if err != nil || resolved.Geo == "" {
		return location, asn
	}
	if resolved.Internal {
		return resolved.Geo, ""
	}
	location = resolved.Geo
	if resolved.ASN != "" {
		asn = resolved.ASN
	}
	return location, asn
}

func (a *API) backfillEventGeo(ctx context.Context) {
	const batchSize = 250
	var cursorTime time.Time
	cursorID := ""
	updated := int64(0)
	for {
		if ctx.Err() != nil {
			return
		}
		var events []store.AttackEvent
		query := a.db.WithContext(ctx).Where("src_ip <> '' AND (geo = '' OR geo IS NULL)")
		if !cursorTime.IsZero() {
			query = query.Where("created_at > ? OR (created_at = ? AND event_id > ?)", cursorTime, cursorTime, cursorID)
		}
		result := query.Order("created_at ASC").Order("event_id ASC").Limit(batchSize).Find(&events)
		if result.Error != nil {
			if ctx.Err() == nil {
				log.Printf("IPIP historical event backfill stopped: %v", result.Error)
			}
			return
		}
		if len(events) == 0 {
			break
		}
		for _, event := range events {
			location, asn := a.eventLocation(event.SrcIP, "", event.ASN)
			if location == "" {
				continue
			}
			values := map[string]any{"geo": location}
			if event.ASN == "" && asn != "" {
				values["asn"] = asn
			}
			write := a.db.WithContext(ctx).Model(&store.AttackEvent{}).
				Where("event_id = ? AND (geo = '' OR geo IS NULL)", event.EventID).Updates(values)
			if write.Error == nil {
				updated += write.RowsAffected
			}
		}
		last := events[len(events)-1]
		cursorTime, cursorID = last.CreatedAt, last.EventID
		if len(events) < batchSize {
			break
		}
	}
	if updated > 0 {
		log.Printf("IPIP geolocation backfilled %d historical events", updated)
	}
}

func recordDecoyHit(tx *gorm.DB, decoyID, nodeID string, hitAt time.Time) error {
	result := tx.Model(&store.Decoy{}).Where("id = ? AND node_id = ?", decoyID, nodeID).Updates(map[string]any{
		"hit_count": gorm.Expr("hit_count + 1"), "last_hit_at": hitAt,
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (a *API) matchNetworkDecoys(ctx context.Context, tx *gorm.DB, source store.AttackEvent) ([]store.AttackEvent, []store.Alert, error) {
	if strings.HasPrefix(source.EventType, "decoy.") || len(source.Payload) == 0 {
		return nil, nil, nil
	}
	var decoys []store.Decoy
	if err := tx.Where("node_id = ? AND type = ? AND status = ?", source.NodeID, "network", "enabled").Find(&decoys).Error; err != nil {
		return nil, nil, err
	}
	relatedEvents := []store.AttackEvent{}
	createdAlerts := []store.Alert{}
	for _, decoy := range decoys {
		spec, isMatch, err := networkDecoyMatch(decoy, source.Payload)
		if err != nil {
			return nil, nil, fmt.Errorf("parse network decoy %s: %w", decoy.ID, err)
		}
		if !isMatch {
			continue
		}
		payload, _ := json.Marshal(map[string]any{
			"decoy_id": decoy.ID, "decoy_name": decoy.Name, "decoy_type": decoy.Type,
			"description": spec.Description, "source_event_id": source.EventID, "source_event_type": source.EventType,
		})
		tags, _ := json.Marshal([]string{"decoy", "correlated"})
		matchedEvent := store.AttackEvent{
			EventID: stableUUID("network-decoy-event", source.EventID+"|"+decoy.ID), NodeID: source.NodeID, DecoyID: decoy.ID, Service: "decoy", EventType: "decoy.network",
			Timestamp: source.Timestamp, SrcIP: source.SrcIP, SrcPort: source.SrcPort, DstIP: source.DstIP, DstPort: source.DstPort,
			Geo: source.Geo, ASN: source.ASN,
			Payload: datatypes.JSON(payload), Tags: datatypes.JSON(tags), Detections: datatypes.JSON(`[]`), CreatedAt: time.Now(),
		}
		if a.analytics != nil {
			receipt := store.EventReceipt{EventID: matchedEvent.EventID, NodeID: matchedEvent.NodeID, ReceivedAt: time.Now()}
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&receipt)
			if result.Error != nil {
				return nil, nil, result.Error
			}
			var existing store.EventReceipt
			if err := tx.First(&existing, "event_id = ?", matchedEvent.EventID).Error; err != nil {
				return nil, nil, err
			}
			if existing.NodeID != matchedEvent.NodeID {
				return nil, nil, errors.New("correlated event_id belongs to another node")
			}
			if existing.ProcessedAt != nil {
				continue
			}
			canonical, err := a.analytics.Get(ctx, matchedEvent.EventID)
			switch {
			case err == nil:
				matchedEvent = analytics.ToAttackEvent(canonical)
				if matchedEvent.NodeID != source.NodeID || matchedEvent.DecoyID != decoy.ID {
					return nil, nil, errors.New("correlated event_id canonical ownership mismatch")
				}
			case errors.Is(err, analytics.ErrNotFound):
				if err := a.analytics.InsertEvent(ctx, analytics.FromAttackEvent(matchedEvent)); err != nil {
					return nil, nil, fmt.Errorf("write correlated decoy event: %w", err)
				}
			default:
				return nil, nil, fmt.Errorf("read correlated decoy event: %w", err)
			}
		} else {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&matchedEvent)
			if result.Error != nil {
				return nil, nil, result.Error
			}
			if result.RowsAffected == 0 {
				var existing store.AttackEvent
				if err := tx.First(&existing, "event_id = ?", matchedEvent.EventID).Error; err != nil {
					return nil, nil, err
				}
				if existing.NodeID != matchedEvent.NodeID {
					return nil, nil, errors.New("correlated event_id belongs to another node")
				}
				var receipt store.EventReceipt
				err := tx.First(&receipt, "event_id = ?", matchedEvent.EventID).Error
				if err == nil && receipt.ProcessedAt != nil {
					continue
				}
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, nil, err
				}
			}
			receipt := store.EventReceipt{EventID: matchedEvent.EventID, NodeID: matchedEvent.NodeID, ReceivedAt: time.Now()}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&receipt).Error; err != nil {
				return nil, nil, err
			}
		}
		if err := recordDecoyHit(tx, decoy.ID, source.NodeID, source.Timestamp); err != nil {
			return nil, nil, err
		}
		alerts, err := a.createAlertAndIOC(ctx, tx, matchedEvent)
		if err != nil {
			return nil, nil, err
		}
		now := time.Now()
		result := tx.Model(&store.EventReceipt{}).Where("event_id = ? AND node_id = ? AND processed_at IS NULL", matchedEvent.EventID, matchedEvent.NodeID).Updates(map[string]any{"processed_at": now, "last_error": ""})
		if result.Error != nil {
			return nil, nil, fmt.Errorf("mark correlated event processed: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return nil, nil, errors.New("correlated event receipt already processed or missing")
		}
		relatedEvents = append(relatedEvents, matchedEvent)
		createdAlerts = append(createdAlerts, alerts...)
	}
	return relatedEvents, createdAlerts, nil
}

func networkDecoyMatch(decoy store.Decoy, payload []byte) (decoyconfig.Spec, bool, error) {
	spec, err := decoyconfig.Parse(decoy.Type, decoy.Config)
	if err != nil {
		return spec, false, err
	}
	return spec, bytes.Contains(payload, []byte(spec.Token)), nil
}

func stableUUID(namespace, value string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(namespace+"|"+value)).String()
}

func truncateEventProcessingError(message string) string {
	if len(message) > 1000 {
		return message[:1000]
	}
	return message
}

func (a *API) processEventBusinessEffects(ctx context.Context, event store.AttackEvent) ([]store.Alert, []store.AttackEvent, error) {
	createdAlerts := []store.Alert{}
	relatedEvents := []store.AttackEvent{}
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if event.DecoyID != "" && strings.HasPrefix(event.EventType, "decoy.") {
			if err := recordDecoyHit(tx, event.DecoyID, event.NodeID, event.Timestamp); err != nil {
				return err
			}
		}
		alerts, err := a.createAlertAndIOC(ctx, tx, event)
		if err != nil {
			return err
		}
		createdAlerts = append(createdAlerts, alerts...)
		related, relatedAlerts, err := a.matchNetworkDecoys(ctx, tx, event)
		if err != nil {
			return err
		}
		relatedEvents = append(relatedEvents, related...)
		createdAlerts = append(createdAlerts, relatedAlerts...)
		now := time.Now()
		result := tx.Model(&store.EventReceipt{}).Where("event_id = ? AND node_id = ? AND processed_at IS NULL", event.EventID, event.NodeID).Updates(map[string]any{"processed_at": now, "last_error": ""})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("event receipt already processed or missing")
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return createdAlerts, relatedEvents, nil
}

func (a *API) createAlertAndIOC(ctx context.Context, tx *gorm.DB, event store.AttackEvent) ([]store.Alert, error) {
	alerts, err := a.createDetectionAlerts(tx, event)
	if err != nil {
		return nil, err
	}
	thresholdAlerts, err := a.evaluateAlertRules(ctx, tx, event)
	if err != nil {
		return nil, err
	}
	alerts = append(alerts, thresholdAlerts...)
	if event.SrcIP != "" {
		now := time.Now()
		ioc := store.IOC{Base: store.NewBase(), Type: "ip", Value: event.SrcIP, Source: "honeynet", Confidence: 60, EventID: event.EventID, FirstSeen: now, LastSeen: now}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "type"}, {Name: "value"}}, DoUpdates: clause.Assignments(map[string]any{"last_seen": now, "event_id": event.EventID, "updated_at": now})}).Create(&ioc).Error; err != nil {
			return nil, err
		}
	}
	return alerts, nil
}

func (a *API) evaluateAlertRules(ctx context.Context, tx *gorm.DB, event store.AttackEvent) ([]store.Alert, error) {
	var rules []store.AlertRule
	if err := tx.Where("enabled = ?", true).Order("created_at ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	createdAlerts := []store.Alert{}
	for _, rule := range rules {
		matched, err := path.Match(rule.EventType, event.EventType)
		if err != nil {
			return nil, fmt.Errorf("invalid alert rule %s pattern: %w", rule.ID, err)
		}
		if !matched {
			continue
		}
		threshold := rule.Threshold
		if threshold < 1 {
			threshold = 1
		}
		window := rule.WindowMinute
		if window < 1 {
			window = 1
		}
		var count int64
		if a.analytics != nil {
			value, err := a.analytics.CountAlertWindow(ctx, analytics.AlertWindowFilter{
				From: event.Timestamp.Add(-time.Duration(window) * time.Minute), To: event.Timestamp,
				EventTypePattern: rule.EventType, Service: event.Service, SourceIP: event.SrcIP, NodeID: event.NodeID,
			})
			if err != nil {
				return nil, err
			}
			count = int64(value)
		} else {
			query := tx.Model(&store.AttackEvent{}).
				Where("ts >= ? AND ts <= ? AND event_type LIKE ? AND service = ?", event.Timestamp.Add(-time.Duration(window)*time.Minute), event.Timestamp, wildcardSQL(rule.EventType), event.Service)
			if event.SrcIP != "" {
				query = query.Where("src_ip = ?", event.SrcIP)
			} else {
				query = query.Where("node_id = ?", event.NodeID)
			}
			if err := query.Count(&count).Error; err != nil {
				return nil, err
			}
		}
		if count < int64(threshold) {
			continue
		}
		silenceRaw := fmt.Sprintf("%s|%s|%s|%s", rule.ID, event.SrcIP, event.NodeID, event.Service)
		silenceSum := sha256.Sum256([]byte(silenceRaw))
		silenceKey := hex.EncodeToString(silenceSum[:])
		if rule.SilenceMinute > 0 {
			var recent int64
			if err := tx.Model(&store.Alert{}).Where("silence_key = ? AND created_at >= ?", silenceKey, time.Now().Add(-time.Duration(rule.SilenceMinute)*time.Minute)).Count(&recent).Error; err != nil {
				return nil, err
			}
			if recent > 0 {
				continue
			}
		}
		fingerprintSum := sha256.Sum256([]byte(silenceKey + "|" + event.EventID))
		fingerprint := hex.EncodeToString(fingerprintSum[:])
		alert := store.Alert{
			Base: store.Base{ID: stableUUID("threshold-alert", rule.ID+"|"+event.EventID)}, EventID: event.EventID, RuleID: rule.ID, Fingerprint: fingerprint, SilenceKey: silenceKey,
			Title: fmt.Sprintf("%s：%s 捕获 %s", rule.Name, strings.ToUpper(event.Service), event.EventType),
			Level: rule.Level, Status: "new", SourceIP: event.SrcIP, NodeID: event.NodeID, Service: event.Service,
			Description: fmt.Sprintf("来源 %s:%d 访问目标端口 %d，%d 分钟窗口内命中 %d 次", event.SrcIP, event.SrcPort, event.DstPort, window, count),
		}
		if alert.Level == "" {
			alert.Level = "medium"
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&alert)
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected > 0 {
			var channelIDs []string
			if len(rule.ChannelIDs) > 0 {
				if err := json.Unmarshal(rule.ChannelIDs, &channelIDs); err != nil {
					return nil, err
				}
			}
			if err := a.alerts.EnqueueWithDB(tx, alert, channelIDs); err != nil {
				return nil, err
			}
			createdAlerts = append(createdAlerts, alert)
		}
	}
	return createdAlerts, nil
}

func wildcardSQL(pattern string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_", "*", "%")
	return replacer.Replace(pattern)
}
func (a *API) agentHeartbeat(c *gin.Context) {
	node := c.MustGet("agent.node").(store.Node)
	var req struct {
		Version string   `json:"version"`
		IP      string   `json:"ip"`
		IPs     []string `json:"ips"`
		OS      string   `json:"os"`
		Arch    string   `json:"arch"`
		Healthy bool     `json:"healthy"`
	}
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "心跳格式错误")
		return
	}
	if len(req.Version) > 64 || len(req.OS) > 64 || len(req.Arch) > 64 || len(req.IPs) > 128 {
		fail(c, 400, "INVALID_ARGUMENT", "心跳字段超过允许范围")
		return
	}
	status := "online"
	if !req.Healthy {
		status = "degraded"
	}
	now := time.Now()
	a.db.Model(&node).Updates(map[string]any{"version": req.Version, "os": req.OS, "arch": req.Arch, "status": status, "last_heartbeat_at": now})
	reportedIPs := req.IPs
	if req.IP != "" {
		reportedIPs = append(append([]string(nil), reportedIPs...), req.IP)
	}
	updated, err := updateNodeAddressReport(a.db, node.ID, requestRemoteIP(c.Request), reportedIPs)
	if err == nil {
		node = updated
	}
	node.Status, node.Version, node.OS, node.Arch, node.LastHeartbeatAt = status, req.Version, req.OS, req.Arch, &now
	a.hub.Broadcast("node.status", node)
	ok(c, gin.H{"server_time": now.Unix(), "heartbeat_interval": 30})
}
