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

type TelnetService struct {
	listener net.Listener
	once     sync.Once
}

func (s *TelnetService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}
func (s *TelnetService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}
func (s *TelnetService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReader(io.LimitReader(conn, 64<<10))
	_, _ = io.WriteString(conn, "\xff\xfb\x01\xff\xfb\x03\r\nUbuntu 22.04 LTS\r\nlogin: ")
	user, err := readLine(reader)
	if err != nil {
		return
	}
	_, _ = io.WriteString(conn, "Password: ")
	password, err := readLine(reader)
	if err != nil {
		return
	}
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	if user != "" || password != "" {
		sink(protocol.NewEvent("telnet.credential", src, dst, map[string]any{"username": user, "password": password}, "credential"))
	}
	_, _ = io.WriteString(conn, "\r\nWelcome to Ubuntu 22.04 LTS\r\n$ ")
	for {
		command, err := readLine(reader)
		if err != nil {
			return
		}
		command = strings.TrimSpace(command)
		if command == "" {
			_, _ = io.WriteString(conn, "$ ")
			continue
		}
		sink(protocol.NewEvent("telnet.command", src, dst, map[string]any{"command": command}, "session"))
		switch command {
		case "exit", "logout":
			_, _ = io.WriteString(conn, "logout\r\n")
			return
		case "id":
			_, _ = io.WriteString(conn, "uid=1000(admin) gid=1000(admin) groups=1000(admin)\r\n")
		case "whoami":
			_, _ = io.WriteString(conn, "admin\r\n")
		case "uname -a":
			_, _ = io.WriteString(conn, "Linux server 5.15.0-91-generic #101-Ubuntu SMP x86_64 GNU/Linux\r\n")
		default:
			_, _ = io.WriteString(conn, "-bash: "+command+": command not found\r\n")
		}
		_, _ = io.WriteString(conn, "$ ")
	}
}
func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	return strings.TrimSpace(strings.Trim(line, "\x00\r\n")), err
}
func acceptLoop(ctx context.Context, listener net.Listener, handler func(net.Conn)) {
	go func() { <-ctx.Done(); _ = listener.Close() }()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go handler(conn)
	}
}
