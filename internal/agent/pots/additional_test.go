package pots

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func TestSupportedCodesCreateServices(t *testing.T) {
	got := SupportedCodes()
	if len(got) != 104 {
		t.Fatalf("SupportedCodes() contains %d services, want 104: %v", len(got), got)
	}
	for _, code := range []string{"tomcat", "edr-sangfor", "oa-tongda", "nas-qnap", "router-openwrt"} {
		if !containsString(got, code) {
			t.Errorf("SupportedCodes() is missing %q", code)
		}
	}
	for _, code := range got {
		if _, err := New(code); err != nil {
			t.Fatalf("New(%q): %v", code, err)
		}
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func TestFTPServiceCapturesCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &FTPService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("ftp", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	if line, _ := reader.ReadString('\n'); !strings.HasPrefix(line, "220 ") {
		t.Fatalf("FTP banner = %q", line)
	}
	_, _ = io.WriteString(conn, "USER backup\r\n")
	_, _ = reader.ReadString('\n')
	_, _ = io.WriteString(conn, "PASS secret\r\n")
	_, _ = reader.ReadString('\n')
	event := waitType(t, events, "ftp.credential")
	if event.Payload["username"] != "backup" || event.Payload["password"] != "secret" {
		t.Fatalf("FTP credential = %#v", event.Payload)
	}
}

func TestSMTPServiceCapturesPlainCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &SMTPService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("smtp", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	_, _ = reader.ReadString('\n')
	_, _ = io.WriteString(conn, "EHLO scanner.local\r\n")
	for index := 0; index < 3; index++ {
		_, _ = reader.ReadString('\n')
	}
	credential := base64.StdEncoding.EncodeToString([]byte("\x00backup\x00secret"))
	_, _ = io.WriteString(conn, "AUTH PLAIN "+credential+"\r\n")
	_, _ = reader.ReadString('\n')
	event := waitType(t, events, "smtp.credential")
	if event.Payload["username"] != "backup" || event.Payload["password"] != "secret" {
		t.Fatalf("SMTP credential = %#v", event.Payload)
	}
}

func TestDNSServiceAnswersAQuery(t *testing.T) {
	events := make(chan protocol.Event, 8)
	packetConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := packetConn.LocalAddr().(*net.UDPAddr).Port
	_ = packetConn.Close()
	service := &DNSService{}
	if err := service.Start(context.Background(), target("dns", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	client, err := net.Dial("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	query := []byte{0x12, 0x34, 0x01, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	for _, label := range strings.Split("db.corp.local", ".") {
		query = append(query, byte(len(label)))
		query = append(query, label...)
	}
	query = append(query, 0, 0, 1, 0, 1)
	_ = client.SetDeadline(time.Now().Add(time.Second))
	if _, err = client.Write(query); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, 512)
	count, err := client.Read(response)
	if err != nil {
		t.Fatal(err)
	}
	if count < 12 || binary.BigEndian.Uint16(response[6:8]) != 1 || !bytes.Equal(response[count-4:count], []byte{192, 0, 2, 10}) {
		t.Fatalf("DNS response = %x", response[:count])
	}
	event := waitType(t, events, "dns.query")
	if event.Payload["name"] != "db.corp.local" || event.Payload["query_type"] != "A" {
		t.Fatalf("DNS event = %#v", event.Payload)
	}
}

func TestPostgreSQLServiceCapturesCredentialAndQuery(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &PostgreSQLService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("postgresql", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	startup := uint32Payload(postgresProtocol30)
	startup = append(startup, []byte("user\x00backup\x00database\x00finance\x00\x00")...)
	packet := make([]byte, 4)
	binary.BigEndian.PutUint32(packet, uint32(len(startup)+4))
	packet = append(packet, startup...)
	if _, err = conn.Write(packet); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	messageType, payload, err := readPostgresMessage(reader)
	if err != nil || messageType != 'R' || binary.BigEndian.Uint32(payload) != 3 {
		t.Fatalf("PostgreSQL auth request = %q %x %v", messageType, payload, err)
	}
	if err = writePostgresMessage(conn, 'p', append([]byte("secret"), 0)); err != nil {
		t.Fatal(err)
	}
	for {
		messageType, _, err = readPostgresMessage(reader)
		if err != nil {
			t.Fatal(err)
		}
		if messageType == 'Z' {
			break
		}
	}
	credential := waitType(t, events, "postgresql.credential")
	if credential.Payload["username"] != "backup" || credential.Payload["password"] != "secret" || credential.Payload["database"] != "finance" {
		t.Fatalf("PostgreSQL credential = %#v", credential.Payload)
	}
	if err = writePostgresMessage(conn, 'Q', append([]byte("SELECT version()"), 0)); err != nil {
		t.Fatal(err)
	}
	event := waitType(t, events, "postgresql.query")
	if event.Payload["query"] != "SELECT version()" {
		t.Fatalf("PostgreSQL query = %#v", event.Payload)
	}
}
