package runtime

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func TestApplyControlsListenerLifecycle(t *testing.T) {
	runtime := New(func(protocol.Event) {})
	t.Cleanup(runtime.Shutdown)
	firstPort := freeTCPPort(t)
	secondPort := freeTCPPort(t)
	for secondPort == firstPort {
		secondPort = freeTCPPort(t)
	}

	target := protocol.PotTarget{
		ID: "pot-crud", Service: "http", Name: "first", Port: firstPort,
		DesiredStatus: "running", Config: map[string]any{"bind": "127.0.0.1"},
	}
	requireResult(t, runtime.Apply(context.Background(), []protocol.PotTarget{target}), "running", true)
	waitTCPState(t, firstPort, true)

	target.Name = "second"
	target.Port = secondPort
	requireResult(t, runtime.Apply(context.Background(), []protocol.PotTarget{target}), "running", true)
	waitTCPState(t, firstPort, false)
	waitTCPState(t, secondPort, true)

	target.DesiredStatus = "stopped"
	requireResult(t, runtime.Apply(context.Background(), []protocol.PotTarget{target}), "stopped", true)
	waitTCPState(t, secondPort, false)

	target.DesiredStatus = "running"
	requireResult(t, runtime.Apply(context.Background(), []protocol.PotTarget{target}), "running", true)
	waitTCPState(t, secondPort, true)

	if results := runtime.Apply(context.Background(), nil); len(results) != 0 {
		t.Fatalf("delete reconciliation returned results: %#v", results)
	}
	waitTCPState(t, secondPort, false)
	if runtime.Count() != 0 {
		t.Fatalf("runtime still has %d instances after delete", runtime.Count())
	}
}

func requireResult(t *testing.T, results []protocol.PotResult, status string, success bool) {
	t.Helper()
	if len(results) != 1 || results[0].Status != status || results[0].Success != success {
		t.Fatalf("Apply() = %#v, want status=%s success=%t", results, status, success)
	}
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitTCPState(t *testing.T, port int, open bool) {
	t.Helper()
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	deadline := time.Now().Add(2 * time.Second)
	for {
		connection, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			_ = connection.Close()
		}
		if (err == nil) == open {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("TCP %s open=%t, want %t (last error: %v)", address, err == nil, open, err)
		}
		time.Sleep(25 * time.Millisecond)
	}
}
