package pots

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type LDAPService struct {
	listener    net.Listener
	once        sync.Once
	secure      bool
	tlsProvider TLSProvider
}

func (s *LDAPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	if s.secure {
		listener, err = wrapTLSListener(listener, s.tlsProvider)
		if err != nil {
			return err
		}
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *LDAPService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *LDAPService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		tag, payload, err := readBERPacket(conn)
		if err != nil {
			return
		}
		if tag != 0x30 {
			return
		}
		messageID, operationTag, operation, err := parseLDAPMessage(payload)
		if err != nil {
			return
		}
		switch operationTag {
		case 0x60:
			dn, password, mechanism, err := parseLDAPBind(operation)
			if err != nil {
				_ = writeLDAPResult(conn, messageID, 0x61, 2, "invalid bind request")
				continue
			}
			sink(protocol.NewEvent("ldap.bind", src, dst, map[string]any{"dn": dn, "mechanism": mechanism}, "directory", "session"))
			if mechanism == "simple" {
				sink(protocol.NewEvent("ldap.credential", src, dst, map[string]any{"username": dn, "password": password, "mechanism": mechanism}, "credential"))
			}
			_ = writeLDAPResult(conn, messageID, 0x61, 0, "")
		case 0x63:
			baseDN := ""
			if valueTag, value, _, parseErr := readBERElement(operation); parseErr == nil && valueTag == 0x04 {
				baseDN = string(value)
			}
			sink(protocol.NewEvent("ldap.search", src, dst, map[string]any{"base_dn": baseDN}, "directory", "recon"))
			_ = writeLDAPResult(conn, messageID, 0x65, 0, "")
		case 0x42:
			sink(protocol.NewEvent("ldap.unbind", src, dst, nil, "directory", "session"))
			return
		default:
			sink(protocol.NewEvent("ldap.operation", src, dst, map[string]any{"operation_tag": operationTag}, "directory", "session"))
			_ = writeLDAPResult(conn, messageID, 0x78, 2, "unsupported operation")
		}
	}
}

func parseLDAPMessage(payload []byte) (int64, byte, []byte, error) {
	tag, messageIDRaw, rest, err := readBERElement(payload)
	if err != nil || tag != 0x02 {
		return 0, 0, nil, errors.New("LDAP message ID is missing")
	}
	messageID, err := parseBERInteger(messageIDRaw)
	if err != nil {
		return 0, 0, nil, err
	}
	operationTag, operation, _, err := readBERElement(rest)
	return messageID, operationTag, operation, err
}

func parseLDAPBind(payload []byte) (string, string, string, error) {
	versionTag, versionRaw, rest, err := readBERElement(payload)
	if err != nil || versionTag != 0x02 {
		return "", "", "", errors.New("LDAP bind version is missing")
	}
	version, err := parseBERInteger(versionRaw)
	if err != nil || version < 2 || version > 3 {
		return "", "", "", errors.New("unsupported LDAP version")
	}
	nameTag, name, rest, err := readBERElement(rest)
	if err != nil || nameTag != 0x04 {
		return "", "", "", errors.New("LDAP bind name is missing")
	}
	authTag, authentication, _, err := readBERElement(rest)
	if err != nil {
		return "", "", "", err
	}
	if authTag == 0x80 {
		return string(name), string(authentication), "simple", nil
	}
	return string(name), "", "sasl", nil
}

func writeLDAPResult(writer io.Writer, messageID int64, operationTag byte, resultCode int64, diagnostic string) error {
	result := append(berElement(0x0a, berIntegerValue(resultCode)), berElement(0x04, nil)...)
	result = append(result, berElement(0x04, []byte(diagnostic))...)
	message := append(berElement(0x02, berIntegerValue(messageID)), berElement(operationTag, result)...)
	_, err := writer.Write(berElement(0x30, message))
	return err
}
