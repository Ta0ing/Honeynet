package alerting

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/store"
)

type Message struct {
	Alert      store.Alert `json:"alert"`
	ConsoleURL string      `json:"console_url,omitempty"`
	Test       bool        `json:"test,omitempty"`
}

func TestMessage(consoleURL string) Message {
	now := time.Now()
	return Message{Test: true, ConsoleURL: consoleURL, Alert: store.Alert{
		Base:    store.Base{ID: "test", CreatedAt: now, UpdatedAt: now},
		EventID: "test-event", Title: "Honeynet 告警通道测试", Level: "info", Status: "new",
		SourceIP: "192.0.2.10", Service: "system", Description: "如果您收到此消息，说明告警通道配置正确。",
	}}
}

func (m Message) PlainText() string {
	lines := []string{
		fmt.Sprintf("[%s] %s", strings.ToUpper(m.Alert.Level), m.Alert.Title),
		fmt.Sprintf("来源 IP: %s", valueOr(m.Alert.SourceIP, "-")),
		fmt.Sprintf("蜜罐服务: %s", valueOr(m.Alert.Service, "-")),
		fmt.Sprintf("详情: %s", valueOr(m.Alert.Description, "-")),
		fmt.Sprintf("时间: %s", m.Alert.CreatedAt.Local().Format("2006-01-02 15:04:05 MST")),
	}
	if m.ConsoleURL != "" {
		lines = append(lines, "控制台: "+strings.TrimRight(m.ConsoleURL, "/")+"/alerts")
	}
	return strings.Join(lines, "\n")
}

func (m Message) Markdown() string {
	lines := []string{
		fmt.Sprintf("### [%s] %s", strings.ToUpper(m.Alert.Level), m.Alert.Title),
		fmt.Sprintf("- **来源 IP**：%s", valueOr(m.Alert.SourceIP, "-")),
		fmt.Sprintf("- **蜜罐服务**：%s", valueOr(m.Alert.Service, "-")),
		fmt.Sprintf("- **详情**：%s", valueOr(m.Alert.Description, "-")),
		fmt.Sprintf("- **时间**：%s", m.Alert.CreatedAt.Local().Format("2006-01-02 15:04:05 MST")),
	}
	if m.ConsoleURL != "" {
		lines = append(lines, fmt.Sprintf("[打开 Honeynet 控制台](%s/alerts)", strings.TrimRight(m.ConsoleURL, "/")))
	}
	return strings.Join(lines, "\n")
}

func (m Message) JSON() []byte {
	payload := struct {
		Version   string      `json:"version"`
		Type      string      `json:"type"`
		Timestamp time.Time   `json:"timestamp"`
		Data      interface{} `json:"data"`
	}{Version: "1", Type: "honeynet.alert", Timestamp: time.Now().UTC(), Data: m}
	data, _ := json.Marshal(payload)
	return data
}

func valueOr(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
