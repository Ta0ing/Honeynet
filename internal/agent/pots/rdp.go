package pots

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxRDPFrame = 64 << 10

type RDPService struct {
	listener net.Listener
	once     sync.Once
}

func (s *RDPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *RDPService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *RDPService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	payload, err := readTPKT(conn)
	if err != nil || len(payload) < 7 || payload[1] != 0xe0 {
		return
	}
	username := rdpCookieUsername(payload[7:])
	requestedProtocols, hasNegotiation := rdpRequestedProtocols(payload[7:])
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	sink(protocol.NewEvent("rdp.connection", src, dst, map[string]any{
		"username": username, "requested_protocols": requestedProtocols, "has_negotiation": hasNegotiation,
	}, "remote-access", "recon"))
	if username != "" {
		sink(protocol.NewEvent("rdp.username", src, dst, map[string]any{"username": username, "source": "mstshash-cookie"}, "credential"))
	}
	response := []byte{0x06, 0xd0, payload[4], payload[5], 0x12, 0x34, 0x00}
	if hasNegotiation {
		response[0] = 0x0e
		response = append(response, 0x02, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00)
	}
	_ = writeTPKT(conn, response)
}

func readTPKT(reader io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, err
	}
	if header[0] != 3 || header[1] != 0 {
		return nil, errors.New("invalid TPKT header")
	}
	length := int(binary.BigEndian.Uint16(header[2:4]))
	if length < 5 || length > maxRDPFrame {
		return nil, errors.New("invalid TPKT length")
	}
	payload := make([]byte, length-4)
	_, err := io.ReadFull(reader, payload)
	return payload, err
}

func writeTPKT(writer io.Writer, payload []byte) error {
	header := []byte{3, 0, 0, 0}
	binary.BigEndian.PutUint16(header[2:4], uint16(len(payload)+4))
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(payload)
	return err
}

func rdpCookieUsername(payload []byte) string {
	text := string(payload)
	for _, line := range strings.Split(text, "\r\n") {
		key, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "Cookie") {
			cookieName, cookieValue, found := strings.Cut(strings.TrimSpace(value), "=")
			if found && strings.EqualFold(strings.TrimSpace(cookieName), "mstshash") {
				return strings.TrimSpace(cookieValue)
			}
		}
	}
	return ""
}

func rdpRequestedProtocols(payload []byte) (uint32, bool) {
	for index := 0; index+8 <= len(payload); index++ {
		if payload[index] == 0x01 && binary.LittleEndian.Uint16(payload[index+2:index+4]) == 8 {
			return binary.LittleEndian.Uint32(payload[index+4 : index+8]), true
		}
	}
	return 0, false
}
