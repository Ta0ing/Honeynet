package pots

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type RedisService struct {
	listener net.Listener
	once     sync.Once
}

func (s *RedisService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}
func (s *RedisService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}
func (s *RedisService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReader(io.LimitReader(conn, 1<<20))
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		args, err := readRESP(reader)
		if err != nil {
			return
		}
		if len(args) == 0 {
			continue
		}
		command := strings.ToUpper(args[0])
		sink(protocol.NewEvent("redis.command", src, dst, map[string]any{"command": command, "args": args[1:]}, "session"))
		switch command {
		case "PING":
			_, _ = io.WriteString(conn, "+PONG\r\n")
		case "AUTH":
			if len(args) > 1 {
				sink(protocol.NewEvent("redis.credential", src, dst, map[string]any{"password": args[len(args)-1]}, "credential"))
			}
			_, _ = io.WriteString(conn, "+OK\r\n")
		case "INFO":
			value := "# Server\r\nredis_version:6.2.14\r\nos:Linux 5.15.0 x86_64\r\nrole:master\r\n"
			_, _ = fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(value), value)
		case "QUIT":
			_, _ = io.WriteString(conn, "+OK\r\n")
			return
		case "CONFIG", "SET", "GET", "DEL", "EXISTS":
			_, _ = io.WriteString(conn, "+OK\r\n")
		default:
			_, _ = io.WriteString(conn, "-ERR unknown command\r\n")
		}
	}
}
func readRESP(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "*") {
		return strings.Fields(line), nil
	}
	count, err := strconv.Atoi(strings.TrimPrefix(line, "*"))
	if err != nil || count < 0 || count > 128 {
		return nil, fmt.Errorf("invalid RESP array")
	}
	args := make([]string, 0, count)
	for i := 0; i < count; i++ {
		lengthLine, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(lengthLine, "$")))
		if err != nil || length < 0 || length > 64<<10 {
			return nil, fmt.Errorf("invalid RESP bulk string")
		}
		data := make([]byte, length+2)
		if _, err = io.ReadFull(r, data); err != nil {
			return nil, err
		}
		args = append(args, string(data[:length]))
	}
	return args, nil
}
