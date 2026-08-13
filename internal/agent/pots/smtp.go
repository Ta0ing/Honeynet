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

type SMTPService struct {
	listener    net.Listener
	once        sync.Once
	secure      bool
	tlsProvider TLSProvider
}

func (s *SMTPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
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

func (s *SMTPService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *SMTPService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReader(io.LimitReader(conn, 128<<10))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	_, _ = io.WriteString(conn, "220 mail.corp.local ESMTP Postfix\r\n")
	mailFrom := ""
	recipients := make([]string, 0, 4)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		command, argument, _ := strings.Cut(line, " ")
		command = strings.ToUpper(command)
		argument = strings.TrimSpace(argument)
		switch command {
		case "EHLO", "HELO":
			sink(protocol.NewEvent("smtp.command", src, dst, map[string]any{"command": command, "argument": argument}, "session"))
			_, _ = io.WriteString(conn, "250-mail.corp.local\r\n250-AUTH PLAIN LOGIN\r\n250 SIZE 10485760\r\n")
		case "AUTH":
			if !s.smtpAuth(reader, conn, src, dst, argument, sink) {
				return
			}
		case "MAIL":
			mailFrom = strings.TrimSpace(strings.TrimPrefix(argument, "FROM:"))
			_, _ = io.WriteString(conn, "250 2.1.0 Ok\r\n")
		case "RCPT":
			recipients = append(recipients, strings.TrimSpace(strings.TrimPrefix(argument, "TO:")))
			_, _ = io.WriteString(conn, "250 2.1.5 Ok\r\n")
		case "DATA":
			_, _ = io.WriteString(conn, "354 End data with <CR><LF>.<CR><LF>\r\n")
			message := readSMTPData(reader, 64<<10)
			sink(protocol.NewEvent("smtp.message", src, dst, map[string]any{"from": mailFrom, "to": recipients, "content": message}, "session"))
			mailFrom = ""
			recipients = recipients[:0]
			_, _ = io.WriteString(conn, "250 2.0.0 queued as 7F3A1202\r\n")
		case "RSET", "NOOP":
			mailFrom = ""
			recipients = recipients[:0]
			_, _ = io.WriteString(conn, "250 2.0.0 Ok\r\n")
		case "QUIT":
			_, _ = io.WriteString(conn, "221 2.0.0 Bye\r\n")
			return
		default:
			_, _ = io.WriteString(conn, "502 5.5.2 Command not recognized\r\n")
		}
	}
}

func (s *SMTPService) smtpAuth(reader *bufio.Reader, conn net.Conn, src, dst protocol.Endpoint, argument string, sink Sink) bool {
	mechanism, initial, _ := strings.Cut(argument, " ")
	switch strings.ToUpper(mechanism) {
	case "PLAIN":
		if initial == "" {
			_, _ = io.WriteString(conn, "334 \r\n")
			line, err := reader.ReadString('\n')
			if err != nil {
				return false
			}
			initial = strings.TrimSpace(line)
		}
		decoded, _ := base64.StdEncoding.DecodeString(initial)
		parts := strings.Split(string(decoded), "\x00")
		username, password := "", ""
		if len(parts) >= 3 {
			username, password = parts[len(parts)-2], parts[len(parts)-1]
		}
		sink(protocol.NewEvent("smtp.credential", src, dst, map[string]any{"username": username, "password": password, "mechanism": "PLAIN"}, "credential"))
	case "LOGIN":
		_, _ = io.WriteString(conn, "334 VXNlcm5hbWU6\r\n")
		userLine, err := reader.ReadString('\n')
		if err != nil {
			return false
		}
		_, _ = io.WriteString(conn, "334 UGFzc3dvcmQ6\r\n")
		passwordLine, err := reader.ReadString('\n')
		if err != nil {
			return false
		}
		username, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(userLine))
		password, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(passwordLine))
		sink(protocol.NewEvent("smtp.credential", src, dst, map[string]any{"username": string(username), "password": string(password), "mechanism": "LOGIN"}, "credential"))
	default:
		_, _ = io.WriteString(conn, "504 5.5.4 Unrecognized authentication type\r\n")
		return true
	}
	_, _ = io.WriteString(conn, "235 2.7.0 Authentication successful\r\n")
	return true
}

func readSMTPData(reader *bufio.Reader, limit int) string {
	var builder strings.Builder
	for builder.Len() < limit {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if line == ".\r\n" || line == ".\n" {
			break
		}
		if strings.HasPrefix(line, "..") {
			line = line[1:]
		}
		remaining := limit - builder.Len()
		if len(line) > remaining {
			line = line[:remaining]
		}
		builder.WriteString(line)
	}
	return builder.String()
}
