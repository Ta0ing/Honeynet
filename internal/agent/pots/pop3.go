package pots

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type POP3Service struct {
	listener    net.Listener
	once        sync.Once
	secure      bool
	tlsProvider TLSProvider
}

func (s *POP3Service) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
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

func (s *POP3Service) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *POP3Service) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReader(io.LimitReader(conn, 128<<10))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	_, _ = io.WriteString(conn, "+OK mail.corp.local POP3 server ready <1896.697170952@mail.corp.local>\r\n")
	username := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		command, argument, _ := strings.Cut(line, " ")
		command = strings.ToUpper(command)
		argument = strings.TrimSpace(argument)
		eventArgument := argument
		if command == "PASS" || command == "AUTH" {
			eventArgument = "***"
		}
		sink(protocol.NewEvent("pop3.command", src, dst, map[string]any{"command": command, "argument": eventArgument}, "mail", "session"))
		switch command {
		case "CAPA":
			_, _ = io.WriteString(conn, "+OK Capability list follows\r\nUSER\r\nSASL PLAIN\r\nUIDL\r\nTOP\r\n.\r\n")
		case "USER":
			username = argument
			_, _ = io.WriteString(conn, "+OK Password required\r\n")
		case "PASS":
			sink(protocol.NewEvent("pop3.credential", src, dst, map[string]any{"username": username, "password": argument, "mechanism": "USER/PASS"}, "credential"))
			_, _ = io.WriteString(conn, "+OK Mailbox locked and ready\r\n")
		case "APOP":
			user, digest, _ := strings.Cut(argument, " ")
			sink(protocol.NewEvent("pop3.authentication", src, dst, map[string]any{"username": user, "digest": strings.TrimSpace(digest), "mechanism": "APOP"}, "credential"))
			_, _ = io.WriteString(conn, "+OK Mailbox locked and ready\r\n")
		case "AUTH":
			mechanism, initial, _ := strings.Cut(argument, " ")
			if strings.EqualFold(mechanism, "PLAIN") {
				if initial == "" {
					_, _ = io.WriteString(conn, "+ \r\n")
					initial, _ = reader.ReadString('\n')
					initial = strings.TrimSpace(initial)
				}
				decoded, _ := base64.StdEncoding.DecodeString(initial)
				parts := strings.Split(string(decoded), "\x00")
				user, password := "", ""
				if len(parts) >= 3 {
					user, password = parts[len(parts)-2], parts[len(parts)-1]
				}
				sink(protocol.NewEvent("pop3.credential", src, dst, map[string]any{"username": user, "password": password, "mechanism": "PLAIN"}, "credential"))
				_, _ = io.WriteString(conn, "+OK Authentication successful\r\n")
			} else {
				_, _ = io.WriteString(conn, "-ERR Unsupported authentication mechanism\r\n")
			}
		case "STAT":
			_, _ = io.WriteString(conn, "+OK 2 640\r\n")
		case "LIST":
			_, _ = io.WriteString(conn, "+OK 2 messages\r\n1 320\r\n2 320\r\n.\r\n")
		case "UIDL":
			_, _ = io.WriteString(conn, "+OK\r\n1 WHQTSWQJQYRAA\r\n2 QHDQKJDTYTREW\r\n.\r\n")
		case "RETR", "TOP":
			_, _ = io.WriteString(conn, "+OK 160 octets\r\nFrom: security@corp.local\r\nTo: employee@corp.local\r\nSubject: Security notice\r\n\r\nPlease review the attached security notice.\r\n.\r\n")
		case "DELE", "NOOP", "RSET":
			_, _ = io.WriteString(conn, "+OK\r\n")
		case "QUIT":
			_, _ = io.WriteString(conn, "+OK mail.corp.local signing off\r\n")
			return
		default:
			_, _ = io.WriteString(conn, "-ERR Unknown command\r\n")
		}
	}
}
