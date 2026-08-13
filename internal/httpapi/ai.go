package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	aimodule "github.com/honeynet/honeynet/internal/ai"
	"github.com/honeynet/honeynet/internal/analytics"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

func (a *API) aiStatus(c *gin.Context) {
	status := a.ai.Status()
	ok(c, gin.H{
		"enabled": status.Enabled, "configured": status.Configured, "has_api_key": status.HasAPIKey,
		"provider": status.Provider, "base_url": status.BaseURL, "model": status.Model,
		"timeout_seconds": status.TimeoutSeconds, "send_raw_packet": status.SendRawPacket,
		"agent_mode": "auditable-harness", "agent_ready": status.Enabled && status.Configured,
		"agent_stage": "rule-improvement", "tool_count": len(a.aiAgent.Tools()),
	})
}

func (a *API) aiAgentCapabilities(c *gin.Context) {
	status := a.ai.Status()
	ok(c, gin.H{
		"mode": "auditable-harness", "stage": "rule-improvement", "ready": status.Enabled && status.Configured,
		"tools": a.aiAgent.Tools(), "max_steps": 8,
		"harness":      gin.H{"persistent_runs": true, "traceable": true, "tool_policy": "allowlisted-read-only", "tool_budget": 3, "sample_limit": harnessMaximumSamples, "minimum_samples": harnessMinimumSamples, "minimum_per_class": harnessMinimumPerClass, "minimum_evaluation_per_class": harnessMinimumEvalClass, "minimum_precision": harnessMinimumPrecision, "minimum_recall": harnessMinimumRecall, "maximum_false_positive_rate": harnessMaximumFPR, "split_strategy": "sha256-stratified-holdout-v1", "approval_required": true, "rollback_supported": true},
		"guardrails":   []string{"攻击证据按不可信数据处理", "只允许显式注册的只读工具与固定预算", "候选规则必须通过静态校验和历史回放", "模型不能直接修改或发布检测规则", "管理员审批后才允许发布，版本退化可回滚"},
		"capabilities": []string{"持久化目标与环境快照", "可审计工具轨迹", "规则优化提案", "离线评估与评分门槛", "人工审批发布", "误报反馈与版本回滚"},
	})
}

func (a *API) runAIAgent(c *gin.Context) {
	status := a.ai.Status()
	if !status.Enabled || !status.Configured {
		fail(c, http.StatusConflict, "AI_DISABLED", "AI Agent 尚未启用或模型配置不完整")
		return
	}
	var request aimodule.AgentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_ARGUMENT", "AI Agent 任务格式错误")
		return
	}
	timeout, _ := a.ai.ExecutionSettings()
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	result, err := a.aiAgent.Run(ctx, request)
	if err != nil {
		fail(c, http.StatusBadGateway, "AI_AGENT_FAILED", truncateAIError(a.ai.SafeError(err)))
		return
	}
	ok(c, result)
}

func (a *API) listAIAnalyses(c *gin.Context) {
	p, size := page(c)
	query := a.db.Model(&store.AIAnalysis{})
	if value := strings.TrimSpace(c.Query("target_type")); value != "" {
		query = query.Where("target_type = ?", value)
	}
	if value := strings.TrimSpace(c.Query("target_id")); value != "" {
		query = query.Where("target_id = ?", value)
	}
	var total int64
	query.Count(&total)
	var items []store.AIAnalysis
	query.Order("created_at DESC").Offset((p - 1) * size).Limit(size).Find(&items)
	ok(c, pageResult(items, total, p, size))
}

func (a *API) analyzeEvent(c *gin.Context) {
	if !a.ai.Status().Enabled {
		fail(c, http.StatusConflict, "AI_DISABLED", "AI 功能未启用，请先在 Server 配置中接入 DeepSeek、GLM 或其他 OpenAI 兼容模型")
		return
	}
	var event store.AttackEvent
	if a.analytics != nil {
		stored, err := a.analytics.Get(c.Request.Context(), c.Param("id"))
		if err != nil {
			if errors.Is(err, analytics.ErrNotFound) {
				fail(c, http.StatusNotFound, "EVENT_NOT_FOUND", "事件不存在")
			} else {
				fail(c, http.StatusServiceUnavailable, "ANALYTICS_UNAVAILABLE", "安全事件分析服务查询失败")
			}
			return
		}
		event = analytics.ToAttackEvent(stored)
	} else if a.db.First(&event, "event_id = ?", c.Param("id")).Error != nil {
		fail(c, http.StatusNotFound, "EVENT_NOT_FOUND", "事件不存在")
		return
	}
	// Provider evidence never receives captured credentials through normalized
	// fields. Administrators may separately opt in to RawPacket forwarding in
	// AI settings; that explicit high-risk option does not weaken this baseline.
	safeEvent := redactAttackEvent(event)
	payload := map[string]any{}
	_ = json.Unmarshal(safeEvent.Payload, &payload)
	evidence := map[string]any{
		"event_id": event.EventID, "time": event.Timestamp, "event_type": event.EventType, "service": event.Service,
		"source":      map[string]any{"ip": event.SrcIP, "port": event.SrcPort, "geo": event.Geo, "asn": event.ASN},
		"destination": map[string]any{"ip": event.DstIP, "port": event.DstPort}, "payload": payload,
		"detections": json.RawMessage(event.Detections), "tags": json.RawMessage(event.Tags),
	}
	_, sendRawPacket := a.ai.ExecutionSettings()
	if sendRawPacket && event.RawPacket != "" {
		evidence["raw_packet"] = event.RawPacket
	}
	a.runAIAnalysis(c, "event", event.EventID, "event-analysis", "分析该蜜罐攻击事件，判断攻击意图、漏洞利用可能性、TTP、IOC 与处置优先级。", evidence)
}

