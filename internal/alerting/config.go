package alerting

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strings"

	"github.com/honeynet/honeynet/internal/store"
)

const Mask = "********"

type WebhookConfig struct {
	URL     string            `json:"url"`
	Secret  string            `json:"secret,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type RobotConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret,omitempty"`
}

type EmailConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	TLSMode  string   `json:"tls_mode,omitempty"`
}

type SyslogConfig struct {
	Address  string `json:"address"`
	Network  string `json:"network,omitempty"`
	Facility int    `json:"facility,omitempty"`
}

func SupportedTypes() []string {
	return []string{"webhook", "email", "syslog", "wecom", "dingtalk", "feishu"}
}

func ValidateChannel(channel store.AlertChannel) error {
	switch channel.Type {
	case "webhook":
		var cfg WebhookConfig
		if err := decode(channel.Config, &cfg); err != nil {
			return err
		}
		return validateHTTPURL(cfg.URL)
	case "wecom", "dingtalk", "feishu":
		var cfg RobotConfig
		if err := decode(channel.Config, &cfg); err != nil {
			return err
		}
		return validateHTTPURL(cfg.WebhookURL)
	case "email":
		var cfg EmailConfig
		if err := decode(channel.Config, &cfg); err != nil {
			return err
		}
		if strings.TrimSpace(cfg.Host) == "" || cfg.Port < 1 || cfg.Port > 65535 {
			return errors.New("email host and valid port are required")
		}
		if _, err := mail.ParseAddress(cfg.From); err != nil {
			return fmt.Errorf("invalid email from address: %w", err)
		}
		if len(cfg.To) == 0 {
			return errors.New("at least one email recipient is required")
		}
		for _, recipient := range cfg.To {
			if _, err := mail.ParseAddress(recipient); err != nil {
				return fmt.Errorf("invalid email recipient: %w", err)
			}
		}
		if cfg.TLSMode != "" && cfg.TLSMode != "starttls" && cfg.TLSMode != "implicit" && cfg.TLSMode != "none" {
			return errors.New("email tls_mode must be starttls, implicit or none")
		}
		return nil
	case "syslog":
		var cfg SyslogConfig
		if err := decode(channel.Config, &cfg); err != nil {
			return err
		}
		if _, _, err := net.SplitHostPort(cfg.Address); err != nil {
			return fmt.Errorf("syslog address must be host:port: %w", err)
		}
		if cfg.Network != "" && cfg.Network != "udp" && cfg.Network != "tcp" && cfg.Network != "tls" {
			return errors.New("syslog network must be udp, tcp or tls")
		}
		if cfg.Facility < 0 || cfg.Facility > 23 {
			return errors.New("syslog facility must be between 0 and 23")
		}
		return nil
	default:
		return fmt.Errorf("unsupported alert channel type %q", channel.Type)
	}
}

func RedactedConfig(channelType string, raw json.RawMessage) json.RawMessage {
	values := map[string]any{}
	if json.Unmarshal(raw, &values) != nil {
		return json.RawMessage("{}")
	}
	for _, key := range secretKeys(channelType) {
		if value, exists := values[key]; exists && value != nil && fmt.Sprint(value) != "" {
			values[key] = Mask
		}
	}
	if headers, ok := values["headers"].(map[string]any); ok {
		for key := range headers {
			headers[key] = Mask
		}
	}
	data, _ := json.Marshal(values)
	return data
}

func MergeConfig(channelType string, existing, incoming json.RawMessage) (json.RawMessage, error) {
	current := map[string]any{}
	next := map[string]any{}
	if len(existing) > 0 && json.Unmarshal(existing, &current) != nil {
		return nil, errors.New("stored channel config is invalid")
	}
	if len(incoming) == 0 || json.Unmarshal(incoming, &next) != nil {
		return nil, errors.New("channel config must be a JSON object")
	}
	secrets := map[string]bool{}
	for _, key := range secretKeys(channelType) {
		secrets[key] = true
	}
	for key, value := range next {
		if secrets[key] && (value == nil || strings.TrimSpace(fmt.Sprint(value)) == "" || fmt.Sprint(value) == Mask) {
			continue
		}
		if key == "headers" {
			incomingHeaders, _ := value.(map[string]any)
			currentHeaders, _ := current[key].(map[string]any)
			if currentHeaders == nil {
				currentHeaders = map[string]any{}
			}
			for header, headerValue := range incomingHeaders {
				if fmt.Sprint(headerValue) != Mask && strings.TrimSpace(fmt.Sprint(headerValue)) != "" {
					currentHeaders[header] = headerValue
				}
			}
			current[key] = currentHeaders
			continue
		}
		current[key] = value
	}
	data, err := json.Marshal(current)
	return data, err
}

func decode(raw []byte, target any) error {
	if len(raw) == 0 || string(raw) == "null" || json.Unmarshal(raw, target) != nil {
		return errors.New("channel config is invalid")
	}
	return nil
}

func validateHTTPURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("a valid http or https webhook URL is required")
	}
	return nil
}

func secretKeys(channelType string) []string {
	switch channelType {
	case "webhook":
		return []string{"url", "secret"}
	case "wecom", "dingtalk", "feishu":
		return []string{"webhook_url", "secret"}
	case "email":
		return []string{"password"}
	default:
		return nil
	}
}
