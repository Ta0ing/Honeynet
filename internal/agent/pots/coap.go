package pots

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxCoAPMessage = 64 << 10

type CoAPService struct {
	conn net.PacketConn
	once sync.Once
}

func (s *CoAPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	conn, err := net.ListenPacket("udp", listenAddress(target))
	if err != nil {
		return err
	}
	s.conn = conn
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	go s.serve(conn, target, sink)
	return nil
}

func (s *CoAPService) Stop() error {
	if s.conn == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.conn.Close() })
	return err
}

func (s *CoAPService) serve(conn net.PacketConn, target protocol.PotTarget, sink Sink) {
	buffer := make([]byte, maxCoAPMessage)
	for {
		count, remote, err := conn.ReadFrom(buffer)
		if err != nil {
			return
		}
		request, err := parseCoAPMessage(buffer[:count])
		if err != nil || request.codeClass != 0 || request.codeDetail < 1 || request.codeDetail > 4 {
			continue
		}
		src, dst := endpoint(remote), endpoint(conn.LocalAddr())
		payload := map[string]any{
			"method": coAPMethod(request.codeDetail), "path": request.path, "query": request.query,
			"message_id": request.messageID, "message_type": coAPTypeName(request.messageType),
			"token": hex.EncodeToString(request.token), "content_format": request.contentFormat, "body": string(request.payload),
		}
		sink(protocol.NewEvent("coap.request", src, dst, payload, "iot", "session"))
		if request.codeDetail == 2 || request.codeDetail == 3 || request.codeDetail == 4 {
			sink(protocol.NewEvent("coap.write", src, dst, payload, "iot", "write"))
		}
		responseBody := configString(target.Config, "response", `{"device":"environment-sensor-01","status":"online"}`)
		_, _ = conn.WriteTo(coAPResponse(request, []byte(responseBody)), remote)
	}
}

type coAPMessage struct {
	messageType   byte
	codeClass     byte
	codeDetail    byte
	messageID     uint16
	token         []byte
	path          string
	query         []string
	contentFormat int
	payload       []byte
}

func parseCoAPMessage(packet []byte) (coAPMessage, error) {
	if len(packet) < 4 || packet[0]>>6 != 1 {
		return coAPMessage{}, errors.New("invalid CoAP header")
	}
	tokenLength := int(packet[0] & 0x0f)
	if tokenLength > 8 || 4+tokenLength > len(packet) {
		return coAPMessage{}, errors.New("invalid CoAP token")
	}
	result := coAPMessage{
		messageType: packet[0] >> 4 & 0x03, codeClass: packet[1] >> 5, codeDetail: packet[1] & 0x1f,
		messageID: uint16(packet[2])<<8 | uint16(packet[3]), token: append([]byte(nil), packet[4:4+tokenLength]...),
		contentFormat: -1,
	}
	offset, optionNumber := 4+tokenLength, 0
	paths := make([]string, 0, 4)
	for offset < len(packet) {
		if packet[offset] == 0xff {
			offset++
			result.payload = append([]byte(nil), packet[offset:]...)
			break
		}
		header := packet[offset]
		offset++
		delta, ok := coAPExtendedValue(header>>4, packet, &offset)
		if !ok {
			return coAPMessage{}, errors.New("invalid CoAP option delta")
		}
		length, ok := coAPExtendedValue(header&0x0f, packet, &offset)
		if !ok || offset+length > len(packet) {
			return coAPMessage{}, errors.New("invalid CoAP option length")
		}
		optionNumber += delta
		value := packet[offset : offset+length]
		offset += length
		switch optionNumber {
		case 11:
			paths = append(paths, string(value))
		case 12:
			result.contentFormat = coAPUint(value)
		case 15:
			result.query = append(result.query, string(value))
		}
	}
	result.path = "/" + strings.Join(paths, "/")
	return result, nil
}

func coAPExtendedValue(nibble byte, packet []byte, offset *int) (int, bool) {
	switch {
	case nibble <= 12:
		return int(nibble), true
	case nibble == 13 && *offset < len(packet):
		value := int(packet[*offset]) + 13
		(*offset)++
		return value, true
	case nibble == 14 && *offset+1 < len(packet):
		value := int(packet[*offset])<<8 | int(packet[*offset+1])
		*offset += 2
		return value + 269, true
	default:
		return 0, false
	}
}

func coAPResponse(request coAPMessage, body []byte) []byte {
	messageType := byte(1)
	if request.messageType == 0 {
		messageType = 2
	}
	code := byte(68)
	if request.codeDetail == 1 {
		code = 69
	} else if request.codeDetail == 4 {
		code = 66
	}
	response := []byte{0x40 | messageType<<4 | byte(len(request.token)), code, byte(request.messageID >> 8), byte(request.messageID)}
	response = append(response, request.token...)
	if len(body) > 0 {
		response = append(response, 0xff)
		response = append(response, body...)
	}
	return response
}

func coAPUint(value []byte) int {
	result := 0
	for _, part := range value {
		result = result<<8 | int(part)
	}
	return result
}

func coAPMethod(detail byte) string {
	switch detail {
	case 1:
		return "GET"
	case 2:
		return "POST"
	case 3:
		return "PUT"
	case 4:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

func coAPTypeName(value byte) string {
	switch value {
	case 0:
		return "confirmable"
	case 1:
		return "non_confirmable"
	case 2:
		return "acknowledgement"
	case 3:
		return "reset"
	default:
		return "unknown"
	}
}