func (a *API) analyzeAttacker(c *gin.Context) {
	if !a.ai.Status().Enabled {
		fail(c, http.StatusConflict, "AI_DISABLED", "AI 功能未启用，请先在 Server 配置中接入 DeepSeek、GLM 或其他 OpenAI 兼容模型")
		return
	}
	ip := strings.TrimSpace(c.Param("ip"))
	if ip == "" || len(ip) > 64 {
		fail(c, http.StatusBadRequest, "INVALID_ARGUMENT", "攻击者 IP 无效")
		return
	}
	var events []store.AttackEvent
	if a.analytics != nil {
		evidence, err := a.analytics.AttackerEvidence(c.Request.Context(), ip, time.Time{}, time.Now(), 100)
		if err == nil {
			events = make([]store.AttackEvent, 0, len(evidence.Events))
			for _, stored := range evidence.Events {
				events = append(events, analytics.ToAttackEvent(stored))
			}
		}
	} else {
		_ = a.db.Where("src_ip = ?", ip).Order("ts DESC").Limit(100).Find(&events).Error
	}
	if len(events) == 0 {
		fail(c, http.StatusNotFound, "ATTACKER_NOT_FOUND", "没有找到该攻击来源的事件")
		return
	}
	services, eventTypes, detections := map[string]int{}, map[string]int{}, map[string]int{}
	firstSeen, lastSeen := events[0].Timestamp, events[0].Timestamp
	for _, event := range events {
		services[event.Service]++
		eventTypes[event.EventType]++
		if event.Timestamp.Before(firstSeen) {
			firstSeen = event.Timestamp
		}
		if event.Timestamp.After(lastSeen) {
			lastSeen = event.Timestamp
		}
		var hits []struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(event.Detections, &hits) == nil {
			for _, hit := range hits {
				detections[hit.Name]++
			}
		}
	}
	evidence := map[string]any{
		"source_ip": ip, "geo": events[0].Geo, "asn": events[0].ASN, "sample_size": len(events),
		"first_seen": firstSeen, "last_seen": lastSeen, "services": sortedCounts(services),
		"event_types": sortedCounts(eventTypes), "detections": sortedCounts(detections),
	}
	a.runAIAnalysis(c, "attacker", ip, "attacker-profile", "基于蜜罐时间线生成攻击者画像，分析能力、偏好目标、可能阶段、风险等级和后续监测建议；不要把地理位置直接等同于攻击者真实归属。", evidence)
}

func (a *API) runAIAnalysis(c *gin.Context, targetType, targetID, kind, task string, evidence map[string]any) {
	evidenceData, _ := json.Marshal(evidence)
	promptHash := sha256.Sum256(append([]byte(task), evidenceData...))
	item := store.AIAnalysis{
		Base: store.NewBase(), TargetType: targetType, TargetID: targetID, Kind: kind, Status: "running",
		Provider: a.ai.Status().Provider, Model: a.ai.Status().Model, Result: datatypes.JSON("{}"), PromptHash: hex.EncodeToString(promptHash[:]),
	}
	if a.db.Create(&item).Error != nil {
		fail(c, http.StatusInternalServerError, "CREATE_FAILED", "创建 AI 分析任务失败")
		return
	}
	timeout, _ := a.ai.ExecutionSettings()
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	response, err := a.ai.Analyze(ctx, aimodule.Request{Task: task, Evidence: evidence})
	if err != nil {
		item.Status, item.Error = "failed", truncateAIError(a.ai.SafeError(err))
		a.db.Model(&item).Updates(map[string]any{"status": item.Status, "error": item.Error})
		fail(c, http.StatusBadGateway, "AI_ANALYSIS_FAILED", item.Error)
		return
	}
	result := response.JSON
	if len(result) == 0 {
		result, _ = json.Marshal(map[string]any{"content": response.Content})
	}
	summary := response.Content
	var structured map[string]any
	if json.Unmarshal(result, &structured) == nil {
		if value, ok := structured["summary"].(string); ok && strings.TrimSpace(value) != "" {
			summary = value
		}
	}
	item.Status, item.Provider, item.Model, item.Summary, item.Result = "completed", response.Provider, response.Model, summary, datatypes.JSON(result)
	a.db.Model(&item).Updates(map[string]any{"status": item.Status, "provider": item.Provider, "model": item.Model, "summary": item.Summary, "result": item.Result, "error": ""})
	// Realtime fan-out is shared by every authenticated role. Never publish
	// model prose/result here because a provider may echo sensitive evidence.
	a.hub.Broadcast("ai.analysis", gin.H{
		"id": item.ID, "target_type": item.TargetType, "target_id": item.TargetID,
		"kind": item.Kind, "status": item.Status, "created_at": item.CreatedAt,
	})
	ok(c, item)
}

type countValue struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func sortedCounts(values map[string]int) []countValue {
	items := make([]countValue, 0, len(values))
	for name, count := range values {
		if name != "" {
			items = append(items, countValue{Name: name, Count: count})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	return items
}

func truncateAIError(value string) string {
	if len(value) > 1000 {
		return value[:1000]
	}
	return value
}
