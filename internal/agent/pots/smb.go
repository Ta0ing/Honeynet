package pots

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxSMBFrame = 1 << 20

type SMBService struct {
	listener net.Listener
	once     sync.Once
}

func (s *SMBService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *SMBService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *SMBService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		frame, err := readNetBIOSFrame(conn)
		if err != nil {
			return
		}
		switch {
		case len(frame) >= 64 && string(frame[:4]) == "\xfeSMB":
			if !handleSMB2Frame(conn, frame, src, dst, sink) {
				return
			}
		case len(frame) >= 32 && string(frame[:4]) == "\xffSMB":
			command := frame[4]
			sink(protocol.NewEvent("smb.request", src, dst, map[string]any{"version": "SMB1", "command": smb1CommandName(command)}, "file-sharing", "recon"))
			response := append([]byte(nil), frame[:32]...)
			binary.LittleEndian.PutUint32(response[5:9], 0xc00000bb)
			response = append(response, 0, 0, 0)
			if err := writeNetBIOSFrame(conn, response); err != nil {
				return
			}
		default:
			return
		}
	}
}

func readNetBIOSFrame(reader io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, err
	}
	if header[0] != 0 {
		return nil, errors.New("unsupported NetBIOS session message")
	}
	length := int(header[1])<<16 | int(header[2])<<8 | int(header[3])
	if length < 4 || length > maxSMBFrame {
		return nil, errors.New("invalid NetBIOS frame length")
	}
	frame := make([]byte, length)
	_, err := io.ReadFull(reader, frame)
	return frame, err
}

func writeNetBIOSFrame(writer io.Writer, frame []byte) error {
	if len(frame) > 0xffffff {
		return errors.New("SMB frame is too large")
	}
	header := []byte{0, byte(len(frame) >> 16), byte(len(frame) >> 8), byte(len(frame))}
	if _, err := writer.Write(header); err != nil {
		return err
	}
	_, err := writer.Write(frame)
	return err
}

func handleSMB2Frame(conn net.Conn, request []byte, src, dst protocol.Endpoint, sink Sink) bool {
	command := binary.LittleEndian.Uint16(request[12:14])
	sink(protocol.NewEvent("smb.request", src, dst, map[string]any{
		"version": "SMB2", "command": smb2CommandName(command), "message_id": binary.LittleEndian.Uint64(request[24:32]),
	}, "file-sharing", "session"))
	switch command {
	case 0:
		dialects := smb2Dialects(request[64:])
		sink(protocol.NewEvent("smb.negotiate", src, dst, map[string]any{"dialects": dialects}, "file-sharing", "recon"))
		body := make([]byte, 64)
		binary.LittleEndian.PutUint16(body[0:2], 65)
		binary.LittleEndian.PutUint16(body[2:4], 1)
		binary.LittleEndian.PutUint16(body[4:6], 0x0210)
		_, _ = rand.Read(body[8:24])
		binary.LittleEndian.PutUint32(body[24:28], 1)
		binary.LittleEndian.PutUint32(body[28:32], 1<<20)
		binary.LittleEndian.PutUint32(body[32:36], 1<<20)
		binary.LittleEndian.PutUint32(body[36:40], 1<<20)
		binary.LittleEndian.PutUint64(body[40:48], uint64(time.Now().UnixNano()/100)+116444736000000000)
		binary.LittleEndian.PutUint64(body[48:56], uint64(time.Now().Add(-24*time.Hour).UnixNano()/100)+116444736000000000)
		binary.LittleEndian.PutUint16(body[56:58], 128)
		return writeNetBIOSFrame(conn, append(smb2ResponseHeader(request, 0, 0), body...)) == nil
	case 1:
		blob := smb2SecurityBlob(request)
		messageType := ntlmMessageType(blob)
		if messageType == 3 {
			authentication := parseNTLMType3(blob)
			sink(protocol.NewEvent("smb.authentication", src, dst, authentication, "credential"))
			body := make([]byte, 8)
			binary.LittleEndian.PutUint16(body[0:2], 9)
			return writeNetBIOSFrame(conn, append(smb2ResponseHeader(request, 0xc000006d, 0x1000), body...)) == nil
		}
		challenge := ntlmType2Challenge()
		body := make([]byte, 8)
		binary.LittleEndian.PutUint16(body[0:2], 9)
		binary.LittleEndian.PutUint16(body[4:6], 72)
		binary.LittleEndian.PutUint16(body[6:8], uint16(len(challenge)))
		body = append(body, challenge...)
		return writeNetBIOSFrame(conn, append(smb2ResponseHeader(request, 0xc0000016, 0x1000), body...)) == nil
	default:
		body := make([]byte, 8)
		binary.LittleEndian.PutUint16(body[0:2], 9)
		return writeNetBIOSFrame(conn, append(smb2ResponseHeader(request, 0xc00000bb, binary.LittleEndian.Uint64(request[40:48])), body...)) == nil
	}
}

