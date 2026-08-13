package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/detection"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type detectionRuntime struct {
	mu        sync.RWMutex
	matcher   *detection.Matcher
	revision  int64
	ruleCount int
}

type detectionImportReport struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
	Pending int `json:"pending_review"`
}

// JSON numbers cannot represent nanosecond Unix revisions exactly in a
// browser. Keep the numeric field for the Agent protocol and expose an exact
// decimal string for console/audit display.
func detectionRuleView(item store.DetectionRule) gin.H {
	data, _ := json.Marshal(item)
	view := gin.H{}
	_ = json.Unmarshal(data, &view)
	view["revision_text"] = strconv.FormatInt(item.Revision, 10)
	return view
}

type detectionRuleRequest struct {
	Key           *string              `json:"key"`
	Name          *string              `json:"name"`
	Description   *string              `json:"description"`
	Severity      *string              `json:"severity"`
	Enabled       *bool                `json:"enabled"`
	AgentEnabled  *bool                `json:"agent_enabled"`
	ServerEnabled *bool                `json:"server_enabled"`
	Patterns      *[]detection.Pattern `json:"patterns"`
}

func (a *API) listDetectionRules(c *gin.Context) {
	p, size := page(c)
	query := a.db.Model(&store.DetectionRule{})
	if source := strings.TrimSpace(c.Query("source")); source != "" {
		query = query.Where("source = ?", source)
	}
	if enabled := strings.TrimSpace(c.Query("enabled")); enabled == "true" || enabled == "false" {
		query = query.Where("enabled = ?", enabled == "true")
	}
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		pattern := store.LikePattern(search)
		query = query.Where("name LIKE ? OR rule_key LIKE ? OR external_id LIKE ?", pattern, pattern, pattern)
	}
	var total int64
	query.Count(&total)
	var items []store.DetectionRule
	query.Order("validation_error <> '' DESC").Order("severity DESC").Order("rule_key ASC").Offset((p - 1) * size).Limit(size).Find(&items)
	views := make([]gin.H, 0, len(items))
	for _, item := range items {
		views = append(views, detectionRuleView(item))
	}
	ok(c, pageResult(views, total, p, size))
}

func (a *API) createDetectionRule(c *gin.Context) {
	var request detectionRuleRequest
	if c.ShouldBindJSON(&request) != nil || request.Key == nil || request.Name == nil || request.Patterns == nil {
		fail(c, http.StatusBadRequest, "INVALID_ARGUMENT", "规则标识、名称和匹配特征不能为空")
		return
	}
	item := store.DetectionRule{
		Base: store.NewBase(), RuleKey: strings.TrimSpace(*request.Key), Name: strings.TrimSpace(*request.Name),
		Severity: "medium", Source: "custom", Enabled: true, AgentEnabled: true, ServerEnabled: true, Revision: time.Now().UnixNano(),
	}
	if err := applyDetectionRuleRequest(&item, request); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_RULE", err.Error())
		return
	}
	if err := validateStoredDetectionRule(item); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_RULE", err.Error())
		return
	}
	if err := a.db.Create(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			fail(c, http.StatusConflict, "RULE_EXISTS", "规则标识已存在")
			return
		}
		fail(c, http.StatusInternalServerError, "CREATE_FAILED", "创建检测规则失败")
		return
	}
	a.afterDetectionRuleChange()
	created(c, detectionRuleView(item))
}

func (a *API) updateDetectionRule(c *gin.Context) {
	var item store.DetectionRule
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, http.StatusNotFound, "RULE_NOT_FOUND", "检测规则不存在")
		return
	}
	var request detectionRuleRequest
	if c.ShouldBindJSON(&request) != nil {
		fail(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请求格式错误")
		return
	}
	if err := applyDetectionRuleRequest(&item, request); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_RULE", err.Error())
		return
	}
	item.Revision = time.Now().UnixNano()
	item.ValidationError = ""
	if err := validateStoredDetectionRule(item); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_RULE", err.Error())
		return
	}
	updates := map[string]any{
		"rule_key": item.RuleKey, "name": item.Name, "description": item.Description, "severity": item.Severity,
		"enabled": item.Enabled, "agent_enabled": item.AgentEnabled, "server_enabled": item.ServerEnabled,
		"patterns": item.Patterns, "revision": item.Revision, "validation_error": "",
	}
	if err := a.db.Model(&store.DetectionRule{}).Where("id = ?", item.ID).Updates(updates).Error; err != nil {
		fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "更新检测规则失败")
		return
	}
	a.db.First(&item, "id = ?", item.ID)
	a.afterDetectionRuleChange()
	ok(c, detectionRuleView(item))
}

