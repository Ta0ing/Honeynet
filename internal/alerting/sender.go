package alerting

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/store"
)

type Sender struct {
	client *http.Client
	now    func() time.Time
}

func NewSender() *Sender {
	return &Sender{client: &http.Client{Timeout: 10 * time.Second}, now: time.Now}
}

func (s *Sender) Send(ctx context.Context, channel store.AlertChannel, message Message) error {
	if err := ValidateChannel(channel); err != nil {
		return err
	}
	switch channel.Type {
	case "webhook":
		return s.sendWebhook(ctx, channel, message)
	case "wecom", "dingtalk", "feishu":
		return s.sendRobot(ctx, channel, message)
	case "email":
		return sendEmail(ctx, channel, message)
	case "syslog":
		return sendSyslog(ctx, channel, message)
	default:
		return fmt.Errorf("unsupported alert channel type %q", channel.Type)
	}
}

func (s *Sender) sendWebhook(ctx context.Context, channel store.AlertChannel, message Message) error {
	var cfg WebhookConfig
	_ = json.Unmarshal(channel.Config, &cfg)
	body := message.JSON()
	headers := map[string]string{}
	for key, value := range cfg.Headers {
		headers[key] = value
	}
	if cfg.Secret != "" {
		timestamp := strconv.FormatInt(s.now().Unix(), 10)
		mac := hmac.New(sha256.New, []byte(cfg.Secret))
		_, _ = mac.Write([]byte(timestamp + "\n"))
		_, _ = mac.Write(body)
		headers["X-Honeynet-Timestamp"] = timestamp
		headers["X-Honeynet-Signature"] = "sha256=" + hex.EncodeToString(mac.Sum(nil))
	}
	_, err := s.postJSON(ctx, cfg.URL, body, headers)
	return err
}

func (s *Sender) sendRobot(ctx context.Context, channel store.AlertChannel, message Message) error {
	var cfg RobotConfig
	_ = json.Unmarshal(channel.Config, &cfg)
	webhookURL := cfg.WebhookURL
	var payload any
	switch channel.Type {
	case "wecom":
		payload = map[string]any{"msgtype": "markdown", "markdown": map[string]string{"content": message.Markdown()}}
	case "dingtalk":
		payload = map[string]any{"msgtype": "markdown", "markdown": map[string]string{"title": message.Alert.Title, "text": message.Markdown()}}
		if cfg.Secret != "" {
			timestamp := strconv.FormatInt(s.now().UnixMilli(), 10)
			mac := hmac.New(sha256.New, []byte(cfg.Secret))
			_, _ = mac.Write([]byte(timestamp + "\n" + cfg.Secret))
			parsed, _ := url.Parse(webhookURL)
			query := parsed.Query()
			query.Set("timestamp", timestamp)
			query.Set("sign", base64.StdEncoding.EncodeToString(mac.Sum(nil)))
			parsed.RawQuery = query.Encode()
			webhookURL = parsed.String()
		}
	case "feishu":
		payload = map[string]any{"msg_type": "text", "content": map[string]string{"text": message.PlainText()}}
		if cfg.Secret != "" {
			timestamp := strconv.FormatInt(s.now().Unix(), 10)
			stringToSign := timestamp + "\n" + cfg.Secret
			mac := hmac.New(sha256.New, []byte(stringToSign))
			payload.(map[string]any)["timestamp"] = timestamp
			payload.(map[string]any)["sign"] = base64.StdEncoding.EncodeToString(mac.Sum(nil))
		}
	}
	body, _ := json.Marshal(payload)
	responseBody, err := s.postJSON(ctx, webhookURL, body, nil)
	if err != nil {
		return err
	}
	return validateRobotResponse(channel.Type, responseBody)
}

func (s *Sender) postJSON(ctx context.Context, target string, body []byte, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("User-Agent", "Honeynet-Server/0.2")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	if readErr != nil {
		return nil, readErr
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("webhook returned HTTP %d: %s", response.StatusCode, truncate(string(data), 256))
	}
	return data, nil
}

func validateRobotResponse(channelType string, body []byte) error {
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil
	}
	var response map[string]any
	if json.Unmarshal(body, &response) != nil {
		return nil
	}
	var code float64
	var message string
	switch channelType {
	case "wecom", "dingtalk":
		code, _ = response["errcode"].(float64)
		message = fmt.Sprint(response["errmsg"])
	case "feishu":
		code, _ = response["code"].(float64)
		if code == 0 {
			code, _ = response["StatusCode"].(float64)
		}
		message = fmt.Sprint(response["msg"])
		if message == "<nil>" {
			message = fmt.Sprint(response["StatusMessage"])
		}
	}
	if code != 0 {
		return fmt.Errorf("%s robot rejected message: code=%v message=%s", channelType, code, message)
	}
	return nil
}

func sendSyslog(ctx context.Context, channel store.AlertChannel, message Message) error {
	var cfg SyslogConfig
	_ = json.Unmarshal(channel.Config, &cfg)
	network := cfg.Network
	if network == "" {
		network = "udp"
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var connection net.Conn
	var err error
	if network == "tls" {
		host, _, _ := net.SplitHostPort(cfg.Address)
		connection, err = (&tls.Dialer{NetDialer: dialer, Config: &tls.Config{MinVersion: tls.VersionTLS12, ServerName: host}}).DialContext(ctx, "tcp", cfg.Address)
	} else {
		connection, err = dialer.DialContext(ctx, network, cfg.Address)
	}
	if err != nil {
		return err
	}
	defer connection.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = connection.SetDeadline(deadline)
	} else {
		_ = connection.SetDeadline(time.Now().Add(10 * time.Second))
	}
	host, _ := os.Hostname()
	host = sanitizeSyslog(host)
	facility := cfg.Facility
	if facility == 0 {
		facility = 16
	}
	priority := facility*8 + syslogSeverity(message.Alert.Level)
	payload := strings.ReplaceAll(message.PlainText(), "\n", " | ")
	line := fmt.Sprintf("<%d>1 %s %s honeynet - ALERT - %s", priority, time.Now().UTC().Format(time.RFC3339Nano), host, payload)
	if network == "tcp" || network == "tls" {
		line = fmt.Sprintf("%d %s", len([]byte(line)), line)
	}
	_, err = io.WriteString(connection, line)
	return err
}

func syslogSeverity(level string) int {
	switch level {
	case "critical":
		return 2
	case "high":
		return 3
	case "medium":
		return 4
	case "low":
		return 5
	default:
		return 6
	}
}

func sanitizeSyslog(value string) string {
	value = strings.Map(func(r rune) rune {
		if r <= 32 || r == 127 {
			return '_'
		}
		return r
	}, value)
	if value == "" {
		return "-"
	}
	return value
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) > limit {
		return value[:limit] + "..."
	}
	return value
}
