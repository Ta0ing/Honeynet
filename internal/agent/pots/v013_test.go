package pots

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func freeUDPPort(t *testing.T) int {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).Port
}

func startProtocolService(t *testing.T, code string, port int, events chan<- protocol.Event) Service {
	t.Helper()
	service, err := New(code)
	if err != nil {
		t.Fatal(err)
	}
	if err = service.Start(context.Background(), target(code, port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = service.Stop() })
	return service
}

func dialProtocolTCP(t *testing.T, port int) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestTFTPServiceAnswersReadRequest(t *testing.T) {
	events := make(chan protocol.Event, 8)
	port := freeUDPPort(t)
	startProtocolService(t, "tftp", port, events)
	conn, err := net.Dial("udp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	request := append([]byte{0, 1}, []byte("router.cfg\x00octet\x00")...)
	if _, err = conn.Write(request); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, 1024)
	count, err := conn.Read(response)
	if err != nil || count < 4 || !bytes.Equal(response[:4], []byte{0, 3, 0, 1}) {
		t.Fatalf("TFTP response = %x, %v", response[:count], err)
	}
	event := waitType(t, events, "tftp.request")
	if event.Payload["filename"] != "router.cfg" || event.Payload["operation"] != "read" {
		t.Fatalf("TFTP event = %#v", event.Payload)
	}
}

func TestVNCServiceCapturesChallengeResponse(t *testing.T) {
	events := make(chan protocol.Event, 8)
	port := freePort(t)
	startProtocolService(t, "vnc", port, events)
	conn := dialProtocolTCP(t, port)
	banner := make([]byte, 12)
	_, _ = io.ReadFull(conn, banner)
	if string(banner) != "RFB 003.008\n" {
		t.Fatalf("VNC banner = %q", banner)
	}
	_, _ = conn.Write([]byte("RFB 003.008\n"))
	security := make([]byte, 2)
	_, _ = io.ReadFull(conn, security)
	if !bytes.Equal(security, []byte{1, 2}) {
		t.Fatalf("VNC security types = %x", security)
	}
	_, _ = conn.Write([]byte{2})
	challenge := make([]byte, 16)
	_, _ = io.ReadFull(conn, challenge)
	_, _ = conn.Write(bytes.Repeat([]byte{0x5a}, 16))
	event := waitType(t, events, "vnc.authentication")
	if event.Payload["response"] != strings.Repeat("5a", 16) || event.Payload["result"] != "denied" {
		t.Fatalf("VNC event = %#v", event.Payload)
	}
}

func TestMemcachedServiceStoresAndReturnsItem(t *testing.T) {
	events := make(chan protocol.Event, 8)
	port := freePort(t)
	startProtocolService(t, "memcached", port, events)
	conn := dialProtocolTCP(t, port)
	reader := bufio.NewReader(conn)
	_, _ = io.WriteString(conn, "set token 7 0 6\r\nsecret\r\n")
	if line, _ := reader.ReadString('\n'); line != "STORED\r\n" {
		t.Fatalf("Memcached set response = %q", line)
	}
	_, _ = io.WriteString(conn, "get token\r\n")
	response := ""
	for !strings.HasSuffix(response, "END\r\n") {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		response += line
	}
	if !strings.Contains(response, "VALUE token 7 6\r\nsecret\r\n") {
		t.Fatalf("Memcached get response = %q", response)
	}
	event := waitType(t, events, "memcached.item")
	if event.Payload["key"] != "token" || event.Payload["value"] != "secret" || event.Payload["stored"] != true {
		t.Fatalf("Memcached event = %#v", event.Payload)
	}
}

func TestOracleServiceCapturesConnectDescriptor(t *testing.T) {
	events := make(chan protocol.Event, 8)
	port := freePort(t)
	startProtocolService(t, "oracle", port, events)
	conn := dialProtocolTCP(t, port)
	descriptor := []byte("(DESCRIPTION=(CONNECT_DATA=(SERVICE_NAME=FINANCE)(USER=SYSTEM)(PROGRAM=sqlplus)))")
	packet := make([]byte, 8+len(descriptor))
	binary.BigEndian.PutUint16(packet[:2], uint16(len(packet)))
	packet[4] = 1
	copy(packet[8:], descriptor)
	_, _ = conn.Write(packet)
	responseHeader := make([]byte, 8)
	if _, err := io.ReadFull(conn, responseHeader); err != nil || responseHeader[4] != 4 {
		t.Fatalf("Oracle response header = %x, %v", responseHeader, err)
	}
	event := waitType(t, events, "oracle.connect")
	if event.Payload["service_name"] != "FINANCE" || event.Payload["user"] != "SYSTEM" || event.Payload["program"] != "sqlplus" {
		t.Fatalf("Oracle event = %#v", event.Payload)
	}
}

func TestZooKeeperServiceAnswersRuok(t *testing.T) {
	events := make(chan protocol.Event, 8)
	port := freePort(t)
	startProtocolService(t, "zookeeper", port, events)
	conn := dialProtocolTCP(t, port)
	_, _ = io.WriteString(conn, "ruok")
	response := make([]byte, 4)
	if _, err := io.ReadFull(conn, response); err != nil || string(response) != "imok" {
		t.Fatalf("ZooKeeper response = %q, %v", response, err)
	}
	event := waitType(t, events, "zookeeper.command")
	if event.Payload["command"] != "ruok" {
		t.Fatalf("ZooKeeper event = %#v", event.Payload)
	}
}

func TestKafkaServiceAnswersAPIVersions(t *testing.T) {
	events := make(chan protocol.Event, 8)
	port := freePort(t)
	startProtocolService(t, "kafka", port, events)
	conn := dialProtocolTCP(t, port)
	requestBody := make([]byte, 14)
	binary.BigEndian.PutUint16(requestBody[:2], 18)
	binary.BigEndian.PutUint16(requestBody[2:4], 0)
	binary.BigEndian.PutUint32(requestBody[4:8], 77)
	binary.BigEndian.PutUint16(requestBody[8:10], 4)
	copy(requestBody[10:], "test")
	frame := make([]byte, 4+len(requestBody))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(requestBody)))
	copy(frame[4:], requestBody)
	_, _ = conn.Write(frame)
	var length [4]byte
	if _, err := io.ReadFull(conn, length[:]); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, binary.BigEndian.Uint32(length[:]))
	if _, err := io.ReadFull(conn, response); err != nil || len(response) < 10 || binary.BigEndian.Uint32(response[:4]) != 77 {
		t.Fatalf("Kafka response = %x, %v", response, err)
	}
	event := waitType(t, events, "kafka.request")
	if event.Payload["api_name"] != "ApiVersions" || event.Payload["client_id"] != "test" || event.Payload["correlation_id"] != int32(77) {
		t.Fatalf("Kafka event = %#v", event.Payload)
	}
}
