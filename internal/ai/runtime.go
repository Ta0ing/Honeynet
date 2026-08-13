package ai

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTimeout = 45 * time.Second
	SecretMask     = "********"
)

// Runtime owns the active provider client. Configuration updates replace the
// client atomically, so in-flight requests finish with their original client
// while subsequent requests immediately use the new configuration.
type Runtime struct {
	mu     sync.RWMutex
	config Config
	client *Client
}

func NewRuntime(config Config) (*Runtime, error) {
	normalized, err := NormalizeConfig(config)
	if err != nil {
		return nil, err
	}
	return &Runtime{config: normalized, client: New(normalized)}, nil
}

func (r *Runtime) Replace(config Config) error {
	normalized, err := NormalizeConfig(config)
	if err != nil {
		return err
	}
	client := New(normalized)
	r.mu.Lock()
	r.config, r.client = normalized, client
	r.mu.Unlock()
	return nil
}

func (r *Runtime) Analyze(ctx context.Context, request Request) (Response, error) {
	r.mu.RLock()
	client := r.client
	r.mu.RUnlock()
	if client == nil {
		return Response{}, errors.New("AI 提供商尚未初始化")
	}
	return client.Analyze(ctx, request)
}

func (r *Runtime) Status() Status {
	r.mu.RLock()
	config := r.config
	r.mu.RUnlock()
	return statusFromConfig(config)
}

func (r *Runtime) ExecutionSettings() (time.Duration, bool) {
	r.mu.RLock()
	timeout, sendRawPacket := r.config.Timeout, r.config.SendRawPacket
	r.mu.RUnlock()
	return timeout, sendRawPacket
}

func (r *Runtime) SafeError(err error) string {
	if err == nil {
		return ""
	}
	r.mu.RLock()
	apiKey := r.config.APIKey
	r.mu.RUnlock()
	return redactSecrets(err.Error(), apiKey)
}

func statusFromConfig(config Config) Status {
	return Status{
		Enabled:        config.Enabled,
		Configured:     config.BaseURL != "" && config.APIKey != "" && config.Model != "",
		HasAPIKey:      config.APIKey != "",
		Provider:       config.Provider,
		BaseURL:        config.BaseURL,
		Model:          config.Model,
		TimeoutSeconds: int(config.Timeout / time.Second),
		SendRawPacket:  config.SendRawPacket,
	}
}

func NormalizeConfig(config Config) (Config, error) {
	config.Provider = strings.TrimSpace(config.Provider)
	if config.Provider == "" {
		config.Provider = "openai-compatible"
	}
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.APIKey = strings.TrimSpace(config.APIKey)
	config.Model = strings.TrimSpace(config.Model)
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}
	if config.Timeout < time.Second || config.Timeout > 5*time.Minute {
		return Config{}, errors.New("AI 请求超时必须在 1 秒到 5 分钟之间")
	}
	if len(config.Provider) > 64 || len(config.BaseURL) > 1024 || len(config.Model) > 128 || len(config.APIKey) > 8192 {
		return Config{}, errors.New("AI 提供商配置超过允许长度")
	}
	if config.BaseURL != "" {
		parsed, err := url.Parse(config.BaseURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return Config{}, errors.New("AI Base URL 必须是有效的 HTTP 或 HTTPS 地址")
		}
		if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return Config{}, errors.New("AI Base URL 不允许包含账号、查询参数或片段")
		}
		if parsed.Scheme == "http" && !isLoopbackAIHost(parsed.Hostname()) {
			return Config{}, errors.New("远程 AI Base URL 必须使用 HTTPS；HTTP 仅允许本机回环地址")
		}
	}
	if config.Enabled && (config.BaseURL == "" || config.APIKey == "" || config.Model == "") {
		return Config{}, errors.New("启用 AI 前必须配置 Base URL、API Key 和模型")
	}
	return config, nil
}

func isLoopbackAIHost(host string) bool {
	host = strings.TrimSpace(strings.TrimSuffix(host, "."))
	if strings.EqualFold(host, "localhost") {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}
