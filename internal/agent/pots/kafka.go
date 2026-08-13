package pots

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxKafkaRequest = 1 << 20

type KafkaService struct {
	listener net.Listener
	once     sync.Once
}

func (s *KafkaService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *KafkaService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *KafkaService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		var sizeBuffer [4]byte
		if _, err := io.ReadFull(conn, sizeBuffer[:]); err != nil {
			return
		}
		length := int(binary.BigEndian.Uint32(sizeBuffer[:]))
		if length < 10 || length > maxKafkaRequest {
			return
		}
		request := make([]byte, length)
		if _, err := io.ReadFull(conn, request); err != nil {
			return
		}
		apiKey := int16(binary.BigEndian.Uint16(request[:2]))
		apiVersion := int16(binary.BigEndian.Uint16(request[2:4]))
		correlationID := int32(binary.BigEndian.Uint32(request[4:8]))
		clientID := ""
		clientLength := int(int16(binary.BigEndian.Uint16(request[8:10])))
		if clientLength >= 0 && clientLength <= len(request)-10 {
			clientID = string(request[10 : 10+clientLength])
		}
		sink(protocol.NewEvent("kafka.request", src, dst, map[string]any{
			"api_key": apiKey, "api_name": kafkaAPIName(apiKey), "api_version": apiVersion,
			"correlation_id": correlationID, "client_id": clientID, "bytes": length,
		}, "streaming", "session"))
		if _, err := conn.Write(kafkaResponse(apiKey, correlationID)); err != nil {
			return
		}
	}
}

func kafkaAPIName(key int16) string {
	switch key {
	case 0:
		return "Produce"
	case 1:
		return "Fetch"
	case 3:
		return "Metadata"
	case 18:
		return "ApiVersions"
	default:
		return "Unknown"
	}
}

func kafkaResponse(apiKey int16, correlationID int32) []byte {
	body := make([]byte, 4)
	binary.BigEndian.PutUint32(body, uint32(correlationID))
	switch apiKey {
	case 18:
		// ApiVersions v0 response: no error, and a deliberately small but valid
		// API range table for the decoy broker.
		body = append(body, 0, 0, 0, 0, 0, 3)
		for _, version := range [][3]uint16{{0, 0, 9}, {1, 0, 12}, {3, 0, 12}} {
			entry := make([]byte, 6)
			binary.BigEndian.PutUint16(entry[:2], version[0])
			binary.BigEndian.PutUint16(entry[2:4], version[1])
			binary.BigEndian.PutUint16(entry[4:6], version[2])
			body = append(body, entry...)
		}
	case 3:
		// Metadata v0: one broker, no topic metadata.
		body = append(body, 0, 0, 0, 1, 0, 0, 0, 0, 0, 9)
		body = append(body, []byte("localhost")...)
		body = append(body, 0, 0, 0x23, 0x84, 0, 0, 0, 0)
	default:
		body = append(body, 0, 0)
	}
	response := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(response[:4], uint32(len(body)))
	copy(response[4:], body)
	return response
}
