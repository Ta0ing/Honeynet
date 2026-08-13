package pots

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type VNCService struct {
	listener net.Listener
	once     sync.Once
}

func (s *VNCService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go acceptLoop(ctx, listener, func(conn net.Conn) { s.handle(conn, sink) })
	return nil
}

func (s *VNCService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}

func (s *VNCService) handle(conn net.Conn, sink Sink) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Minute))
	if _, err := io.WriteString(conn, "RFB 003.008\n"); err != nil {
		return
	}
	version := make([]byte, 12)
	if _, err := io.ReadFull(conn, version); err != nil || !strings.HasPrefix(string(version), "RFB ") {
		return
	}
	src, dst := endpoint(conn.RemoteAddr()), endpoint(conn.LocalAddr())
	sink(protocol.NewEvent("vnc.connection", src, dst, map[string]any{"version": strings.TrimSpace(string(version))}, "remote-access", "recon"))

	// RFB 3.8: offer VNC password authentication so scanners and clients reveal
	// the challenge response without ever granting a desktop session.
	if _, err := conn.Write([]byte{1, 2}); err != nil {
		return
	}
	selection := make([]byte, 1)
	if _, err := io.ReadFull(conn, selection); err != nil || selection[0] != 2 {
		return
	}
	challenge := make([]byte, 16)
	if _, err := rand.Read(challenge); err != nil {
		return
	}
	if _, err := conn.Write(challenge); err != nil {
		return
	}
	response := make([]byte, 16)
	if _, err := io.ReadFull(conn, response); err != nil {
		return
	}
	sink(protocol.NewEvent("vnc.authentication", src, dst, map[string]any{
		"challenge": hex.EncodeToString(challenge), "response": hex.EncodeToString(response), "result": "denied",
	}, "credential", "remote-access"))
	reason := []byte("Authentication failed")
	result := []byte{0, 0, 0, 1, 0, 0, 0, byte(len(reason))}
	_, _ = conn.Write(append(result, reason...))
}