func (a *API) deleteDetectionRule(c *gin.Context) {
	result := a.db.Delete(&store.DetectionRule{}, "id = ?", c.Param("id"))
	if result.Error != nil {
		fail(c, http.StatusInternalServerError, "DELETE_FAILED", "删除检测规则失败")
		return
	}
	if result.RowsAffected == 0 {
		fail(c, http.StatusNotFound, "RULE_NOT_FOUND", "检测规则不存在")
		return
	}
	a.afterDetectionRuleChange()
	okEmpty(c)
}

func (a *API) testDetectionRule(c *gin.Context) {
	var item store.DetectionRule
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, http.StatusNotFound, "RULE_NOT_FOUND", "检测规则不存在")
		return
	}
	var request struct {
		EventType string         `json:"event_type"`
		Service   string         `json:"service"`
		RawPacket string         `json:"raw_packet"`
		Payload   map[string]any `json:"payload"`
	}
	if c.ShouldBindJSON(&request) != nil {
		fail(c, http.StatusBadRequest, "INVALID_ARGUMENT", "测试事件格式错误")
		return
	}
	rule, err := detectionProtocolRule(item)
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID_RULE", err.Error())
		return
	}
	matcher, _ := detection.Compile([]detection.Rule{rule})
	hits := matcher.Match(detection.Event{EventType: request.EventType, Service: request.Service, RawPacket: request.RawPacket, Payload: request.Payload}, "test")
	ok(c, gin.H{"matched": len(hits) > 0, "hits": hits})
}

func (a *API) importDetectionRules(c *gin.Context) {
	report, err := a.importConfiguredDetectionRules()
	if err != nil {
		fail(c, http.StatusBadRequest, "IMPORT_FAILED", err.Error())
		return
	}
	a.afterDetectionRuleChange()
	ok(c, report)
}

