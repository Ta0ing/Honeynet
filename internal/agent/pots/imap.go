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

type IMAPService struct {
	listener    net.Listener
	once        sync.Once
	secure      bool
	tlsProvider TLSProvider
}

func (s *IMAPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
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

func (s *IMAPService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *IMAPService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReader(io.LimitReader(conn, 256<<10))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	_, _ = io.WriteString(conn, "* OK [CAPABILITY IMAP4rev1 AUTH=PLAIN] mail.corp.local IMAP server ready\r\n")
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		arguments := splitIMAPArguments(strings.TrimSpace(line))
		if len(arguments) < 2 {
			_, _ = io.WriteString(conn, "* BAD Invalid command\r\n")
			continue
		}
		tag, command := arguments[0], strings.ToUpper(arguments[1])
		eventArguments := append([]string(nil), arguments[2:]...)
		if command == "LOGIN" || command == "AUTHENTICATE" {
			eventArguments = []string{"***"}
		}
		sink(protocol.NewEvent("imap.command", src, dst, map[string]any{"tag": tag, "command": command, "arguments": eventArguments}, "mail", "session"))
		switch command {
		case "CAPABILITY":
			_, _ = io.WriteString(conn, "* CAPABILITY IMAP4rev1 AUTH=PLAIN IDLE UIDPLUS NAMESPACE\r\n"+tag+" OK CAPABILITY completed\r\n")
		case "NOOP", "CHECK", "CLOSE":
			_, _ = io.WriteString(conn, tag+" OK "+command+" completed\r\n")
		case "LOGIN":
			if len(arguments) < 4 {
				_, _ = io.WriteString(conn, tag+" BAD LOGIN requires username and password\r\n")
				continue
			}
			sink(protocol.NewEvent("imap.credential", src, dst, map[string]any{"username": arguments[2], "password": arguments[3], "mechanism": "LOGIN"}, "credential"))
			_, _ = io.WriteString(conn, tag+" OK [CAPABILITY IMAP4rev1] LOGIN completed\r\n")
		case "AUTHENTICATE":
			if len(arguments) < 3 || !strings.EqualFold(arguments[2], "PLAIN") {
				_, _ = io.WriteString(conn, tag+" NO Unsupported authentication mechanism\r\n")
				continue
			}
			initial := ""
			if len(arguments) >= 4 {
				initial = arguments[3]
			} else {
				_, _ = io.WriteString(conn, "+ \r\n")
				initial, _ = reader.ReadString('\n')
				initial = strings.TrimSpace(initial)
			}
			decoded, _ := base64.StdEncoding.DecodeString(initial)
			parts := strings.Split(string(decoded), "\x00")
			username, password := "", ""
			if len(parts) >= 3 {
				username, password = parts[len(parts)-2], parts[len(parts)-1]
			}
			sink(protocol.NewEvent("imap.credential", src, dst, map[string]any{"username": username, "password": password, "mechanism": "PLAIN"}, "credential"))
			_, _ = io.WriteString(conn, tag+" OK AUTHENTICATE completed\r\n")
		case "LIST", "LSUB":
			_, _ = io.WriteString(conn, "* LIST (\\HasNoChildren) \"/\" \"INBOX\"\r\n"+tag+" OK LIST completed\r\n")
		case "SELECT", "EXAMINE":
			_, _ = io.WriteString(conn, "* 2 EXISTS\r\n* 0 RECENT\r\n* OK [UNSEEN 1]\r\n* OK [UIDVALIDITY 1735689600]\r\n"+tag+" OK [READ-WRITE] "+command+" completed\r\n")
		case "STATUS":
			_, _ = io.WriteString(conn, "* STATUS INBOX (MESSAGES 2 RECENT 0 UNSEEN 1 UIDNEXT 3)\r\n"+tag+" OK STATUS completed\r\n")
		case "FETCH", "UID":
			_, _ = io.WriteString(conn, "* 1 FETCH (FLAGS (\\Seen) RFC822.SIZE 320 BODY[HEADER] {75}\r\nFrom: security@corp.local\r\nTo: employee@corp.local\r\nSubject: Security notice\r\n\r\n)\r\n"+tag+" OK FETCH completed\r\n")
		case "LOGOUT":
			_, _ = io.WriteString(conn, "* BYE mail.corp.local IMAP server logging out\r\n"+tag+" OK LOGOUT completed\r\n")
			return
		default:
			_, _ = io.WriteString(conn, tag+" BAD Command not supported\r\n")
		}
	}
}

func splitIMAPArguments(line string) []string {
	arguments := make([]string, 0, 6)
	var current strings.Builder
	quoted, escaped := false, false
	flush := func() {
		if current.Len() > 0 {
			arguments = append(arguments, current.String())
			current.Reset()
		}
	}
	for _, value := range line {
		switch {
		case escaped:
			current.WriteRune(value)
			escaped = false
		case quoted && value == '\\':
			escaped = true
		case value == '"':
			quoted = !quoted
		case !quoted && (value == ' ' || value == '\t'):
			flush()
		default:
			current.WriteRune(value)
		}
	}
	flush()
	return arguments
}
