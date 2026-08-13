package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/honeynet/honeynet/internal/agent/protocol"
	"github.com/honeynet/honeynet/internal/agent/sense"
	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func ensureNodeSenseConfig(db *gorm.DB, nodeID string) (store.NodeSenseConfig, error) {
	return store.EnsureNodeSenseConfig(db, nodeID)
}

func nodeSenseProtocolConfig(item store.NodeSenseConfig) (protocol.SenseConfig, error) {
	config := protocol.SenseConfig{
		Enabled: item.Enabled, Interface: item.Interface,
		TCPEnabled: item.TCPEnabled, UDPEnabled: item.UDPEnabled,
		DistinctPorts: item.DistinctPorts, WindowSeconds: item.WindowSeconds,
		CooldownSeconds: item.CooldownSeconds,
	}
	if len(item.ExcludedPorts) > 0 {
		if err := json.Unmarshal(item.ExcludedPorts, &config.ExcludedPorts); err != nil {
			return config, err
		}
	}
	if len(item.IgnoredCIDRs) > 0 {
		if err := json.Unmarshal(item.IgnoredCIDRs, &config.IgnoredCIDRs); err != nil {
			return config, err
		}
	}
	return sense.NormalizeConfig(config)
}

func (a *API) getNodeSense(c *gin.Context) {
	var node store.Node
	if a.db.First(&node, "id = ?", c.Param("id")).Error != nil {
		fail(c, http.StatusNotFound, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	item, err := ensureNodeSenseConfig(a.db, node.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "QUERY_FAILED", "读取节点感知配置失败")
		return
	}
	ok(c, item)
}

func (a *API) updateNodeSense(c *gin.Context) {
	var node store.Node
	if a.db.First(&node, "id = ?", c.Param("id")).Error != nil {
		fail(c, http.StatusNotFound, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	var request protocol.SenseConfig
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
	if err := c.ShouldBindJSON(&request); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_ARGUMENT", "感知配置格式错误")
		return
	}
	config, err := sense.NormalizeConfig(request)
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID_SENSE_CONFIG", err.Error())
		return
	}
	excluded, _ := json.Marshal(config.ExcludedPorts)
	ignored, _ := json.Marshal(config.IgnoredCIDRs)
	item, err := ensureNodeSenseConfig(a.db, node.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "初始化节点感知配置失败")
		return
	}
	updates := map[string]any{
		"enabled": config.Enabled, "interface": config.Interface,
		"tcp_enabled": config.TCPEnabled, "udp_enabled": config.UDPEnabled,
		"distinct_ports": config.DistinctPorts, "window_seconds": config.WindowSeconds,
		"cooldown_seconds": config.CooldownSeconds,
		"excluded_ports":   datatypes.JSON(excluded), "ignored_cidrs": datatypes.JSON(ignored),
		"actual_status": "pending", "last_error": "", "started_at": nil,
	}
	if err := a.db.Model(&item).Updates(updates).Error; err != nil {
		fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "保存节点感知配置失败")
		return
	}
	if err := a.agents.SendSenseApply(node.ID); err != nil && !errors.Is(err, errNodeOffline) {
		fail(c, http.StatusInternalServerError, "DISPATCH_FAILED", "感知配置已保存，但下发失败")
		return
	}
	a.db.First(&item, "id = ?", item.ID)
	ok(c, item)
}
