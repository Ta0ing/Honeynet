package pots

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
	gossh "golang.org/x/crypto/ssh"
)

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
func target(code string, port int) protocol.PotTarget {
	return protocol.PotTarget{ID: "test", Service: code, Port: port, DesiredStatus: "running", Config: map[string]any{"bind": "127.0.0.1"}}
}
func waitType(t *testing.T, events <-chan protocol.Event, eventType string) protocol.Event {
	t.Helper()
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	for {
		select {
		case event := <-events:
			if event.EventType == eventType {
				return event
			}
		case <-timer.C:
			t.Fatalf("did not receive %s event", eventType)
		}
	}
}

func TestHTTPServiceCapturesCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &HTTPService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("http", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	form := url.Values{"username": {"root"}, "password": {"secret"}}
	resp, err := http.PostForm("http://127.0.0.1:"+strconv.Itoa(port)+"/login", form)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	event := waitType(t, events, "web.credential")
	if event.Payload["username"] != "root" || event.Payload["password"] != "secret" {
		t.Fatalf("unexpected payload: %#v", event.Payload)
	}
}

func TestTemplateHTTPServiceRoutesAndCapturesJSON(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &TemplateHTTPService{}
	port := freePort(t)
	target := target("web-template", port)
	target.Template = &protocol.WebTemplate{ID: "template-1", Name: "Fake OA", Version: 2, YAML: `name: fake-oa
listen:
  port: 8080
pages:
  - path: /api/login
    method: POST
    capture:
      fields: [username, password]
    response:
      status: 401
      headers:
        X-Portal: fake-oa
      body: denied-v2
`}
	if err := service.Start(context.Background(), target, func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	body := strings.NewReader(`{"username":"admin","password":"secret"}`)
	request, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:"+strconv.Itoa(port)+"/api/login", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	responseBody, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized || response.Header.Get("X-Portal") != "fake-oa" || string(responseBody) != "denied-v2" {
		t.Fatalf("unexpected response: status=%d headers=%v body=%q", response.StatusCode, response.Header, responseBody)
	}
	requestEvent := waitType(t, events, "web.request")
	if requestEvent.Payload["template_id"] != "template-1" || requestEvent.Payload["matched"] != true {
		t.Fatalf("unexpected request event: %#v", requestEvent.Payload)
	}
	credential := waitType(t, events, "web.credential")
	if credential.Payload["username"] != "admin" || credential.Payload["password"] != "secret" || credential.Payload["template_version"] != 2 {
		t.Fatalf("unexpected credential event: %#v", credential.Payload)
	}
}

func TestTemplateHTTPServiceRejectsMissingTemplate(t *testing.T) {
	service := &TemplateHTTPService{}
	if err := service.Start(context.Background(), target("web-template", freePort(t)), func(protocol.Event) {}); err == nil {
		t.Fatal("Start accepted a missing template")
	}
}

func TestSSHServiceHandshakeAndCommand(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &SSHService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("ssh", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	config := &gossh.ClientConfig{User: "admin", Auth: []gossh.AuthMethod{gossh.Password("secret")}, HostKeyCallback: gossh.InsecureIgnoreHostKey(), Timeout: 3 * time.Second}
	client, err := gossh.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), config)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	output, err := session.CombinedOutput("id")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output, []byte("uid=1000")) {
		t.Fatalf("unexpected command output: %q", output)
	}
	credential := waitType(t, events, "ssh.credential")
	if credential.Payload["username"] != "admin" || credential.Payload["password"] != "secret" {
		t.Fatalf("unexpected credential: %#v", credential.Payload)
	}
	waitType(t, events, "ssh.command")
}

func TestTelnetServiceCapturesCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &TelnetService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("telnet", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	if _, err = reader.ReadString(':'); err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(conn, "admin\r\n")
	if _, err = reader.ReadString(':'); err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(conn, "secret\r\n")
	event := waitType(t, events, "telnet.credential")
	if event.Payload["username"] != "admin" || event.Payload["password"] != "secret" {
		t.Fatalf("unexpected credential: %#v", event.Payload)
	}
}

func TestTelnetServiceDoesNotCaptureDisconnectedLogin(t *testing.T) {
	tests := []struct {
		name         string
		sendUsername bool
	}{
		{name: "disconnect before username"},
		{name: "disconnect before password", sendUsername: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := make(chan protocol.Event, 2)
			service := &TelnetService{}
			port := freePort(t)
			if err := service.Start(context.Background(), target("telnet", port), func(event protocol.Event) { events <- event }); err != nil {
				t.Fatal(err)
			}
			defer service.Stop()
			conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
			if err != nil {
				t.Fatal(err)
			}
			reader := bufio.NewReader(conn)
			if _, err = reader.ReadString(':'); err != nil {
				t.Fatal(err)
			}
			if test.sendUsername {
				_, _ = io.WriteString(conn, "admin\r\n")
				if _, err = reader.ReadString(':'); err != nil {
					t.Fatal(err)
				}
			}
			_ = conn.Close()
			select {
			case event := <-events:
				t.Fatalf("disconnected login emitted %s with payload %#v", event.EventType, event.Payload)
			case <-time.After(200 * time.Millisecond):
			}
		})
	}
}

func TestRedisServiceProtocol(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &RedisService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("redis", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	_, _ = io.WriteString(conn, "*1\r\n$4\r\nPING\r\n")
	line, err := reader.ReadString('\n')
	if err != nil || line != "+PONG\r\n" {
		t.Fatalf("PING response = %q, %v", line, err)
	}
	_, _ = io.WriteString(conn, "*2\r\n$4\r\nAUTH\r\n$6\r\nsecret\r\n")
	line, err = reader.ReadString('\n')
	if err != nil || line != "+OK\r\n" {
		t.Fatalf("AUTH response = %q, %v", line, err)
	}
	event := waitType(t, events, "redis.credential")
	if event.Payload["password"] != "secret" {
		t.Fatalf("unexpected credential: %#v", event.Payload)
	}
}

func TestMySQLServiceHandshake(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &MySQLService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("mysql", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, _, err = readMySQLPacket(conn); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, 32)
	binary.LittleEndian.PutUint32(response, mysqlCapabilities)
	response = append(response, []byte("root")...)
	response = append(response, 0, 4, 1, 2, 3, 4)
	if err = writeMySQLPacket(conn, 1, response); err != nil {
		t.Fatal(err)
	}
	ok, _, err := readMySQLPacket(conn)
	if err != nil {
		t.Fatal(err)
	}
	if len(ok) == 0 || ok[0] != 0x00 {
		t.Fatalf("unexpected OK packet: %x", ok)
	}
	event := waitType(t, events, "mysql.credential")
	if event.Payload["username"] != "root" || !strings.Contains(event.Payload["auth_response"].(string), "01020304") {
		t.Fatalf("unexpected credential: %#v", event.Payload)
	}
}
