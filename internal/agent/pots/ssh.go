package pots

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
	gossh "golang.org/x/crypto/ssh"
)

type SSHService struct {
	listener net.Listener
	once     sync.Once
}

func (s *SSHService) Start(ctx context.Context, target protocol.PotTarget, sink Sink) error {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	signer, err := gossh.NewSignerFromKey(privateKey)
	if err != nil {
		return err
	}
	config := &gossh.ServerConfig{ServerVersion: "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6", PasswordCallback: func(meta gossh.ConnMetadata, password []byte) (*gossh.Permissions, error) {
		sink(protocol.NewEvent("ssh.credential", endpoint(meta.RemoteAddr()), endpoint(meta.LocalAddr()), map[string]any{"username": meta.User(), "password": string(password)}, "credential"))
		return nil, nil
	}}
	config.AddHostKey(signer)
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	go func() {
		go func() { <-ctx.Done(); _ = listener.Close() }()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go s.handle(conn, config, sink)
		}
	}()
	return nil
}
func (s *SSHService) Stop() error {
	if s.listener == nil {
		return nil
	}
	var err error
	s.once.Do(func() { err = s.listener.Close() })
	return err
}
func (s *SSHService) handle(raw net.Conn, config *gossh.ServerConfig, sink Sink) {
	defer raw.Close()
	_ = raw.SetDeadline(time.Now().Add(5 * time.Minute))
	serverConn, channels, requests, err := gossh.NewServerConn(raw, config)
	if err != nil {
		return
	}
	defer serverConn.Close()
	go gossh.DiscardRequests(requests)
	src, dst := endpoint(serverConn.RemoteAddr()), endpoint(serverConn.LocalAddr())
	sink(protocol.NewEvent("ssh.session", src, dst, map[string]any{"username": serverConn.User(), "client_version": string(serverConn.ClientVersion())}, "session"))
	for newChannel := range channels {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(gossh.UnknownChannelType, "unsupported channel")
			continue
		}
		channel, reqs, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleChannel(channel, reqs, src, dst, sink)
	}
}
func (s *SSHService) handleChannel(channel gossh.Channel, requests <-chan *gossh.Request, src, dst protocol.Endpoint, sink Sink) {
	defer channel.Close()
	started := false
	for req := range requests {
		switch req.Type {
		case "pty-req", "window-change":
			req.Reply(true, nil)
		case "shell":
			req.Reply(true, nil)
			if !started {
				started = true
				s.shell(channel, src, dst, sink)
				return
			}
		case "exec":
			var command struct{ Value string }
			_ = gossh.Unmarshal(req.Payload, &command)
			req.Reply(true, nil)
			sink(protocol.NewEvent("ssh.command", src, dst, map[string]any{"command": command.Value}, "session"))
			_, _ = io.WriteString(channel, commandOutput(command.Value))
			_, _ = channel.SendRequest("exit-status", false, gossh.Marshal(struct{ Status uint32 }{0}))
			return
		default:
			req.Reply(false, nil)
		}
	}
}
func (s *SSHService) shell(channel gossh.Channel, src, dst protocol.Endpoint, sink Sink) {
	_, _ = io.WriteString(channel, "Welcome to Ubuntu 22.04.3 LTS\r\n\r\nadmin@server:~$ ")
	scanner := bufio.NewScanner(io.LimitReader(channel, 1<<20))
	for scanner.Scan() {
		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			_, _ = io.WriteString(channel, "admin@server:~$ ")
			continue
		}
		sink(protocol.NewEvent("ssh.command", src, dst, map[string]any{"command": command}, "session"))
		if command == "exit" || command == "logout" {
			_, _ = io.WriteString(channel, "logout\r\n")
			return
		}
		_, _ = io.WriteString(channel, commandOutput(command))
		_, _ = io.WriteString(channel, "admin@server:~$ ")
	}
}
func commandOutput(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}
	switch command {
	case "id":
		return "uid=1000(admin) gid=1000(admin) groups=1000(admin),27(sudo)\r\n"
	case "whoami":
		return "admin\r\n"
	case "pwd":
		return "/home/admin\r\n"
	case "ls", "ls -la":
		return "total 32\r\ndrwxr-xr-x 4 admin admin 4096 Jan 12 09:32 .\r\n-rw-r--r-- 1 admin admin  220 Jan  6  2022 .bash_logout\r\ndrwxr-xr-x 2 admin admin 4096 Jan 12 09:31 backups\r\n"
	case "uname -a":
		return "Linux server 5.15.0-91-generic #101-Ubuntu SMP x86_64 GNU/Linux\r\n"
	case "cat /etc/os-release":
		return "PRETTY_NAME=\"Ubuntu 22.04.3 LTS\"\r\nNAME=\"Ubuntu\"\r\nVERSION_ID=\"22.04\"\r\n"
	default:
		return fmt.Sprintf("-bash: %s: command not found\r\n", strings.Fields(command)[0])
	}
}