func (a *API) importConfiguredDetectionRules() (detectionImportReport, error) {
	items, err := detection.ImportRuleDirectory(a.cfg.DetectionRulesDir)
	if err != nil {
		return detectionImportReport{}, err
	}
	report := detectionImportReport{Total: len(items)}
	var stored []store.DetectionRule
	if err := a.db.Find(&stored).Error; err != nil {
		return report, err
	}
	byKey := make(map[string]*store.DetectionRule, len(stored))
	byExternalID := make(map[string]*store.DetectionRule, len(stored))
	for index := range stored {
		item := &stored[index]
		byKey[item.RuleKey] = item
		if item.ExternalID != "" && item.Source != "custom" {
			byExternalID[item.ExternalID] = item
		}
	}
	for _, imported := range items {
		patterns, _ := json.Marshal(imported.Rule.Patterns)
		existing := byKey[imported.Rule.Key]
		if existing == nil {
			existing = byExternalID[imported.Rule.ExternalID]
		}
		if existing == nil {
			enabled := imported.ValidationError == ""
			item := store.DetectionRule{
				Base: store.NewBase(), RuleKey: imported.Rule.Key, Name: imported.Rule.Name, Description: imported.Rule.Description,
				Severity: imported.Rule.Severity, Source: imported.Rule.Source, ExternalID: imported.Rule.ExternalID,
				Enabled: enabled, AgentEnabled: enabled, ServerEnabled: enabled, Revision: time.Now().UnixNano(),
				Patterns: datatypes.JSON(patterns), OriginalCondition: imported.Original, ValidationError: imported.ValidationError,
			}
			if err := a.db.Create(&item).Error; err != nil {
				return report, err
			}
			byKey[item.RuleKey] = &item
			byExternalID[item.ExternalID] = &item
			report.Created++
		} else {
			needsUpdate := existing.RuleKey != imported.Rule.Key || existing.Name != imported.Rule.Name ||
				existing.Description != imported.Rule.Description || existing.Severity != imported.Rule.Severity ||
				existing.Source != imported.Rule.Source || existing.ExternalID != imported.Rule.ExternalID ||
				!storedPatternsEqual(existing.Patterns, imported.Rule.Patterns) || existing.OriginalCondition != imported.Original ||
				existing.ValidationError != imported.ValidationError
			if imported.ValidationError != "" && (existing.Enabled || existing.AgentEnabled || existing.ServerEnabled) {
				needsUpdate = true
			}
			if !needsUpdate {
				if imported.ValidationError != "" {
					report.Pending++
				}
				continue
			}
			updates := map[string]any{
				"rule_key": imported.Rule.Key, "name": imported.Rule.Name, "description": imported.Rule.Description,
				"severity": imported.Rule.Severity, "source": imported.Rule.Source,
				"external_id": imported.Rule.ExternalID, "patterns": datatypes.JSON(patterns),
				"original_condition": imported.Original, "validation_error": imported.ValidationError, "revision": time.Now().UnixNano(),
			}
			if imported.ValidationError != "" {
				updates["enabled"], updates["agent_enabled"], updates["server_enabled"] = false, false, false
			}
			if err := a.db.Model(&store.DetectionRule{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				return report, err
			}
			report.Updated++
		}
		if imported.ValidationError != "" {
			report.Pending++
		}
	}
	return report, nil
}

func storedPatternsEqual(raw []byte, expected []detection.Pattern) bool {
	var current []detection.Pattern
	return json.Unmarshal(raw, &current) == nil && reflect.DeepEqual(current, expected)
}

func applyDetectionRuleRequest(item *store.DetectionRule, request detectionRuleRequest) error {
	if request.Key != nil {
		item.RuleKey = strings.TrimSpace(*request.Key)
	}
	if request.Name != nil {
		item.Name = strings.TrimSpace(*request.Name)
	}
	if request.Description != nil {
		item.Description = strings.TrimSpace(*request.Description)
	}
	if request.Severity != nil {
		item.Severity = strings.ToLower(strings.TrimSpace(*request.Severity))
	}
	if request.Enabled != nil {
		item.Enabled = *request.Enabled
	}
	if request.AgentEnabled != nil {
		item.AgentEnabled = *request.AgentEnabled
	}
	if request.ServerEnabled != nil {
		item.ServerEnabled = *request.ServerEnabled
	}
	if request.Patterns != nil {
		data, err := json.Marshal(*request.Patterns)
		if err != nil {
			return err
		}
		item.Patterns = data
	}
	return nil
}

func validateStoredDetectionRule(item store.DetectionRule) error {
	switch item.Severity {
	case "critical", "high", "medium", "low", "info":
	default:
		return errors.New("风险级别必须为 critical、high、medium、low 或 info")
	}
	rule, err := detectionProtocolRule(item)
	if err != nil {
		return err
	}
	return detection.ValidateRule(rule)
}

func detectionProtocolRule(item store.DetectionRule) (detection.Rule, error) {
	var patterns []detection.Pattern
	if err := json.Unmarshal(item.Patterns, &patterns); err != nil {
		return detection.Rule{}, err
	}
	return detection.Rule{
		ID: item.ID, Key: item.RuleKey, Name: item.Name, Description: item.Description, Severity: item.Severity,
		Source: item.Source, ExternalID: item.ExternalID, Revision: item.Revision, Patterns: patterns,
	}, nil
}

func (a *API) refreshDetectionMatcher() error {
	var items []store.DetectionRule
	if err := a.db.Where("enabled = ? AND server_enabled = ? AND validation_error = ''", true, true).Order("rule_key ASC").Find(&items).Error; err != nil {
		return err
	}
	rules := make([]detection.Rule, 0, len(items))
	var revision int64
	for _, item := range items {
		rule, err := detectionProtocolRule(item)
		if err != nil {
			return err
		}
		rules = append(rules, rule)
		if item.Revision > revision {
			revision = item.Revision
		}
	}
	matcher, err := detection.Compile(rules)
	if err != nil {
		return err
	}
	a.detection.mu.Lock()
	a.detection.matcher = matcher
	a.detection.revision = revision
	a.detection.ruleCount = len(rules)
	a.detection.mu.Unlock()
	return nil
}

func (a *API) afterDetectionRuleChange() {
	if err := a.refreshDetectionMatcher(); err != nil {
		return
	}
	a.agents.BroadcastDetectionRules()
	revision, count := a.serverDetectionRuleStatus()
	a.db.Model(&store.Node{}).Where("detection_rule_revision <> ?", revision).Update("detection_rule_status", "stale")
	a.hub.Broadcast("detection.rules", gin.H{"revision": revision, "revision_text": strconv.FormatInt(revision, 10), "rule_count": count})
}

func detectionRuleSetStatus(db *gorm.DB, agent bool) (int64, int, error) {
	var items []store.DetectionRule
	if err := db.Where("enabled = ? AND validation_error = ''", true).Find(&items).Error; err != nil {
		return 0, 0, err
	}
	revision, count := summarizeDetectionRules(items, agent)
	return revision, count, nil
}

func summarizeDetectionRules(items []store.DetectionRule, agent bool) (int64, int) {
	var revision int64
	count := 0
	for _, item := range items {
		applies := item.AgentEnabled
		if !agent {
			applies = item.ServerEnabled
		}
		if !item.Enabled || item.ValidationError != "" || !applies {
			continue
		}
		count++
		if item.Revision > revision {
			revision = item.Revision
		}
	}
	return revision, count
}

func (a *API) matchServerDetections(event detection.Event) []detection.Hit {
	hits, _ := a.matchServerDetectionsWithRevision(event)
	return hits
}

func (a *API) matchServerDetectionsWithRevision(event detection.Event) ([]detection.Hit, int64) {
	a.detection.mu.RLock()
	matcher := a.detection.matcher
	revision := a.detection.revision
	a.detection.mu.RUnlock()
	return matcher.Match(event, "server"), revision
}

func (a *API) serverDetectionRuleStatus() (int64, int) {
	a.detection.mu.RLock()
	defer a.detection.mu.RUnlock()
	return a.detection.revision, a.detection.ruleCount
}

func (a *API) detectionPipelineStatus(c *gin.Context) {
	serverRevision, serverCount := a.serverDetectionRuleStatus()
	agentRevision, agentCount, _ := detectionRuleSetStatus(a.db, true)
	var nodes []store.Node
	_ = a.db.Select("id", "name", "group_name", "version", "status", "queued_events", "detection_rule_revision", "detection_rule_count", "detection_rule_status", "detection_rule_synced_at", "detection_rule_error").Where("status <> ?", "revoked").Order("name ASC").Find(&nodes).Error
	nodeStatus := gin.H{"total": len(nodes), "online": 0, "synced": 0, "stale": 0, "error": 0, "pending": 0, "queued_events": 0}
	nodeItems := make([]gin.H, 0, len(nodes))
	for _, node := range nodes {
		if node.Status == "online" || node.Status == "degraded" {
			nodeStatus["online"] = nodeStatus["online"].(int) + 1
		}
		status := node.DetectionRuleStatus
		if status == "" {
			status = "pending"
		}
		if _, exists := nodeStatus[status]; !exists {
			status = "pending"
		}
		nodeStatus[status] = nodeStatus[status].(int) + 1
		nodeStatus["queued_events"] = nodeStatus["queued_events"].(int) + node.QueuedEvents
		nodeItems = append(nodeItems, detectionPipelineNodeView(node, agentRevision, status))
	}
	nodeStatus["items"] = nodeItems
	ok(c, gin.H{
		"mode":         "edge-prefilter-central-review",
		"rule_format":  "portable-yara-compatible",
		"agent_rules":  gin.H{"revision": agentRevision, "revision_text": strconv.FormatInt(agentRevision, 10), "rule_count": agentCount},
		"server_rules": gin.H{"revision": serverRevision, "revision_text": strconv.FormatInt(serverRevision, 10), "rule_count": serverCount},
		"nodes":        nodeStatus,
		"stages": []gin.H{
			{"key": "agent_scan", "name": "Agent 实时 YARA 兼容扫描", "purpose": "攻击发生时就地打标并分摊计算"},
			{"key": "rule_distribution", "name": "Server 规则存储与下发", "purpose": "MySQL 统一管理规则版本并推送在线节点"},
			{"key": "server_review", "name": "Server 入库复核", "purpose": "使用中心统一版本兜底重扫，补偿节点离线或规则滞后"},
			{"key": "alert_generation", "name": "Server 告警生成", "purpose": "仅基于中心复核命中生成 rule_alert，避免边缘误报直接升级"},
		},
	})
}

func detectionPipelineNodeView(node store.Node, targetRevision int64, syncStatus string) gin.H {
	return gin.H{
		"id": node.ID, "name": node.Name, "group": node.GroupName, "agent_version": node.Version,
		"node_status": node.Status, "sync_status": syncStatus, "rule_count": node.DetectionRuleCount,
		"revision_text":        strconv.FormatInt(node.DetectionRuleRevision, 10),
		"target_revision_text": strconv.FormatInt(targetRevision, 10),
		"synced_at":            node.DetectionRuleSyncedAt, "error": node.DetectionRuleError,
		"queued_events": node.QueuedEvents,
	}
}

func mergeDetectionHits(serverHits, agentHits []detection.Hit) []detection.Hit {
	agentByKey := make(map[string]detection.Hit, len(agentHits))
	for _, hit := range agentHits {
		hit = normalizeDetectionHit(hit)
		if hit.RuleKey != "" {
			hit.Stage = "agent"
			agentByKey[hit.RuleKey] = hit
		}
	}
	result := make([]detection.Hit, 0, len(serverHits)+len(agentHits))
	for _, hit := range serverHits {
		hit = normalizeDetectionHit(hit)
		if _, exists := agentByKey[hit.RuleKey]; exists {
			hit.Confirmed = true
			delete(agentByKey, hit.RuleKey)
		}
		result = append(result, hit)
	}
	for _, hit := range agentByKey {
		hit.Confirmed = false
		result = append(result, hit)
	}
	return result
}

func normalizeDetectionHit(hit detection.Hit) detection.Hit {
	if hit.Source == "custom" {
		return hit
	}
	if hit.Source != "builtin" {
		hit.Source = "imported"
	}
	if hit.ExternalID != "" {
		hit.Source = "builtin"
		hit.RuleKey = "builtin:" + hit.ExternalID
	}
	return hit
}

func normalizeEventDetections(event *store.AttackEvent) {
	var hits []detection.Hit
	if event == nil || json.Unmarshal(event.Detections, &hits) != nil {
		return
	}
	for index := range hits {
		hits[index] = normalizeDetectionHit(hits[index])
	}
	data, err := json.Marshal(hits)
	if err == nil {
		event.Detections = data
	}
}

func appendUniqueStrings(values []string, additions ...string) []string {
	seen := make(map[string]bool, len(values)+len(additions))
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}

func (a *API) createDetectionAlerts(tx *gorm.DB, event store.AttackEvent) ([]store.Alert, error) {
	var hits []detection.Hit
	if len(bytes.TrimSpace(event.Detections)) == 0 {
		return nil, nil
	}
	if json.Unmarshal(event.Detections, &hits) != nil {
		return nil, errors.New("invalid event detections")
	}
	createdAlerts := make([]store.Alert, 0, len(hits))
	for _, hit := range hits {
		if hit.Stage != "server" {
			continue
		}
		fingerprintSum := sha256.Sum256([]byte("detection|" + event.EventID + "|" + hit.RuleKey))
		verification := "Server 已二次确认"
		if hit.Confirmed {
			verification = "Agent 命中，Server 已二次确认"
		}
		description := verification + "；来源 " + event.SrcIP + "，目标端口 " + strconv.Itoa(event.DstPort)
		if hit.Description != "" {
			description += "；" + hit.Description
		}
		fingerprint := hex.EncodeToString(fingerprintSum[:])
		alert := store.Alert{
			Base: store.Base{ID: stableUUID("detection-alert", fingerprint)}, EventID: event.EventID, RuleID: hit.RuleID, Fingerprint: fingerprint,
			Title: hit.Name, Level: hit.Severity, Status: "new", SourceIP: event.SrcIP, NodeID: event.NodeID,
			Service: event.Service, Description: description,
		}
		if alert.Level == "" {
			alert.Level = "medium"
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&alert)
		if result.Error != nil {
			return nil, fmt.Errorf("create detection alert: %w", result.Error)
		}
		if result.RowsAffected > 0 {
			if err := a.alerts.EnqueueWithDB(tx, alert, nil); err != nil {
				return nil, fmt.Errorf("enqueue detection alert: %w", err)
			}
			createdAlerts = append(createdAlerts, alert)
		}
	}
	return createdAlerts, nil
}
