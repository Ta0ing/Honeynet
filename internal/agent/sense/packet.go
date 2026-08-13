package sense

import (
	"encoding/binary"
	"net"
)

const (
	etherIPv4 = 0x0800
	etherIPv6 = 0x86dd
	etherVLAN = 0x8100
	etherQinQ = 0x88a8
)

func DecodePacket(packet []byte) (Probe, bool) {
	if len(packet) < 14 {
		return Probe{}, false
	}
	offset := 14
	etherType := binary.BigEndian.Uint16(packet[12:14])
	for etherType == etherVLAN || etherType == etherQinQ {
		if len(packet) < offset+4 {
			return Probe{}, false
		}
		etherType = binary.BigEndian.Uint16(packet[offset+2 : offset+4])
		offset += 4
	}
	switch etherType {
	case etherIPv4:
		return decodeIPv4(packet[offset:])
	case etherIPv6:
		return decodeIPv6(packet[offset:])
	default:
		return Probe{}, false
	}
}

func decodeIPv4(packet []byte) (Probe, bool) {
	if len(packet) < 20 || packet[0]>>4 != 4 {
		return Probe{}, false
	}
	headerLength := int(packet[0]&0x0f) * 4
	if headerLength < 20 || len(packet) < headerLength+8 || binary.BigEndian.Uint16(packet[6:8])&0x1fff != 0 {
		return Probe{}, false
	}
	return decodeTransport(packet[9], packet[headerLength:], net.IP(packet[12:16]).String(), net.IP(packet[16:20]).String())
}

func decodeIPv6(packet []byte) (Probe, bool) {
	if len(packet) < 40 || packet[0]>>4 != 6 {
		return Probe{}, false
	}
	nextHeader := packet[6]
	payload := packet[40:]
	// Walk the common IPv6 extension headers before TCP/UDP. Bound the chain
	// so malformed packets cannot consume unbounded CPU in the capture loop.
	for range 8 {
		switch nextHeader {
		case 0, 43, 60: // Hop-by-Hop, Routing, Destination Options
			if len(payload) < 8 {
				return Probe{}, false
			}
			length := (int(payload[1]) + 1) * 8
			if length > len(payload) {
				return Probe{}, false
			}
			nextHeader, payload = payload[0], payload[length:]
		case 44: // Fragment
			if len(payload) < 8 {
				return Probe{}, false
			}
			// Only the first fragment contains the transport header. Non-first
			// fragments cannot be classified safely without reassembly.
			if binary.BigEndian.Uint16(payload[2:4])>>3 != 0 {
				return Probe{}, false
			}
			nextHeader, payload = payload[0], payload[8:]
		case 51: // Authentication Header
			if len(payload) < 8 {
				return Probe{}, false
			}
			length := (int(payload[1]) + 2) * 4
			if length > len(payload) {
				return Probe{}, false
			}
			nextHeader, payload = payload[0], payload[length:]
		default:
			return decodeTransport(nextHeader, payload, net.IP(packet[8:24]).String(), net.IP(packet[24:40]).String())
		}
	}
	return Probe{}, false
}

func decodeTransport(nextHeader byte, payload []byte, sourceIP, destIP string) (Probe, bool) {
	if len(payload) < 4 {
		return Probe{}, false
	}
	probe := Probe{SourceIP: sourceIP, DestIP: destIP, SourcePort: int(binary.BigEndian.Uint16(payload[0:2])), DestPort: int(binary.BigEndian.Uint16(payload[2:4]))}
	switch nextHeader {
	case 6:
		if len(payload) < 20 || payload[13]&0x02 == 0 || payload[13]&0x10 != 0 {
			return Probe{}, false
		}
		probe.Protocol = "tcp"
	case 17:
		if len(payload) < 8 {
			return Probe{}, false
		}
		probe.Protocol = "udp"
	default:
		return Probe{}, false
	}
	return probe, true
}
