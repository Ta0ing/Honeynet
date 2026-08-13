package pots

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func TestListenAddressAndEndpointSupportIPv6(t *testing.T) {
	for _, bind := range []string{"::", "[::]"} {
		target := protocol.PotTarget{Port: 8080, Config: map[string]any{"bind": bind}}
		if got := listenAddress(target); got != "[::]:8080" {
			t.Fatalf("listenAddress(%q) = %q", bind, got)
		}
	}
	endpoint := endpoint(&net.TCPAddr{IP: net.ParseIP("2001:db8::20"), Port: 443})
	if endpoint.IP != "2001:db8::20" || endpoint.Port != 443 {
		t.Fatalf("IPv6 endpoint = %#v", endpoint)
	}
	if got := canonicalEndpointIP("fe80::1%eth0"); got != "fe80::1" {
		t.Fatalf("scoped endpoint IP = %q", got)
	}
}

func TestHTTPServiceAcceptsIPv6Connection(t *testing.T) {
	probe, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 loopback is unavailable: %v", err)
	}
	port := probe.Addr().(*net.TCPAddr).Port
	_ = probe.Close()

	events := make(chan protocol.Event, 2)
	service := &HTTPService{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx, protocol.PotTarget{
		ID: "ipv6-http", Service: "http", Port: port,
		Config: map[string]any{"bind": "::1"},
	}, func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()

	client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{Proxy: nil}}
	response, err := client.Get(fmt.Sprintf("http://[::1]:%d/ipv6", port))
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	select {
	case event := <-events:
		if event.Src.IP != "::1" || event.Dst.IP != "::1" {
			t.Fatalf("IPv6 event endpoints = %#v -> %#v", event.Src, event.Dst)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("IPv6 request did not produce an event")
	}
}
