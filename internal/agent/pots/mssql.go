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
	"unicode/utf16"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const (
	tdsPacketSQLBatch = 0x01
	tdsPacketResponse = 0x04
	tdsPacketLogin7   = 0x10
	tdsPacketPrelogin = 0x12
	maxTDSMessage     = 1 << 20
)

type MSSQLService struct {
	listener net.Listener
	once     sync.Once
}

func (s *MSSQLService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *MSSQLService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *MSSQLService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		packetType, payload, err := readTDSMessage(conn)
		if err != nil {
			return
		}
		switch packetType {
		case tdsPacketPrelogin:
			sink(protocol.NewEvent("mssql.prelogin", src, dst, map[string]any{"size": len(payload)}, "database", "recon"))
			if err := writeTDSPacket(conn, tdsPacketResponse, tdsPreloginResponse()); err != nil {
				return
			}
		case tdsPacketLogin7:
			login, err := parseTDSLogin7(payload)
			if err != nil {
				return
			}
			sink(protocol.NewEvent("mssql.credential", src, dst, map[string]any{
				"username": login.username, "password": login.password, "database": login.database,
				"hostname": login.hostname, "application": login.application, "server": login.server,
			}, "credential"))
			if err := writeTDSPacket(conn, tdsPacketResponse, tdsLoginSuccess()); err != nil {
				return
			}
		case tdsPacketSQLBatch:
			query := strings.TrimSpace(decodeUTF16LE(payload))
			sink(protocol.NewEvent("mssql.query", src, dst, map[string]any{"query": query}, "database", "session"))
			if err := writeTDSPacket(conn, tdsPacketResponse, tdsDoneToken()); err != nil {
				return
			}
		default:
			if err := writeTDSPacket(conn, tdsPacketResponse, tdsDoneToken()); err != nil {
				return
			}
		}
	}
}

func readTDSMessage(reader io.Reader) (byte, []byte, error) {
	var packetType byte
	payload := make([]byte, 0, 4096)
	for {
		var header [8]byte
		if _, err := io.ReadFull(reader, header[:]); err != nil {
			return 0, nil, err
		}
		length := int(binary.BigEndian.Uint16(header[2:4]))
		if length < 8 || len(payload)+length-8 > maxTDSMessage {
			return 0, nil, errors.New("invalid TDS packet length")
		}
		if packetType == 0 {
			packetType = header[0]
		} else if header[0] != packetType {
			return 0, nil, errors.New("mixed TDS packet types")
		}
		fragment := make([]byte, length-8)
		if _, err := io.ReadFull(reader, fragment); err != nil {
			return 0, nil, err
		}
		payload = append(payload, fragment...)
		if header[1]&0x01 != 0 {
			return packetType, payload, nil
		}
	}
}

func writeTDSPacket(writer io.Writer, packetType byte, payload []byte) error {
	if len(payload)+8 > 0xffff {
		return errors.New("TDS response is too large")
	}
	header := []byte{packetType, 0x01, 0, 0, 0, 0, 1, 0}
	binary.BigEndian.PutUint16(header[2:4], uint16(len(payload)+8))
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(payload)
	return err
}

func tdsPreloginResponse() []byte {
	payload := make([]byte, 0, 39)
	for _, option := range []struct {
		token          byte
		offset, length uint16
	}{
		{0x00, 26, 6}, {0x01, 32, 1}, {0x02, 33, 1}, {0x03, 34, 4}, {0x04, 38, 1},
	} {
		payload = append(payload, option.token, byte(option.offset>>8), byte(option.offset), byte(option.length>>8), byte(option.length))
	}
	payload = append(payload, 0xff)
	payload = append(payload, 0x0f, 0x00, 0x07, 0xd0, 0x00, 0x00)
	payload = append(payload, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
	return payload
}

type tdsLogin struct {
	hostname, username, password, application, server, database string
}

func parseTDSLogin7(payload []byte) (tdsLogin, error) {
	if len(payload) < 94 {
		return tdsLogin{}, errors.New("invalid Login7 payload")
	}
	readField := func(pairOffset int) (string, error) {
		offset := int(binary.LittleEndian.Uint16(payload[pairOffset : pairOffset+2]))
		characters := int(binary.LittleEndian.Uint16(payload[pairOffset+2 : pairOffset+4]))
		length := characters * 2
		if offset < 0 || length < 0 || offset+length > len(payload) {
			return "", errors.New("invalid Login7 field")
		}
		return decodeUTF16LE(payload[offset : offset+length]), nil
	}
	hostname, err := readField(36)
	if err != nil {
		return tdsLogin{}, err
	}
	username, err := readField(40)
	if err != nil {
		return tdsLogin{}, err
	}
	passwordOffset := int(binary.LittleEndian.Uint16(payload[44:46]))
	passwordLength := int(binary.LittleEndian.Uint16(payload[46:48])) * 2
	if passwordOffset < 0 || passwordLength < 0 || passwordOffset+passwordLength > len(payload) {
		return tdsLogin{}, errors.New("invalid Login7 password")
	}
	passwordBytes := append([]byte(nil), payload[passwordOffset:passwordOffset+passwordLength]...)
	for index, value := range passwordBytes {
		value ^= 0xa5
		passwordBytes[index] = value<<4 | value>>4
	}
	application, err := readField(48)
	if err != nil {
		return tdsLogin{}, err
	}
	server, err := readField(52)
	if err != nil {
		return tdsLogin{}, err
	}
	database, err := readField(68)
	if err != nil {
		return tdsLogin{}, err
	}
	return tdsLogin{
		hostname: hostname, username: username, password: decodeUTF16LE(passwordBytes),
		application: application, server: server, database: database,
	}, nil
}

func decodeUTF16LE(raw []byte) string {
	if len(raw)%2 != 0 {
		raw = raw[:len(raw)-1]
	}
	values := make([]uint16, len(raw)/2)
	for index := range values {
		values[index] = binary.LittleEndian.Uint16(raw[index*2 : index*2+2])
	}
	return string(utf16.Decode(values))
}

func encodeUTF16LE(value string) []byte {
	values := utf16.Encode([]rune(value))
	raw := make([]byte, len(values)*2)
	for index, item := range values {
		binary.LittleEndian.PutUint16(raw[index*2:index*2+2], item)
	}
	return raw
}

func tdsLoginSuccess() []byte {
	programName := encodeUTF16LE("Microsoft SQL Server")
	loginAck := []byte{0xad, 0, 0, 0x01, 0x74, 0x00, 0x00, 0x04, byte(len(programName) / 2)}
	loginAck = append(loginAck, programName...)
	loginAck = append(loginAck, 0x0f, 0x00, 0x07, 0xd0)
	binary.LittleEndian.PutUint16(loginAck[1:3], uint16(len(loginAck)-3))
	return append(loginAck, tdsDoneToken()...)
}

func tdsDoneToken() []byte {
	return []byte{0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}
