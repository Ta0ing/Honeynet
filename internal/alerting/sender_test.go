package alerting

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

func testAlertMessage() Message {
	return Message{ConsoleURL: "https://honeynet.example", Alert: store.Alert{
		Base:    store.Base{ID: "alert-1", CreatedAt: time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)},
		EventID: "event-1", Title: "SSH 凭据捕获", Level: "high", SourceIP: "203.0.113.9",
		Service: "ssh", Description: "来源 203.0.113.9 尝试弱口令登录",
	}}
}

func TestWebhookSignatureAndPayload(t *testing.T) {
	var receivedBody []byte
	var timestamp, signature string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody = make([]byte, r.ContentLength)
		_, _ = r.Body.Read(receivedBody)
		timestamp = r.Header.Get("X-Honeynet-Timestamp")
		signature = r.Header.Get("X-Honeynet-Signature")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	config, _ := json.Marshal(WebhookConfig{URL: server.URL, Secret: "webhook-secret"})
	channel := store.AlertChannel{Type: "webhook", Config: datatypes.JSON(config)}
	sender := NewSender()
	sender.now = func() time.Time { return time.Unix(1_786_363_200, 0) }
	if err := sender.Send(context.Background(), channel, testAlertMessage()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(receivedBody), `"type":"honeynet.alert"`) {
		t.Fatalf("unexpected body: %s", receivedBody)
	}
	mac := hmac.New(sha256.New, []byte("webhook-secret"))
	_, _ = mac.Write([]byte(timestamp + "\n"))
	_, _ = mac.Write(receivedBody)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if signature != want {
		t.Fatalf("signature = %q, want %q", signature, want)
	}
}

func TestRobotPayloads(t *testing.T) {
	tests := []struct {
		channelType string
		response    string
		bodyPart    string
	}{
		{"wecom", `{"errcode":0,"errmsg":"ok"}`, `"msgtype":"markdown"`},
		{"dingtalk", `{"errcode":0,"errmsg":"ok"}`, `"msgtype":"markdown"`},
		{"feishu", `{"code":0,"msg":"success"}`, `"msg_type":"text"`},
	}
	for _, test := range tests {
		t.Run(test.channelType, func(t *testing.T) {
			var body string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data := make([]byte, r.ContentLength)
				_, _ = r.Body.Read(data)
				body = string(data)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(test.response))
			}))
			defer server.Close()
			config, _ := json.Marshal(RobotConfig{WebhookURL: server.URL, Secret: "robot-secret"})
			channel := store.AlertChannel{Type: test.channelType, Config: datatypes.JSON(config)}
			if err := NewSender().Send(context.Background(), channel, testAlertMessage()); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(body, test.bodyPart) {
				t.Fatalf("unexpected body: %s", body)
			}
			if test.channelType == "feishu" && (!strings.Contains(body, `"timestamp"`) || !strings.Contains(body, `"sign"`)) {
				t.Fatalf("signed Feishu payload missing fields: %s", body)
			}
		})
	}
}

func TestRobotErrorIsReturned(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":310000,"errmsg":"invalid token"}`))
	}))
	defer server.Close()
	config, _ := json.Marshal(RobotConfig{WebhookURL: server.URL})
	err := NewSender().Send(context.Background(), store.AlertChannel{Type: "wecom", Config: datatypes.JSON(config)}, testAlertMessage())
	if err == nil || !strings.Contains(err.Error(), "310000") {
		t.Fatalf("expected robot error, got %v", err)
	}
}

func TestSyslogUDP(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	config, _ := json.Marshal(SyslogConfig{Address: listener.LocalAddr().String(), Network: "udp", Facility: 16})
	channel := store.AlertChannel{Type: "syslog", Config: datatypes.JSON(config)}
	if err := NewSender().Send(context.Background(), channel, testAlertMessage()); err != nil {
		t.Fatal(err)
	}
	_ = listener.SetReadDeadline(time.Now().Add(time.Second))
	buffer := make([]byte, 4096)
	n, _, err := listener.ReadFrom(buffer)
	if err != nil {
		t.Fatal(err)
	}
	if line := string(buffer[:n]); !strings.HasPrefix(line, "<131>1 ") || !strings.Contains(line, "203.0.113.9") {
		t.Fatalf("unexpected RFC5424 message: %s", line)
	}
}

func TestConfigRedactionAndMerge(t *testing.T) {
	existing := json.RawMessage(`{"url":"https://example.test/hook/token","secret":"secret","headers":{"Authorization":"Bearer token"}}`)
	redacted := RedactedConfig("webhook", existing)
	if strings.Contains(string(redacted), `"secret":"secret"`) || strings.Contains(string(redacted), "Bearer token") || strings.Contains(string(redacted), "/hook/token") {
		t.Fatalf("config was not redacted: %s", redacted)
	}
	merged, err := MergeConfig("webhook", existing, json.RawMessage(`{"url":"********","secret":"********","headers":{"Authorization":"********"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(merged) == string(redacted) || !strings.Contains(string(merged), "Bearer token") {
		t.Fatalf("secret values were not preserved: %s", merged)
	}
}

func TestEmailBodyUsesUTF8Subject(t *testing.T) {
	body := emailBody("alerts@example.test", []string{"soc@example.test"}, testAlertMessage())
	if !strings.Contains(body, "Subject: =?UTF-8?B?") || !strings.Contains(body, "203.0.113.9") {
		t.Fatalf("unexpected email body: %s", body)
	}
}

func TestSMTPDelivery(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	messageBody := make(chan string, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		_, _ = writer.WriteString("220 smtp.example.test ESMTP\r\n")
		_ = writer.Flush()
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return
			}
			command := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(command, "EHLO"):
				_, _ = writer.WriteString("250-smtp.example.test\r\n250 OK\r\n")
			case strings.HasPrefix(command, "MAIL FROM"), strings.HasPrefix(command, "RCPT TO"):
				_, _ = writer.WriteString("250 OK\r\n")
			case command == "DATA":
				_, _ = writer.WriteString("354 End data with <CR><LF>.<CR><LF>\r\n")
				_ = writer.Flush()
				var data strings.Builder
				for {
					dataLine, dataErr := reader.ReadString('\n')
					if dataErr != nil {
						return
					}
					if dataLine == ".\r\n" {
						break
					}
					data.WriteString(dataLine)
				}
				messageBody <- data.String()
				_, _ = writer.WriteString("250 queued\r\n")
			case command == "QUIT":
				_, _ = writer.WriteString("221 bye\r\n")
				_ = writer.Flush()
				return
			default:
				_, _ = writer.WriteString("250 OK\r\n")
			}
			_ = writer.Flush()
		}
	}()
	host, portText, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portText)
	config, _ := json.Marshal(EmailConfig{Host: host, Port: port, From: "alerts@example.test", To: []string{"soc@example.test"}, TLSMode: "none"})
	channel := store.AlertChannel{Type: "email", Config: datatypes.JSON(config)}
	if err := NewSender().Send(context.Background(), channel, testAlertMessage()); err != nil {
		t.Fatal(err)
	}
	select {
	case body := <-messageBody:
		if !strings.Contains(body, "203.0.113.9") || !strings.Contains(body, "Content-Type: text/plain") {
			t.Fatalf("unexpected SMTP body: %s", body)
		}
	case <-time.After(time.Second):
		t.Fatal("SMTP server did not receive a message")
	}
}