func smb2ResponseHeader(request []byte, status uint32, sessionID uint64) []byte {
	header := make([]byte, 64)
	copy(header[:4], "\xfeSMB")
	binary.LittleEndian.PutUint16(header[4:6], 64)
	binary.LittleEndian.PutUint32(header[8:12], status)
	copy(header[12:16], request[12:16])
	binary.LittleEndian.PutUint32(header[16:20], 1)
	copy(header[24:40], request[24:40])
	if sessionID == 0 {
		sessionID = binary.LittleEndian.Uint64(request[40:48])
	}
	binary.LittleEndian.PutUint64(header[40:48], sessionID)
	return header
}

func smb2Dialects(payload []byte) []string {
	if len(payload) < 36 {
		return nil
	}
	count := int(binary.LittleEndian.Uint16(payload[2:4]))
	if count > 64 || 36+count*2 > len(payload) {
		return nil
	}
	dialects := make([]string, 0, count)
	for index := 0; index < count; index++ {
		value := binary.LittleEndian.Uint16(payload[36+index*2 : 38+index*2])
		dialects = append(dialects, "0x"+hex.EncodeToString([]byte{byte(value >> 8), byte(value)}))
	}
	return dialects
}

func smb2SecurityBlob(request []byte) []byte {
	if len(request) < 88 {
		return nil
	}
	offset := int(binary.LittleEndian.Uint16(request[76:78]))
	length := int(binary.LittleEndian.Uint16(request[78:80]))
	if offset < 64 || length < 0 || offset+length > len(request) {
		return nil
	}
	return request[offset : offset+length]
}

func ntlmMessageType(blob []byte) uint32 {
	index := strings.Index(string(blob), "NTLMSSP\x00")
	if index < 0 || index+12 > len(blob) {
		return 0
	}
	return binary.LittleEndian.Uint32(blob[index+8 : index+12])
}

func ntlmType2Challenge() []byte {
	challenge := make([]byte, 48)
	copy(challenge, "NTLMSSP\x00")
	binary.LittleEndian.PutUint32(challenge[8:12], 2)
	binary.LittleEndian.PutUint32(challenge[20:24], 0x00820201)
	_, _ = rand.Read(challenge[24:32])
	binary.LittleEndian.PutUint32(challenge[44:48], 48)
	return challenge
}

func parseNTLMType3(blob []byte) map[string]any {
	index := strings.Index(string(blob), "NTLMSSP\x00")
	if index < 0 {
		return map[string]any{"mechanism": "NTLM", "malformed": true}
	}
	message := blob[index:]
	readField := func(offset int, unicode bool) string {
		if offset+8 > len(message) {
			return ""
		}
		length := int(binary.LittleEndian.Uint16(message[offset : offset+2]))
		start := int(binary.LittleEndian.Uint32(message[offset+4 : offset+8]))
		if length < 0 || start < 0 || start+length > len(message) {
			return ""
		}
		if unicode {
			return decodeUTF16LE(message[start : start+length])
		}
		return string(message[start : start+length])
	}
	flags := uint32(0)
	if len(message) >= 64 {
		flags = binary.LittleEndian.Uint32(message[60:64])
	}
	unicode := flags&0x00000001 != 0
	ntResponse := ""
	ntResponseTruncated := false
	if len(message) >= 28 {
		length := int(binary.LittleEndian.Uint16(message[20:22]))
		start := int(binary.LittleEndian.Uint32(message[24:28]))
		if length >= 0 && start >= 0 && start+length <= len(message) {
			if length > 512 {
				length = 512
				ntResponseTruncated = true
			}
			ntResponse = hex.EncodeToString(message[start : start+length])
		}
	}
	return map[string]any{
		"mechanism": "NTLM", "domain": readField(28, unicode), "username": readField(36, unicode),
		"workstation": readField(44, unicode), "nt_response": ntResponse, "nt_response_truncated": ntResponseTruncated,
	}
}

func smb2CommandName(command uint16) string {
	names := []string{"NEGOTIATE", "SESSION_SETUP", "LOGOFF", "TREE_CONNECT", "TREE_DISCONNECT", "CREATE", "CLOSE", "FLUSH", "READ", "WRITE", "LOCK", "IOCTL", "CANCEL", "ECHO", "QUERY_DIRECTORY", "CHANGE_NOTIFY", "QUERY_INFO", "SET_INFO", "OPLOCK_BREAK"}
	if int(command) < len(names) {
		return names[command]
	}
	return "UNKNOWN"
}

func smb1CommandName(command byte) string {
	if command == 0x72 {
		return "NEGOTIATE"
	}
	if command == 0x73 {
		return "SESSION_SETUP_ANDX"
	}
	return "UNKNOWN"
}
