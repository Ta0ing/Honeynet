package pots

import (
	"bufio"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type FTPService struct {
	listener net.Listener
	once     sync.Once
}

func (s *FTPService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *FTPService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *FTPService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReader(io.LimitReader(conn, 64<<10))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	_, _ = io.WriteString(conn, "220 (vsFTPd 3.0.5)\r\n")
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
		sink(protocol.NewEvent("ftp.command", src, dst, map[string]any{"command": command, "argument": redactedFTPArgument(command, argument)}, "session"))
		switch command {
		case "USER":
			username = argument
			_, _ = io.WriteString(conn, "331 Please specify the password.\r\n")
		case "PASS":
			sink(protocol.NewEvent("ftp.credential", src, dst, map[string]any{"username": username, "password": argument}, "credential"))
			_, _ = io.WriteString(conn, "230 Login successful.\r\n")
		case "SYST":
			_, _ = io.WriteString(conn, "215 UNIX Type: L8\r\n")
		case "FEAT":
			_, _ = io.WriteString(conn, "211-Features:\r\n UTF8\r\n EPSV\r\n211 End\r\n")
		case "PWD", "XPWD":
			_, _ = io.WriteString(conn, "257 \"/srv/ftp\" is the current directory\r\n")
		case "TYPE":
			_, _ = io.WriteString(conn, "200 Switching transfer mode.\r\n")
		case "PASV", "EPSV":
			_, _ = io.WriteString(conn, "425 Cannot open data connection.\r\n")
		case "QUIT":
			_, _ = io.WriteString(conn, "221 Goodbye.\r\n")
			return
		default:
			_, _ = io.WriteString(conn, "200 Command okay.\r\n")
		}
	}
}

func redactedFTPArgument(command, argument string) string {
	if command == "PASS" {
		return "***"
	}
	return argument
}
