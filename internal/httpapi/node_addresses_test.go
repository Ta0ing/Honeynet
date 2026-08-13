package httpapi

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/honeynet/honeynet/internal/store"
	"gorm.io/datatypes"
)

func TestMergeNodeAddressReportPrefersObservedPublicIP(t *testing.T) {
	node := store.Node{AddressMode: nodeAddressAuto, PrivateIPs: datatypes.JSON(`[]`)}
	mergeNodeAddressReport(&node, "47.96.80.162", []string{"10.3.0.6", "192.168.16.2"})

	if node.IP != "47.96.80.162" || node.PublicIP != "47.96.80.162" {
		t.Fatalf("selected/public IP = %q/%q", node.IP, node.PublicIP)
	}
	if got := decodeNodeIPs(node.PrivateIPs); !reflect.DeepEqual(got, []string{"10.3.0.6", "192.168.16.2"}) {
		t.Fatalf("private IPs = %#v", got)
	}
	mergeNodeAddressReport(&node, "", []string{"198.51.100.20", "10.3.0.6"})
	if node.IP != "47.96.80.162" || node.PublicIP != "47.96.80.162" {
		t.Fatalf("heartbeat replaced the observed NAT address: %q/%q", node.IP, node.PublicIP)
	}
}

func TestMergeNodeAddressReportPreservesManualSelection(t *testing.T) {
	tests := []store.Node{
		{AddressMode: nodeAddressPrivate, IP: "192.168.16.2", PublicIP: "47.96.80.162", PrivateIPs: datatypes.JSON(`["192.168.16.2"]`)},
		{AddressMode: nodeAddressPublic, IP: "47.100.1.2", PublicIP: "47.96.80.162", PrivateIPs: datatypes.JSON(`["10.3.0.6"]`)},
		{AddressMode: nodeAddressCustom, IP: "172.20.1.10", PublicIP: "47.96.80.162", PrivateIPs: datatypes.JSON(`["10.3.0.6"]`)},
	}
	for _, node := range tests {
		want := node.IP
		mergeNodeAddressReport(&node, "47.96.80.163", []string{"10.3.0.7"})
		if node.IP != want {
			t.Fatalf("mode %s changed selected IP from %q to %q", node.AddressMode, want, node.IP)
		}
	}
}

func TestMergeNodeAddressReportUsesPrivateIPWithoutPublicCandidate(t *testing.T) {
	node := store.Node{AddressMode: nodeAddressAuto}
	mergeNodeAddressReport(&node, "10.3.0.6", []string{"192.168.16.2"})
	if node.IP != "192.168.16.2" {
		t.Fatalf("selected IP = %q", node.IP)
	}
	if got := decodeNodeIPs(node.PrivateIPs); !reflect.DeepEqual(got, []string{"192.168.16.2", "10.3.0.6"}) {
		t.Fatalf("private IPs = %#v", got)
	}
}

func TestConfigureNodeAddress(t *testing.T) {
	node := store.Node{PublicIP: "47.96.80.162", PrivateIPs: datatypes.JSON(`["10.3.0.6","192.168.16.2"]`)}
	if err := configureNodeAddress(&node, nodeAddressPrivate, "192.168.16.2"); err != nil {
		t.Fatal(err)
	}
	if node.AddressMode != nodeAddressPrivate || node.IP != "192.168.16.2" {
		t.Fatalf("configured node = %#v", node)
	}
	if err := configureNodeAddress(&node, nodeAddressPublic, "10.3.0.6"); err == nil {
		t.Fatal("configureNodeAddress accepted a private address as public")
	}
	if err := configureNodeAddress(&node, "unknown", "47.96.80.162"); err == nil {
		t.Fatal("configureNodeAddress accepted an unknown mode")
	}
}

func TestRequestRemoteIP(t *testing.T) {
	for _, test := range []struct {
		remote string
		want   string
	}{
		{remote: "47.96.80.162:45678", want: "47.96.80.162"},
		{remote: "[2001:4860:4860::8888]:45678", want: "2001:4860:4860::8888"},
	} {
		request := &http.Request{RemoteAddr: test.remote}
		if got := requestRemoteIP(request); got != test.want {
			t.Fatalf("requestRemoteIP(%q) = %q, want %q", test.remote, got, test.want)
		}
	}
}

func TestMergeNodeAddressReportSupportsDualStackIPv6Selection(t *testing.T) {
	node := store.Node{AddressMode: nodeAddressAuto, PublicIPs: datatypes.JSON(`[]`), PrivateIPs: datatypes.JSON(`[]`)}
	mergeNodeAddressReport(&node, "2001:4860:4860::8844", []string{
		"192.168.16.2", "fd12:3456:789a::20", "47.96.80.162", "2001:4860:4860::8888",
	})
	if node.IP != "2001:4860:4860::8844" || node.PublicIP != "2001:4860:4860::8844" {
		t.Fatalf("selected/public IP = %q/%q", node.IP, node.PublicIP)
	}
	if got := decodePublicNodeIPs(node.PublicIPs); !reflect.DeepEqual(got, []string{"2001:4860:4860::8844", "47.96.80.162", "2001:4860:4860::8888"}) {
		t.Fatalf("public IPs = %#v", got)
	}
	if got := decodeNodeIPs(node.PrivateIPs); !reflect.DeepEqual(got, []string{"192.168.16.2", "fd12:3456:789a::20"}) {
		t.Fatalf("private IPs = %#v", got)
	}

	if err := configureNodeAddress(&node, nodeAddressPublic, "2001:4860:4860::8888"); err != nil {
		t.Fatal(err)
	}
	mergeNodeAddressReport(&node, "47.96.80.163", []string{"47.96.80.162", "2001:4860:4860::8888", "fd12:3456:789a::20"})
	if node.IP != "2001:4860:4860::8888" || node.PublicIP != "2001:4860:4860::8888" {
		t.Fatalf("manual IPv6 selection changed after heartbeat: %#v", node)
	}
}

func TestConfigureNodeAddressAcceptsPrivateIPv6(t *testing.T) {
	node := store.Node{PrivateIPs: datatypes.JSON(`["fd12:3456:789a::20"]`)}
	if err := configureNodeAddress(&node, nodeAddressPrivate, "fd12:3456:789a::20"); err != nil {
		t.Fatal(err)
	}
	if node.IP != "fd12:3456:789a::20" {
		t.Fatalf("selected private IPv6 = %q", node.IP)
	}
}
