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

const (
	maxMemcachedValue = 1 << 20
	maxMemcachedItems = 4096
)

type memcachedItem struct {
	flags string
	value []byte
}

type MemcachedService struct {
	listener net.Listener
	once     sync.Once
	mu       sync.RWMutex
	items    map[string]memcachedItem
}

func (s *MemcachedService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	s.items = make(map[string]memcachedItem)
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *MemcachedService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *MemcachedService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Minute))
	reader := bufio.NewReaderSize(conn, 16<<10)
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	for {
		line, err := reader.ReadSlice('\n')
		if err != nil {
			return
		}
		fields := strings.Fields(string(line))
		if len(fields) == 0 {
			continue
		}
		command := strings.ToLower(fields[0])
		sink(protocol.NewEvent("memcached.command", src, dst, map[string]any{"command": command, "arguments": fields[1:]}, "cache", "database"))
		switch command {
		case "version":
			_, _ = io.WriteString(conn, "VERSION 1.6.21\r\n")
		case "stats":
			_, _ = io.WriteString(conn, "STAT version 1.6.21\r\nSTAT curr_connections 2\r\nSTAT curr_items 0\r\nEND\r\n")
		case "get", "gets":
			s.writeItems(conn, fields[1:])
		case "set", "add", "replace", "append", "prepend":
			if !s.storeItem(reader, conn, command, fields, src, dst, sink) {
				return
			}
		case "delete":
			if len(fields) < 2 {
				_, _ = io.WriteString(conn, "CLIENT_ERROR bad command line format\r\n")
				continue
			}
			s.mu.Lock()
			_, exists := s.items[fields[1]]
			delete(s.items, fields[1])
			s.mu.Unlock()
			if exists {
				_, _ = io.WriteString(conn, "DELETED\r\n")
			} else {
				_, _ = io.WriteString(conn, "NOT_FOUND\r\n")
			}
		case "incr", "decr":
			_, _ = io.WriteString(conn, "NOT_FOUND\r\n")
		case "quit":
			return
		default:
			_, _ = io.WriteString(conn, "ERROR\r\n")
		}
	}
}

func (s *MemcachedService) storeItem(reader *bufio.Reader, conn net.Conn, command string, fields []string, src, dst protocol.Endpoint, sink Sink) bool {
	if len(fields) < 5 {
		_, _ = io.WriteString(conn, "CLIENT_ERROR bad command line format\r\n")
		return true
	}
	length, err := strconv.Atoi(fields[4])
	if err != nil || length < 0 || length > maxMemcachedValue {
		_, _ = io.WriteString(conn, "CLIENT_ERROR bad data chunk\r\n")
		return false
	}
	data := make([]byte, length+2)
	if _, err = io.ReadFull(reader, data); err != nil || string(data[length:]) != "\r\n" {
		return false
	}
	key := fields[1]
	item := memcachedItem{flags: fields[2], value: append([]byte(nil), data[:length]...)}
	s.mu.Lock()
	_, exists := s.items[key]
	stored := command == "set" || (command == "add" && !exists) || (command == "replace" && exists)
	if command == "append" && exists {
		existing := s.items[key]
		if len(existing.value)+len(item.value) <= maxMemcachedValue {
			item.flags = existing.flags
			item.value = append(append([]byte(nil), existing.value...), item.value...)
			stored = true
		}
	}
	if command == "prepend" && exists {
		existing := s.items[key]
		if len(existing.value)+len(item.value) <= maxMemcachedValue {
			item.flags = existing.flags
			item.value = append(item.value, existing.value...)
			stored = true
		}
	}
	if !exists && len(s.items) >= maxMemcachedItems {
		stored = false
	}
	if stored {
		s.items[key] = item
	}
	s.mu.Unlock()
	sink(protocol.NewEvent("memcached.item", src, dst, map[string]any{
		"command": command, "key": key, "flags": fields[2], "bytes": length, "value": string(data[:length]), "stored": stored,
	}, "cache", "write"))
	if stored {
		_, _ = io.WriteString(conn, "STORED\r\n")
	} else {
		_, _ = io.WriteString(conn, "NOT_STORED\r\n")
	}
	return true
}

func (s *MemcachedService) writeItems(conn net.Conn, keys []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, key := range keys {
		item, ok := s.items[key]
		if !ok {
			continue
		}
		_, _ = fmt.Fprintf(conn, "VALUE %s %s %d\r\n", key, item.flags, len(item.value))
		_, _ = conn.Write(item.value)
		_, _ = io.WriteString(conn, "\r\n")
	}
	_, _ = io.WriteString(conn, "END\r\n")
}
