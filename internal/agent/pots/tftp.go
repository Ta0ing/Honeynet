package pots

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type TFTPService struct {
	conn net.PacketConn
	once sync.Once
}

func (s *TFTPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
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

func (s *TFTPService) Stop() error {
	if s.conn == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.conn.Close() })
	return err
}

func (s *TFTPService) serve(conn net.PacketConn, target protocol.PotTarget, sink Sink) {
	buffer := make([]byte, 4<<10)
	for {
		count, remote, err := conn.ReadFrom(buffer)
		if err != nil {
			return
		}
		if count < 2 {
			continue
		}
		packet := append([]byte(nil), buffer[:count]...)
		opcode := binary.BigEndian.Uint16(packet[:2])
		src, dst := endpoint(remote), endpoint(conn.LocalAddr())
		switch opcode {
		case 1, 2:
			filename, mode, err := parseTFTPRequest(packet[2:])
			if err != nil {
				_, _ = conn.WriteTo(tftpError(0, "invalid request"), remote)
				continue
			}
			operation := "read"
			if opcode == 2 {
				operation = "write"
			}
			sink(protocol.NewEvent("tftp.request", src, dst, map[string]any{"operation": operation, "filename": filename, "mode": mode}, "file-transfer", "recon"))
			if opcode == 1 {
				content := []byte(configString(target.Config, "content", "hostname=edge-router\nconfig_version=2025.03\n"))
				if len(content) > 512 {
					content = content[:512]
				}
				response := []byte{0, 3, 0, 1}
				_, _ = conn.WriteTo(append(response, content...), remote)
			} else {
				_, _ = conn.WriteTo([]byte{0, 4, 0, 0}, remote)
			}
		case 3:
			if len(packet) < 4 {
				continue
			}
			block := binary.BigEndian.Uint16(packet[2:4])
			sink(protocol.NewEvent("tftp.data", src, dst, map[string]any{"block": block, "size": len(packet) - 4}, "file-transfer", "write"))
			ack := []byte{0, 4, packet[2], packet[3]}
			_, _ = conn.WriteTo(ack, remote)
		}
	}
}

func parseTFTPRequest(payload []byte) (string, string, error) {
	filenameEnd := strings.IndexByte(string(payload), 0)
	if filenameEnd <= 0 {
		return "", "", errors.New("TFTP filename is missing")
	}
	rest := payload[filenameEnd+1:]
	modeEnd := strings.IndexByte(string(rest), 0)
	if modeEnd <= 0 {
		return "", "", errors.New("TFTP mode is missing")
	}
	mode := strings.ToLower(string(rest[:modeEnd]))
	if mode != "octet" && mode != "netascii" {
		return "", "", errors.New("unsupported TFTP mode")
	}
	return string(payload[:filenameEnd]), mode, nil
}

func tftpError(code uint16, message string) []byte {
	response := []byte{0, 5, byte(code >> 8), byte(code)}
	response = append(response, message...)
	return append(response, 0)
}
