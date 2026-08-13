package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	aimodule "github.com/honeynet/honeynet/internal/ai"
)

type aiSettingsRequest struct {
	Enabled        *bool   `json:"enabled"`
	Provider       *string `json:"provider"`
	BaseURL        *string `json:"base_url"`
	APIKey         *string `json:"api_key"`
	Model          *string `json:"model"`
	TimeoutSeconds *int    `json:"timeout_seconds"`
	SendRawPacket  *bool   `json:"send_raw_packet"`
	ClearAPIKey    bool    `json:"clear_api_key"`
}

func (a *API) getAIConfig(c *gin.Context) {
	settings, err := a.aiSettings.Load(c.Request.Context())
	if err != nil {
		fail(c, http.StatusInternalServerError, "AI_CONFIG_READ_FAILED", "读取 AI 配置失败")
		return
	}
	ok(c, settings.View())
}

func (a *API) updateAIConfig(c *gin.Context) {
	var req aiSettingsRequest
	if c.ShouldBindJSON(&req) != nil {
		fail(c, http.StatusBadRequest, "INVALID_ARGUMENT", "AI 配置格式错误")
		return
	}
	settings, err := a.aiSettings.Update(c.Request.Context(), aimodule.SettingsUpdate{
		Enabled: req.Enabled, Provider: req.Provider, BaseURL: req.BaseURL, APIKey: req.APIKey,
		Model: req.Model, TimeoutSeconds: req.TimeoutSeconds, SendRawPacket: req.SendRawPacket,
		ClearAPIKey: req.ClearAPIKey,
	})
	if err != nil {
		if errors.Is(err, aimodule.ErrInvalidSettings) {
			message := strings.TrimPrefix(err.Error(), aimodule.ErrInvalidSettings.Error()+": ")
			fail(c, http.StatusBadRequest, "INVALID_AI_CONFIG", message)
			return
		}
		fail(c, http.StatusInternalServerError, "AI_CONFIG_SAVE_FAILED", "保存 AI 配置失败")
		return
	}
	if err := a.ai.Replace(settings.Config); err != nil {
		fail(c, http.StatusInternalServerError, "AI_CONFIG_APPLY_FAILED", "AI 配置已保存，但运行时更新失败")
		return
	}
	ok(c, settings.View())
}

func (a *API) testAIConfig(c *gin.Context) {
	status := a.ai.Status()
	if !status.Configured {
		fail(c, http.StatusBadRequest, "AI_NOT_CONFIGURED", "请先保存完整的 AI 提供商配置")
		return
	}
	timeout, _ := a.ai.ExecutionSettings()
	if timeout > 30*time.Second {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	response, err := a.ai.Analyze(ctx, aimodule.Request{
		Task:     "连接测试：请仅返回简短 JSON，summary 字段值为连接成功。",
		Evidence: map[string]any{"test": true},
	})
	if err != nil {
		fail(c, http.StatusBadGateway, "AI_CONNECTION_FAILED", truncateAIError(a.ai.SafeError(err)))
		return
	}
	ok(c, gin.H{"success": true, "provider": response.Provider, "model": response.Model})
}
