package pots

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type DNSService struct {
	conn net.PacketConn
	once sync.Once
}

func (s *DNSService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	conn, err := net.ListenPacket("udp", listenAddress(target))
	if err != nil {
		return err
	}
	s.conn = conn
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	go s.serve(conn, sink)
	return nil
}

func (s *DNSService) Stop() error {
	if s.conn == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.conn.Close() })
	return err
}

func (s *DNSService) serve(conn net.PacketConn, sink Sink) {
	buffer := make([]byte, 4096)
	for {
		count, remote, err := conn.ReadFrom(buffer)
		if err != nil {
			return
		}
		request := append([]byte(nil), buffer[:count]...)
		name, queryType, questionEnd, err := parseDNSQuestion(request)
		if err != nil {
			continue
		}
		sink(protocol.NewEvent("dns.query", endpoint(remote), endpoint(conn.LocalAddr()), map[string]any{"name": name, "query_type": dnsTypeName(queryType)}, "recon"))
		response := dnsResponse(request, questionEnd, queryType)
		_, _ = conn.WriteTo(response, remote)
	}
}

func parseDNSQuestion(packet []byte) (string, uint16, int, error) {
	if len(packet) < 12 || binary.BigEndian.Uint16(packet[4:6]) == 0 {
		return "", 0, 0, errors.New("DNS question is missing")
	}
	offset := 12
	labels := make([]string, 0, 4)
	for {
		if offset >= len(packet) {
			return "", 0, 0, errors.New("truncated DNS name")
		}
		length := int(packet[offset])
		offset++
		if length == 0 {
			break
		}
		if length > 63 || offset+length > len(packet) {
			return "", 0, 0, errors.New("invalid DNS label")
		}
		labels = append(labels, string(packet[offset:offset+length]))
		offset += length
	}
	if offset+4 > len(packet) {
		return "", 0, 0, errors.New("truncated DNS question")
	}
	queryType := binary.BigEndian.Uint16(packet[offset : offset+2])
	return strings.Join(labels, "."), queryType, offset + 4, nil
}

func dnsResponse(request []byte, questionEnd int, queryType uint16) []byte {
	answerCount := uint16(0)
	if queryType == 1 {
		answerCount = 1
	}
	response := make([]byte, 12, 64)
	copy(response[0:2], request[0:2])
	binary.BigEndian.PutUint16(response[2:4], 0x8180)
	binary.BigEndian.PutUint16(response[4:6], 1)
	binary.BigEndian.PutUint16(response[6:8], answerCount)
	response = append(response, request[12:questionEnd]...)
	if answerCount == 1 {
		answer := []byte{0xc0, 0x0c, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x3c, 0x00, 0x04, 192, 0, 2, 10}
		response = append(response, answer...)
	}
	return response
}

func dnsTypeName(value uint16) string {
	switch value {
	case 1:
		return "A"
	case 2:
		return "NS"
	case 5:
		return "CNAME"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	case 28:
		return "AAAA"
	default:
		return "TYPE" + strconv.Itoa(int(value))
	}
}
