package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	agentconfig "github.com/honeynet/honeynet/internal/agent/config"
)

func TestControlChannelReconnectsWhenServerGoesSilent(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Drain the Agent hello/heartbeat frames, but deliberately never send an
		// acknowledgement. This models a TCP session that stays established while
		// the remote process or one direction of the network path is dead.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	root := t.TempDir()
	cfg := &agentconfig.Config{AgentURL: server.URL, StateDir: root, TemplateRoot: root}
	client, err := New(cfg, "test")
	if err != nil {
		t.Fatal(err)
	}
	client.tls = &tls.Config{MinVersion: tls.VersionTLS13}
	client.serverSilenceTimeout = 150 * time.Millisecond
	started := time.Now()
	established, err := client.connect(context.Background())
	if !established || err == nil {
		t.Fatalf("connect established=%v err=%v; silent established channel must time out", established, err)
	}
	var networkError net.Error
	if !errors.As(err, &networkError) || !networkError.Timeout() {
		t.Fatalf("connect error=%T %v, want timeout", err, err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("silent channel detection took %s", elapsed)
	}
}
