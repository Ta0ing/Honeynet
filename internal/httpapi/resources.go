package httpapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/decoyconfig"
	"github.com/honeynet/honeynet/internal/store"
	"github.com/honeynet/honeynet/internal/webtemplate"
	"gorm.io/datatypes"
)

func validateTemplate(content string) error {
	_, err := webtemplate.Parse(content)
	return err
}

func (a *API) listPotTemplates(c *gin.Context) {
	p, s := page(c)
	q := a.db.Model(&store.PotTemplate{})
	var total int64
	q.Count(&total)
	var items []store.PotTemplate
	q.Select("pot_templates.*, (SELECT COUNT(*) FROM pot_instances WHERE pot_instances.template_id = pot_templates.id AND pot_instances.deleted_at IS NULL) AS instance_count").Order("updated_at DESC").Offset((p - 1) * s).Limit(s).Find(&items)
	ok(c, pageResult(items, total, p, s))
}

func (a *API) createPotTemplate(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
		YAML string `json:"yaml"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_TEMPLATE", "模板请求格式无效")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	templateErr := validateTemplate(req.YAML)
	if req.Name == "" || len(req.Name) > 128 || templateErr != nil {
		fail(c, 400, "INVALID_TEMPLATE", "模板必须包含 name、有效 listen.port 及至少一个 pages 路由")
		return
	}
	item := store.PotTemplate{Base: store.NewBase(), Name: req.Name, YAML: req.YAML, Version: 1, CreatedBy: currentUser(c).ID}
	if a.db.Create(&item).Error != nil {
		fail(c, 409, "TEMPLATE_EXISTS", "模板名称已存在")
		return
	}
	created(c, item)
}

func (a *API) updatePotTemplate(c *gin.Context) {
	var item store.PotTemplate
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "TEMPLATE_NOT_FOUND", "模板不存在")
		return
	}
	var req struct {
		Name string `json:"name"`
		YAML string `json:"yaml"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_TEMPLATE", "模板请求格式无效")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 128 || validateTemplate(req.YAML) != nil {
		fail(c, 400, "INVALID_TEMPLATE", "模板格式无效")
		return
	}
	if err := a.db.Model(&item).Updates(map[string]any{"name": req.Name, "yaml": req.YAML, "version": item.Version + 1}).Error; err != nil {
		fail(c, 409, "TEMPLATE_EXISTS", "模板名称已存在")
		return
	}
	a.db.First(&item, "id = ?", item.ID)
	a.reloadTemplatePots(item.ID)
	ok(c, item)
}

func (a *API) deletePotTemplate(c *gin.Context) {
	var count int64
	a.db.Model(&store.PotInstance{}).Where("template_id = ?", c.Param("id")).Count(&count)
	if count > 0 {
		fail(c, http.StatusConflict, "TEMPLATE_IN_USE", "模板仍被蜜罐实例引用，请先删除相关实例")
		return
	}
	result := a.db.Delete(&store.PotTemplate{}, "id = ?", c.Param("id"))
	if result.RowsAffected == 0 {
		fail(c, 404, "TEMPLATE_NOT_FOUND", "模板不存在")
		return
	}
	okEmpty(c)
}

func (a *API) reloadTemplatePots(templateID string) {
	a.db.Model(&store.PotInstance{}).Where("template_id = ?", templateID).Update("actual_status", "pending")
	var nodeIDs []string
	a.db.Model(&store.PotInstance{}).Where("template_id = ?", templateID).Distinct().Pluck("node_id", &nodeIDs)
	for _, nodeID := range nodeIDs {
		_ = a.agents.SendPotApply(nodeID)
	}
}

type decoyRequest struct {
	NodeID string          `json:"node_id"`
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
	Status string          `json:"status"`
}

func validDecoyType(value string) bool {
	return value == "credential" || value == "file" || value == "network"
}

func normalizeDecoyConfig(kind string, raw json.RawMessage) (datatypes.JSON, error) {
	spec, err := decoyconfig.Parse(kind, raw)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(decoyconfig.Canonical(spec)), nil
}

