package sense

import (
	"encoding/binary"
	"reflect"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func TestDetectorAggregatesDistinctPorts(t *testing.T) {
	detector, err := NewDetector(protocol.SenseConfig{TCPEnabled: true, DistinctPorts: 3, WindowSeconds: 10, CooldownSeconds: 60, ExcludedPorts: []int{22}, IgnoredCIDRs: []string{"10.0.0.0/8"}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1000, 0)
	if result := detector.Observe(now, Probe{Protocol: "tcp", SourceIP: "10.1.2.3", DestPort: 80}); result != nil {
		t.Fatal("ignored network produced a detection")
	}
	for index, port := range []int{22, 80, 443, 8080} {
		result := detector.Observe(now.Add(time.Duration(index)*time.Second), Probe{Protocol: "tcp", SourceIP: "192.0.2.10", SourcePort: 50000 + index, DestIP: "192.0.2.20", DestPort: port})
		if index < 3 && result != nil {
			t.Fatalf("detected too early at index %d", index)
		}
		if index == 3 {
			if result == nil || result.DistinctPorts != 3 || result.Attempts != 3 {
				t.Fatalf("unexpected detection: %#v", result)
			}
		}
	}
	for _, port := range []int{9000, 9001, 9002} {
		if detector.Observe(now.Add(10*time.Second), Probe{Protocol: "tcp", SourceIP: "192.0.2.10", DestPort: port}) != nil {
			t.Fatal("cooldown did not suppress a repeated scan")
		}
	}
}

func TestDetectorWindowExpires(t *testing.T) {
	detector, err := NewDetector(protocol.SenseConfig{UDPEnabled: true, DistinctPorts: 3, WindowSeconds: 2, CooldownSeconds: 10})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(2000, 0)
	for _, port := range []int{53, 67} {
		if detector.Observe(now, Probe{Protocol: "udp", SourceIP: "198.51.100.1", DestPort: port}) != nil {
			t.Fatal("detected too early")
		}
	}
	if detector.Observe(now.Add(3*time.Second), Probe{Protocol: "udp", SourceIP: "198.51.100.1", DestPort: 161}) != nil {
		t.Fatal("expired session was not reset")
	}
}

func TestNormalizeConfigDoesNotMutateInput(t *testing.T) {
	input := DefaultConfig()
	input.ExcludedPorts = []int{443, 22, 443}
	input.IgnoredCIDRs = []string{" 10.0.0.0/8 ", "192.168.0.0/16"}
	portsBefore := append([]int(nil), input.ExcludedPorts...)
	cidrsBefore := append([]string(nil), input.IgnoredCIDRs...)

	normalized, err := NormalizeConfig(input)
	if err != nil {
		t.Fatalf("NormalizeConfig() error = %v", err)
	}
	if !reflect.DeepEqual(input.ExcludedPorts, portsBefore) || !reflect.DeepEqual(input.IgnoredCIDRs, cidrsBefore) {
		t.Fatalf("NormalizeConfig mutated its input: %#v", input)
	}
	if !reflect.DeepEqual(normalized.ExcludedPorts, []int{22, 443}) {
		t.Fatalf("ExcludedPorts = %#v", normalized.ExcludedPorts)
	}
	if !reflect.DeepEqual(normalized.IgnoredCIDRs, []string{"10.0.0.0/8", "192.168.0.0/16"}) {
		t.Fatalf("IgnoredCIDRs = %#v", normalized.IgnoredCIDRs)
	}
}

func TestNormalizeConfigCanonicalizesIPv6CIDRs(t *testing.T) {
	input := DefaultConfig()
	input.IgnoredCIDRs = []string{"fd12:3456:789a::1/48", "fd12:3456:789a::/48", "2001:0db8::/32"}
	normalized, err := NormalizeConfig(input)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(normalized.IgnoredCIDRs, []string{"2001:db8::/32", "fd12:3456:789a::/48"}) {
		t.Fatalf("canonical IPv6 CIDRs = %#v", normalized.IgnoredCIDRs)
	}
}

func TestDetectorAggregatesIPv6AndHonorsIgnoredULA(t *testing.T) {
	detector, err := NewDetector(protocol.SenseConfig{
		TCPEnabled: true, DistinctPorts: 3, WindowSeconds: 10, CooldownSeconds: 60,
		IgnoredCIDRs: []string{"fd00::/8"},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(3000, 0)
	if result := detector.Observe(now, Probe{Protocol: "tcp", SourceIP: "fd12:3456::20", DestIP: "2001:db8::20", DestPort: 80}); result != nil {
		t.Fatal("ignored IPv6 ULA produced a detection")
	}
	var detected *Detection
	for index, port := range []int{80, 443, 8443} {
		detected = detector.Observe(now.Add(time.Duration(index)*time.Second), Probe{
			Protocol: "tcp", SourceIP: "2001:0db8:0:0::10", SourcePort: 50000 + index,
			DestIP: "2001:0db8:0:0::20", DestPort: port,
		})
	}
	if detected == nil || detected.SourceIP != "2001:db8::10" || detected.DestIP != "2001:db8::20" {
		t.Fatalf("IPv6 detection = %#v", detected)
	}
}

func TestDecodeIPv4TCPSYNAndUDP(t *testing.T) {
	tcp := ethernetIPv4Packet(6, 40000, 443, 0x02)
	probe, ok := DecodePacket(tcp)
	if !ok || probe.Protocol != "tcp" || probe.SourceIP != "192.0.2.1" || probe.DestPort != 443 {
		t.Fatalf("unexpected TCP probe: %#v %v", probe, ok)
	}
	ack := ethernetIPv4Packet(6, 40000, 443, 0x12)
	if _, ok := DecodePacket(ack); ok {
		t.Fatal("SYN-ACK was treated as a scan probe")
	}
	udp := ethernetIPv4Packet(17, 50000, 53, 0)
	probe, ok = DecodePacket(udp)
	if !ok || probe.Protocol != "udp" || probe.DestPort != 53 {
		t.Fatalf("unexpected UDP probe: %#v %v", probe, ok)
	}
}

func TestDecodeIPv6TransportAndExtensionHeaders(t *testing.T) {
	for _, test := range []struct {
		name      string
		protocol  byte
		extension bool
		flags     byte
		wantProto string
		wantDest  int
	}{
		{name: "tcp", protocol: 6, flags: 0x02, wantProto: "tcp", wantDest: 443},
		{name: "udp hop by hop", protocol: 17, extension: true, wantProto: "udp", wantDest: 53},
	} {
		t.Run(test.name, func(t *testing.T) {
			packet := ethernetIPv6Packet(test.protocol, 50000, test.wantDest, test.flags, test.extension)
			probe, ok := DecodePacket(packet)
			if !ok || probe.Protocol != test.wantProto || probe.SourceIP != "2001:db8::1" || probe.DestIP != "2001:db8::2" || probe.DestPort != test.wantDest {
				t.Fatalf("IPv6 probe = %#v, %v", probe, ok)
			}
		})
	}

	nonFirstFragment := ethernetIPv6Packet(6, 50000, 443, 0x02, false)
	ip := nonFirstFragment[14:]
	transport := append([]byte(nil), ip[40:]...)
	ip[6] = 44
	fragment := make([]byte, 8)
	fragment[0] = 6
	binary.BigEndian.PutUint16(fragment[2:4], 8) // fragment offset 1
	nonFirstFragment = append(nonFirstFragment[:14+40], append(fragment, transport...)...)
	if _, ok := DecodePacket(nonFirstFragment); ok {
		t.Fatal("non-first IPv6 fragment was treated as a scan probe")
	}
}

func ethernetIPv4Packet(protocol byte, sourcePort, destPort int, flags byte) []byte {
	transportLength := 20
	if protocol == 17 {
		transportLength = 8
	}
	packet := make([]byte, 14+20+transportLength)
	binary.BigEndian.PutUint16(packet[12:14], etherIPv4)
	ip := packet[14:]
	ip[0] = 0x45
	ip[9] = protocol
	copy(ip[12:16], []byte{192, 0, 2, 1})
	copy(ip[16:20], []byte{192, 0, 2, 2})
	transport := ip[20:]
	binary.BigEndian.PutUint16(transport[0:2], uint16(sourcePort))
	binary.BigEndian.PutUint16(transport[2:4], uint16(destPort))
	if protocol == 6 {
		transport[12] = 0x50
		transport[13] = flags
	}
	return packet
}

func ethernetIPv6Packet(protocol byte, sourcePort, destPort int, flags byte, extension bool) []byte {
	transportLength := 20
	if protocol == 17 {
		transportLength = 8
	}
	extensionLength := 0
	if extension {
		extensionLength = 8
	}
	packet := make([]byte, 14+40+extensionLength+transportLength)
	binary.BigEndian.PutUint16(packet[12:14], etherIPv6)
	ip := packet[14:]
	ip[0] = 0x60
	ip[6] = protocol
	if extension {
		ip[6] = 0
	}
	copy(ip[8:24], []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	copy(ip[24:40], []byte{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2})
	offset := 40
	if extension {
		ip[offset] = protocol
		offset += 8
	}
	transport := ip[offset:]
	binary.BigEndian.PutUint16(transport[0:2], uint16(sourcePort))
	binary.BigEndian.PutUint16(transport[2:4], uint16(destPort))
	if protocol == 6 {
		transport[12] = 0x50
		transport[13] = flags
	}
	return packet
}
