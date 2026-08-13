package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	aimodule "github.com/honeynet/honeynet/internal/ai"
	"github.com/honeynet/honeynet/internal/analytics"
	"github.com/honeynet/honeynet/internal/store"
)

type aiRecentEventsTool struct{ analytics analytics.Store }

func (aiRecentEventsTool) Name() string { return "recent_security_events" }
func (aiRecentEventsTool) Description() string {
	return "按时间范围、节点和攻击来源查询最近安全事件；返回归一化元数据，不返回原始请求包或凭据"
}
func (t aiRecentEventsTool) Execute(ctx context.Context, raw json.RawMessage) (any, error) {
	if t.analytics == nil {
		return nil, errors.New("安全事件分析存储尚未启用")
	}
	input := struct {
		SourceIP string `json:"source_ip"`
		NodeID   string `json:"node_id"`
		Hours    int    `json:"hours"`
		Limit    int    `json:"limit"`
	}{Hours: 24, Limit: 25}
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, errors.New("invalid recent events tool input")
		}
	}
	input.SourceIP, input.NodeID = strings.TrimSpace(input.SourceIP), strings.TrimSpace(input.NodeID)
	if input.SourceIP != "" && net.ParseIP(input.SourceIP) == nil {
		return nil, errors.New("source_ip must be a valid IPv4 or IPv6 address")
	}
	if len(input.NodeID) > 64 {
		return nil, errors.New("node_id is too long")
	}
	if input.Hours < 1 || input.Hours > 24*30 {
		return nil, errors.New("hours must be between 1 and 720")
	}
	if input.Limit < 1 || input.Limit > 100 {
		return nil, errors.New("limit must be between 1 and 100")
	}
	page, err := t.analytics.List(ctx, analytics.EventFilter{
		ExactSourceIP: input.SourceIP, NodeID: input.NodeID, From: time.Now().Add(-time.Duration(input.Hours) * time.Hour),
		To: time.Now(), Limit: input.Limit, CursorMode: true,
	})
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(page.Items))
	for _, event := range page.Items {
		items = append(items, map[string]any{
			"event_id": event.EventID, "time": event.EventTime, "node_id": event.NodeID, "service": event.Service,
			"event_type": event.EventType, "source_ip": event.SourceIP, "target_ip": event.TargetIP,
			"target_port": event.TargetPort, "geo": event.Geo, "asn": event.ASN, "detections": event.Detections,
		})
	}
	return map[string]any{"items": items, "returned": len(items), "has_more": page.HasMore}, nil
}

type aiAttackerEvidenceTool struct{ analytics analytics.Store }

func (aiAttackerEvidenceTool) Name() string { return "attacker_timeline" }
func (aiAttackerEvidenceTool) Description() string {
	return "按IPv4或IPv6攻击来源聚合首次/最近活动、服务偏好和事件类型，不返回凭据或原始报文"
}
func (t aiAttackerEvidenceTool) Execute(ctx context.Context, raw json.RawMessage) (any, error) {
	if t.analytics == nil {
		return nil, errors.New("安全事件分析存储尚未启用")
	}
	var input struct {
		SourceIP string `json:"source_ip"`
		Days     int    `json:"days"`
	}
	input.Days = 30
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, errors.New("invalid attacker timeline tool input")
	}
	input.SourceIP = strings.TrimSpace(input.SourceIP)
	if net.ParseIP(input.SourceIP) == nil {
		return nil, errors.New("source_ip must be a valid IPv4 or IPv6 address")
	}
	if input.Days < 1 || input.Days > 365 {
		return nil, errors.New("days must be between 1 and 365")
	}
	evidence, err := t.analytics.AttackerEvidence(ctx, input.SourceIP, time.Now().AddDate(0, 0, -input.Days), time.Now(), 50)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"source_ip": evidence.SourceIP, "geo": evidence.Geo, "asn": evidence.ASN, "count": evidence.Count,
		"first_seen": evidence.FirstSeen, "last_seen": evidence.LastSeen, "services": evidence.Services,
		"event_types": evidence.EventTypes,
	}, nil
}

type aiRuleSummaryTool struct{ api *API }

func (aiRuleSummaryTool) Name() string { return "detection_rule_summary" }
func (aiRuleSummaryTool) Description() string {
	return "读取当前启用检测规则的名称、级别、来源和版本摘要，用于解释规则命中"
}
func (t aiRuleSummaryTool) Execute(ctx context.Context, raw json.RawMessage) (any, error) {
	var input struct {
		RuleKey string `json:"rule_key"`
	}
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, errors.New("invalid rule summary tool input")
		}
	}
	input.RuleKey = strings.TrimSpace(input.RuleKey)
	if len(input.RuleKey) > 160 {
		return nil, errors.New("rule_key is too long")
	}
	query := t.api.db.WithContext(ctx).Model(&store.DetectionRule{}).Where("enabled = ?", true)
	if input.RuleKey != "" {
		query = query.Where("rule_key = ?", input.RuleKey)
	}
	var rules []store.DetectionRule
	if err := query.Order("revision DESC").Limit(100).Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("query detection rules: %w", err)
	}
	items := make([]map[string]any, 0, len(rules))
	for _, rule := range rules {
		items = append(items, map[string]any{"rule_key": rule.RuleKey, "name": rule.Name, "severity": rule.Severity, "source": rule.Source, "revision_text": fmt.Sprint(rule.Revision)})
	}
	return map[string]any{"items": items, "count": len(items)}, nil
}

func (a *API) newAIAgent() (*aimodule.Agent, error) {
	tools := []aimodule.Tool{aiRuleSummaryTool{api: a}}
	if a.analytics != nil {
		tools = append(tools, aiRecentEventsTool{analytics: a.analytics}, aiAttackerEvidenceTool{analytics: a.analytics})
	}
	return aimodule.NewAgent(a.ai, tools...)
}