func (a *API) listDecoys(c *gin.Context) {
	p, s := page(c)
	q := a.db.Model(&store.Decoy{})
	if nodeID := c.Query("node_id"); nodeID != "" {
		q = q.Where("node_id = ?", nodeID)
	}
	var total int64
	q.Count(&total)
	var items []store.Decoy
	q.Preload("Node").Order("created_at DESC").Offset((p - 1) * s).Limit(s).Find(&items)
	ok(c, pageResult(items, total, p, s))
}
func (a *API) createDecoy(c *gin.Context) {
	var req decoyRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, decoyconfig.MaxConfigBytes+4096)
	if c.ShouldBindJSON(&req) != nil || req.NodeID == "" || req.Name == "" || !validDecoyType(req.Type) {
		fail(c, 400, "INVALID_ARGUMENT", "节点、名称和有效蜜饵类型不能为空")
		return
	}
	var node store.Node
	if a.db.First(&node, "id = ?", req.NodeID).Error != nil {
		fail(c, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	if len(req.Config) == 0 {
		req.Config = json.RawMessage("{}")
	}
	normalized, err := normalizeDecoyConfig(req.Type, req.Config)
	if err != nil {
		fail(c, 400, "INVALID_DECOY_CONFIG", err.Error())
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 128 {
		fail(c, 400, "INVALID_ARGUMENT", "蜜饵名称不能超过 128 个字符")
		return
	}
	if req.Status == "" {
		req.Status = "enabled"
	}
	if req.Status != "enabled" && req.Status != "disabled" {
		fail(c, 400, "INVALID_ARGUMENT", "蜜饵状态必须为 enabled 或 disabled")
		return
	}
	item := store.Decoy{Base: store.NewBase(), NodeID: req.NodeID, Name: req.Name, Type: req.Type, Config: normalized, Status: req.Status, ActualStatus: "pending"}
	if a.db.Create(&item).Error != nil {
		fail(c, 500, "CREATE_FAILED", "创建蜜饵失败")
		return
	}
	a.db.Preload("Node").First(&item, "id = ?", item.ID)
	_ = a.agents.SendDecoyApply(item.NodeID)
	created(c, item)
}
func (a *API) updateDecoy(c *gin.Context) {
	var item store.Decoy
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "DECOY_NOT_FOUND", "蜜饵不存在")
		return
	}
	var req struct {
		Name   string          `json:"name"`
		Config json.RawMessage `json:"config"`
		Status string          `json:"status"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, decoyconfig.MaxConfigBytes+4096)
	if c.ShouldBindJSON(&req) != nil {
		fail(c, 400, "INVALID_ARGUMENT", "请求格式错误")
		return
	}
	updates := map[string]any{}
	if req.Name != "" {
		name := strings.TrimSpace(req.Name)
		if name == "" || len(name) > 128 {
			fail(c, 400, "INVALID_ARGUMENT", "蜜饵名称不能超过 128 个字符")
			return
		}
		updates["name"] = name
	}
	if len(req.Config) > 0 {
		normalized, err := normalizeDecoyConfig(item.Type, req.Config)
		if err != nil {
			fail(c, 400, "INVALID_DECOY_CONFIG", err.Error())
			return
		}
		updates["config"] = normalized
	}
	if req.Status != "" && req.Status != "enabled" && req.Status != "disabled" {
		fail(c, 400, "INVALID_ARGUMENT", "蜜饵状态必须为 enabled 或 disabled")
		return
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	updates["actual_status"] = "pending"
	updates["last_error"] = ""
	a.db.Model(&item).Updates(updates)
	a.db.Preload("Node").First(&item, "id = ?", item.ID)
	_ = a.agents.SendDecoyApply(item.NodeID)
	ok(c, item)
}
func (a *API) deleteDecoy(c *gin.Context) {
	var item store.Decoy
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "DECOY_NOT_FOUND", "蜜饵不存在")
		return
	}
	a.db.Delete(&item)
	_ = a.agents.SendDecoyApply(item.NodeID)
	okEmpty(c)
}

func (a *API) listIntelFeeds(c *gin.Context) {
	var items []store.IntelFeed
	a.db.Order("created_at DESC").Find(&items)
	ok(c, items)
}
func validateFeed(rawURL, feedType string) bool {
	if feedType != "csv" && feedType != "stix" && feedType != "taxii" {
		return false
	}
	parsed, err := url.Parse(rawURL)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
func (a *API) createIntelFeed(c *gin.Context) {
	var req struct {
		Name     string          `json:"name"`
		Type     string          `json:"type"`
		URL      string          `json:"url"`
		Schedule string          `json:"schedule"`
		Auth     json.RawMessage `json:"auth"`
		Enabled  *bool           `json:"enabled"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Name == "" || !validateFeed(req.URL, req.Type) {
		fail(c, 400, "INVALID_ARGUMENT", "Feed 名称、类型和 HTTP(S) 地址无效")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if len(req.Auth) == 0 {
		req.Auth = json.RawMessage("{}")
	}
	item := store.IntelFeed{Base: store.NewBase(), Name: req.Name, Type: req.Type, URL: req.URL, Schedule: req.Schedule, Auth: datatypes.JSON(req.Auth), Enabled: enabled}
	if a.db.Create(&item).Error != nil {
		fail(c, 500, "CREATE_FAILED", "创建订阅源失败")
		return
	}
	created(c, item)
}
func (a *API) updateIntelFeed(c *gin.Context) {
	var item store.IntelFeed
	if a.db.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "FEED_NOT_FOUND", "订阅源不存在")
		return
	}
	var req struct {
		Name     string          `json:"name"`
		Type     string          `json:"type"`
		URL      string          `json:"url"`
		Schedule string          `json:"schedule"`
		Auth     json.RawMessage `json:"auth"`
		Enabled  *bool           `json:"enabled"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Name == "" || !validateFeed(req.URL, req.Type) {
		fail(c, 400, "INVALID_ARGUMENT", "订阅源配置无效")
		return
	}
	updates := map[string]any{"name": req.Name, "type": req.Type, "url": req.URL, "schedule": req.Schedule}
	if len(req.Auth) > 0 {
		updates["auth"] = req.Auth
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	a.db.Model(&item).Updates(updates)
	a.db.First(&item, "id = ?", item.ID)
	ok(c, item)
}
func (a *API) deleteIntelFeed(c *gin.Context) {
	result := a.db.Delete(&store.IntelFeed{}, "id = ?", c.Param("id"))
	if result.RowsAffected == 0 {
		fail(c, 404, "FEED_NOT_FOUND", "订阅源不存在")
		return
	}
	okEmpty(c)
}
