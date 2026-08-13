package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/config"
	"github.com/honeynet/honeynet/internal/nodepki"
)

func TestBootstrapClientPinsCAAndVerifiesHostname(t *testing.T) {
	authority, err := nodepki.LoadOrCreate(t.TempDir(), []string{"127.0.0.1"}, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	tlsListener := tlsListener(listener, authority)
	server := &http.Server{Handler: http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })}
	go server.Serve(tlsListener)
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })

	cfg := &config.Config{
		ServerURL: "http://127.0.0.1:8080", AgentURL: "https://" + listener.Addr().String(), NodeID: "node-pin",
		CAFingerprint: authority.CAFingerprint(), StateDir: t.TempDir(), ConfigPath: t.TempDir() + "/agent.json",
	}
	if err := cfg.Normalize(); err != nil {
		t.Fatal(err)
	}
	client := &Client{cfg: cfg}
	httpClient, err := client.bootstrapHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	response, err := httpClient.Get(cfg.AgentURL)
	if err != nil {
		t.Fatalf("pinned TLS request failed: %v", err)
	}
	response.Body.Close()

	cfg.CAFingerprint = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	httpClient, err = client.bootstrapHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	if response, err = httpClient.Get(cfg.AgentURL); err == nil {
		response.Body.Close()
		t.Fatal("request with the wrong CA fingerprint succeeded")
	}
}

func TestBootstrapClientPinsCAOverIPv6(t *testing.T) {
	authority, err := nodepki.LoadOrCreate(t.TempDir(), []string{"::1"}, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 loopback is unavailable: %v", err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })}
	go server.Serve(tlsListener(listener, authority))
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })

	cfg := &config.Config{
		ServerURL: "http://[::1]:8080", AgentURL: "https://" + listener.Addr().String(), NodeID: "node-ipv6-pin",
		CAFingerprint: authority.CAFingerprint(), StateDir: t.TempDir(), ConfigPath: t.TempDir() + "/agent.json",
	}
	if err := cfg.Normalize(); err != nil {
		t.Fatal(err)
	}
	client := &Client{cfg: cfg}
	httpClient, err := client.bootstrapHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	response, err := httpClient.Get(cfg.AgentURL)
	if err != nil {
		t.Fatalf("pinned IPv6 TLS request failed: %v", err)
	}
	response.Body.Close()
}

func tlsListener(listener net.Listener, authority *nodepki.Authority) net.Listener {
	return tls.NewListener(listener, authority.ServerTLSConfig())
}
