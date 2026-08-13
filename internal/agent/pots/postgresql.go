package pots

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const (
	postgresProtocol30 = 196608
	postgresSSLRequest = 80877103
	maxPostgresMessage = 1 << 20
)

type PostgreSQLService struct {
	listener net.Listener
	once     sync.Once
}

func (s *PostgreSQLService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *PostgreSQLService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *PostgreSQLService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReader(conn)
	startup, err := readPostgresStartup(reader)
	if err != nil {
		return
	}
	if len(startup) == 4 && binary.BigEndian.Uint32(startup) == postgresSSLRequest {
		_, _ = conn.Write([]byte{'N'})
		startup, err = readPostgresStartup(reader)
		if err != nil {
			return
		}
	}
	if len(startup) < 4 || binary.BigEndian.Uint32(startup[:4]) != postgresProtocol30 {
		return
	}
	parameters := parsePostgresParameters(startup[4:])
	if err := writePostgresMessage(conn, 'R', uint32Payload(3)); err != nil {
		return
	}
	messageType, passwordPayload, err := readPostgresMessage(reader)
	if err != nil || messageType != 'p' {
		return
	}
	password := strings.TrimRight(string(passwordPayload), "\x00")
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	sink(protocol.NewEvent("postgresql.credential", src, dst, map[string]any{
		"username": parameters["user"], "password": password, "database": parameters["database"],
		"application_name": parameters["application_name"],
	}, "credential"))
	if !writePostgresReady(conn) {
		return
	}
	for {
		messageType, payload, err := readPostgresMessage(reader)
		if err != nil {
			return
		}
		switch messageType {
		case 'Q':
			query := strings.TrimSpace(strings.TrimRight(string(payload), "\x00"))
			sink(protocol.NewEvent("postgresql.query", src, dst, map[string]any{"query": query, "database": parameters["database"]}, "session"))
			if strings.HasPrefix(strings.ToLower(query), "select") {
				_ = writePostgresTextRow(conn, "PostgreSQL 14.11 on x86_64-pc-linux-gnu")
			} else {
				_ = writePostgresMessage(conn, 'C', append([]byte("OK"), 0))
			}
			_ = writePostgresMessage(conn, 'Z', []byte{'I'})
		case 'X':
			return
		case 'S':
			_ = writePostgresMessage(conn, 'Z', []byte{'I'})
		default:
			_ = writePostgresError(conn, "0A000", "feature not supported")
			_ = writePostgresMessage(conn, 'Z', []byte{'I'})
		}
	}
}

func readPostgresStartup(reader io.Reader) ([]byte, error) {
	var lengthBytes [4]byte
	if _, err := io.ReadFull(reader, lengthBytes[:]); err != nil {
		return nil, err
	}
	length := int(binary.BigEndian.Uint32(lengthBytes[:]))
	if length < 8 || length > maxPostgresMessage {
		return nil, errors.New("invalid PostgreSQL startup length")
	}
	payload := make([]byte, length-4)
	_, err := io.ReadFull(reader, payload)
	return payload, err
}

func readPostgresMessage(reader io.Reader) (byte, []byte, error) {
	var header [5]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return 0, nil, err
	}
	length := int(binary.BigEndian.Uint32(header[1:]))
	if length < 4 || length > maxPostgresMessage {
		return 0, nil, errors.New("invalid PostgreSQL message length")
	}
	payload := make([]byte, length-4)
	_, err := io.ReadFull(reader, payload)
	return header[0], payload, err
}

func writePostgresMessage(writer io.Writer, messageType byte, payload []byte) error {
	header := []byte{messageType, 0, 0, 0, 0}
	binary.BigEndian.PutUint32(header[1:], uint32(len(payload)+4))
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(payload)
	return err
}

func parsePostgresParameters(payload []byte) map[string]string {
	parts := strings.Split(string(payload), "\x00")
	parameters := make(map[string]string, len(parts)/2)
	for index := 0; index+1 < len(parts); index += 2 {
		if parts[index] != "" {
			parameters[parts[index]] = parts[index+1]
		}
	}
	return parameters
}

func uint32Payload(value uint32) []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, value)
	return payload
}

func writePostgresReady(writer io.Writer) bool {
	if writePostgresMessage(writer, 'R', uint32Payload(0)) != nil {
		return false
	}
	if writePostgresMessage(writer, 'S', append(append([]byte("server_version"), 0), append([]byte("14.11"), 0)...)) != nil {
		return false
	}
	if writePostgresMessage(writer, 'S', append(append([]byte("client_encoding"), 0), append([]byte("UTF8"), 0)...)) != nil {
		return false
	}
	backendKey := make([]byte, 8)
	_, _ = rand.Read(backendKey)
	return writePostgresMessage(writer, 'K', backendKey) == nil && writePostgresMessage(writer, 'Z', []byte{'I'}) == nil
}

func writePostgresTextRow(writer io.Writer, value string) error {
	field := make([]byte, 0, 32)
	field = append(field, []byte("result")...)
	field = append(field, 0)
	metadata := make([]byte, 18)
	binary.BigEndian.PutUint32(metadata[6:10], 25)
	binary.BigEndian.PutUint16(metadata[10:12], uint16(0xffff))
	binary.BigEndian.PutUint32(metadata[12:16], uint32(0xffffffff))
	field = append(field, metadata...)
	rowDescription := append([]byte{0, 1}, field...)
	if err := writePostgresMessage(writer, 'T', rowDescription); err != nil {
		return err
	}
	dataRow := make([]byte, 6+len(value))
	binary.BigEndian.PutUint16(dataRow[0:2], 1)
	binary.BigEndian.PutUint32(dataRow[2:6], uint32(len(value)))
	copy(dataRow[6:], value)
	if err := writePostgresMessage(writer, 'D', dataRow); err != nil {
		return err
	}
	return writePostgresMessage(writer, 'C', append([]byte("SELECT 1"), 0))
}

func writePostgresError(writer io.Writer, code, message string) error {
	payload := []byte{'S'}
	payload = append(payload, []byte("ERROR")...)
	payload = append(payload, 0, 'C')
	payload = append(payload, []byte(code)...)
	payload = append(payload, 0, 'M')
	payload = append(payload, []byte(message)...)
	payload = append(payload, 0, 0)
	return writePostgresMessage(writer, 'E', payload)
}
