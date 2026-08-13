package pots

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxTNSPacket = 64 << 10

var tnsAttributePattern = regexp.MustCompile(`(?i)\((SERVICE_NAME|SID|USER|PROGRAM|HOST)=([^()]*)\)`)

type OracleService struct {
	listener net.Listener
	once     sync.Once
}

func (s *OracleService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *OracleService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *OracleService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Minute))
	header := make([]byte, 8)
	if _, err := io.ReadFull(conn, header); err != nil {
		return
	}
	length := int(binary.BigEndian.Uint16(header[:2]))
	if length < len(header) || length > maxTNSPacket {
		return
	}
	payload := make([]byte, length-len(header))
	if _, err := io.ReadFull(conn, payload); err != nil {
		return
	}
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	packetType := int(header[4])
	attributes := parseTNSAttributes(string(payload))
	eventPayload := map[string]any{"packet_type": packetType, "descriptor": printableTNS(payload)}
	for key, value := range attributes {
		eventPayload[key] = value
	}
	sink(protocol.NewEvent("oracle.connect", src, dst, eventPayload, "database", "session"))

	// Return a syntactically framed TNS REFUSE packet. It keeps clients engaged
	// long enough to record the connect descriptor without exposing a SQL engine.
	reason := []byte("(DESCRIPTION=(TMP=)(VSNNUM=318767104)(ERR=12514)(ERROR_STACK=(ERROR=(CODE=12514)(EMFI=4))))")
	responsePayload := append([]byte{0, 0, 0, 0}, reason...)
	response := make([]byte, 8+len(responsePayload))
	binary.BigEndian.PutUint16(response[:2], uint16(len(response)))
	response[4] = 4
	copy(response[8:], responsePayload)
	_, _ = conn.Write(response)
}

func parseTNSAttributes(descriptor string) map[string]any {
	result := make(map[string]any)
	for _, match := range tnsAttributePattern.FindAllStringSubmatch(descriptor, -1) {
		key := strings.ToLower(match[1])
		if _, exists := result[key]; !exists {
			result[key] = strings.TrimSpace(match[2])
		}
	}
	return result
}

func printableTNS(payload []byte) string {
	var builder strings.Builder
	limit := len(payload)
	if limit > 4096 {
		limit = 4096
	}
	for _, value := range payload[:limit] {
		if value >= 0x20 && value <= 0x7e {
			builder.WriteByte(value)
		}
	}
	return builder.String()
}
