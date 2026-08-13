package pots

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

const maxZooKeeperPacket = 1 << 20

type ZooKeeperService struct {
	listener net.Listener
	once     sync.Once
}

func (s *ZooKeeperService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *ZooKeeperService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *ZooKeeperService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Minute))
	first := make([]byte, 4)
	if _, err := io.ReadFull(conn, first); err != nil {
		return
	}
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	command := string(first)
	if response, ok := zooKeeperCommandResponse(command); ok {
		sink(protocol.NewEvent("zookeeper.command", src, dst, map[string]any{"command": command}, "coordination", "recon"))
		_, _ = io.WriteString(conn, response)
		return
	}
	length := int(binary.BigEndian.Uint32(first))
	if length < 8 || length > maxZooKeeperPacket {
		return
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return
	}
	request := map[string]any{"bytes": length}
	if len(payload) >= 24 {
		request["protocol_version"] = int32(binary.BigEndian.Uint32(payload[:4]))
		request["last_zxid_seen"] = int64(binary.BigEndian.Uint64(payload[4:12]))
		request["timeout_ms"] = int32(binary.BigEndian.Uint32(payload[12:16]))
		request["session_id"] = int64(binary.BigEndian.Uint64(payload[16:24]))
	}
	sink(protocol.NewEvent("zookeeper.connect", src, dst, request, "coordination", "session"))
	responsePayload := make([]byte, 36)
	binary.BigEndian.PutUint32(responsePayload[:4], 0)
	binary.BigEndian.PutUint32(responsePayload[4:8], 30000)
	binary.BigEndian.PutUint64(responsePayload[8:16], 0x100000001)
	binary.BigEndian.PutUint32(responsePayload[16:20], 16)
	response := make([]byte, 4+len(responsePayload))
	binary.BigEndian.PutUint32(response[:4], uint32(len(responsePayload)))
	copy(response[4:], responsePayload)
	_, _ = conn.Write(response)
}

func zooKeeperCommandResponse(command string) (string, bool) {
	switch command {
	case "ruok":
		return "imok", true
	case "stat", "srvr":
		return "Zookeeper version: 3.8.4, built on 2024-02-12\nLatency min/avg/max: 0/0/0\nMode: follower\nNode count: 42\n", true
	case "envi":
		return "Environment:\nzookeeper.version=3.8.4\nhost.name=zk-01\njava.version=17.0.10\n", true
	case "conf":
		return "clientPort=2181\ndataDir=/var/lib/zookeeper\ntickTime=2000\n", true
	case "mntr":
		return "zk_version\t3.8.4\nzk_server_state\tfollower\nzk_num_alive_connections\t2\nzk_znode_count\t42\n", true
	default:
		return "", false
	}
}
