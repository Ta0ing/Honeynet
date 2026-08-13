package pots

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const mysqlCapabilities uint32 = 0x00088201

type MySQLService struct {
	listener net.Listener
	once     sync.Once
}

func (s *MySQLService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}
func (s *MySQLService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}
func (s *MySQLService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	salt := make([]byte, 20)
	_, _ = rand.Read(salt)
	if writeMySQLPacket(conn, 0, mysqlHandshake(salt)) != nil {
		return
	}
	response, seq, err := readMySQLPacket(conn)
	if err != nil {
		return
	}
	username, auth, database := parseMySQLLogin(response)
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	sink(protocol.NewEvent("mysql.credential", src, dst, map[string]any{"username": username, "auth_response": hex.EncodeToString(auth), "database": database}, "credential"))
	if writeMySQLPacket(conn, seq+1, []byte{0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}) != nil {
		return
	}
	for {
		packet, sequence, err := readMySQLPacket(conn)
		if err != nil {
			return
		}
		if len(packet) == 0 {
			continue
		}
		switch packet[0] {
		case 0x01:
			return
		case 0x03:
			query := strings.TrimSpace(string(packet[1:]))
			sink(protocol.NewEvent("mysql.query", src, dst, map[string]any{"query": query}, "session"))
			message := []byte("#42000You have an error in your SQL syntax")
			errorPacket := append([]byte{0xff, 0x28, 0x04}, message...)
			if writeMySQLPacket(conn, sequence+1, errorPacket) != nil {
				return
			}
		case 0x0e:
			if writeMySQLPacket(conn, sequence+1, []byte{0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}) != nil {
				return
			}
		default:
			if writeMySQLPacket(conn, sequence+1, []byte{0xff, 0x48, 0x04, '#', 'H', 'Y', '0', '0', '0', 'U', 'n', 's', 'u', 'p', 'p', 'o', 'r', 't', 'e', 'd'}) != nil {
				return
			}
		}
	}
}
func mysqlHandshake(salt []byte) []byte {
	var b bytes.Buffer
	b.WriteByte(0x0a)
	b.WriteString("8.0.36-0ubuntu0.22.04.1")
	b.WriteByte(0)
	_ = binary.Write(&b, binary.LittleEndian, uint32(1))
	b.Write(salt[:8])
	b.WriteByte(0)
	_ = binary.Write(&b, binary.LittleEndian, uint16(mysqlCapabilities&0xffff))
	b.WriteByte(0xff)
	_ = binary.Write(&b, binary.LittleEndian, uint16(0x0002))
	_ = binary.Write(&b, binary.LittleEndian, uint16(mysqlCapabilities>>16))
	b.WriteByte(21)
	b.Write(make([]byte, 10))
	b.Write(salt[8:])
	b.WriteByte(0)
	b.WriteString("mysql_native_password")
	b.WriteByte(0)
	return b.Bytes()
}
func parseMySQLLogin(packet []byte) (string, []byte, string) {
	if len(packet) < 32 {
		return "", nil, ""
	}
	capabilities := binary.LittleEndian.Uint32(packet[:4])
	offset := 32
	end := bytes.IndexByte(packet[offset:], 0)
	if end < 0 {
		return "", nil, ""
	}
	username := string(packet[offset : offset+end])
	offset += end + 1
	var auth []byte
	if offset < len(packet) {
		length := int(packet[offset])
		offset++
		if offset+length <= len(packet) {
			auth = append([]byte(nil), packet[offset:offset+length]...)
			offset += length
		}
	}
	database := ""
	if capabilities&0x00000008 != 0 && offset < len(packet) {
		end = bytes.IndexByte(packet[offset:], 0)
		if end >= 0 {
			database = string(packet[offset : offset+end])
		}
	}
	return username, auth, database
}
func writeMySQLPacket(w io.Writer, sequence byte, payload []byte) error {
	if len(payload) > 0xffffff {
		return fmt.Errorf("packet too large")
	}
	header := []byte{byte(len(payload)), byte(len(payload) >> 8), byte(len(payload) >> 16), sequence}
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}
func readMySQLPacket(r io.Reader) ([]byte, byte, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, 0, err
	}
	length := int(header[0]) | int(header[1])<<8 | int(header[2])<<16
	if length > 1<<20 {
		return nil, 0, fmt.Errorf("packet too large")
	}
	payload := make([]byte, length)
	_, err := io.ReadFull(r, payload)
	return payload, header[3], err
}
