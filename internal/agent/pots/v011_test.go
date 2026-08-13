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

func TestPOP3ServiceCapturesCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &POP3Service{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("pop3", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	reader := bufio.NewReader(conn)
	if line, _ := reader.ReadString('\n'); !strings.HasPrefix(line, "+OK") {
		t.Fatalf("POP3 greeting = %q", line)
	}
	_, _ = io.WriteString(conn, "USER backup\r\n")
	_, _ = reader.ReadString('\n')
	_, _ = io.WriteString(conn, "PASS secret\r\n")
	_, _ = reader.ReadString('\n')
	event := waitType(t, events, "pop3.credential")
	if event.Payload["username"] != "backup" || event.Payload["password"] != "secret" {
		t.Fatalf("POP3 credential = %#v", event.Payload)
	}
}

func TestIMAPServiceCapturesQuotedLogin(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &IMAPService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("imap", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	reader := bufio.NewReader(conn)
	_, _ = reader.ReadString('\n')
	_, _ = io.WriteString(conn, "A001 LOGIN \"backup user\" \"secret value\"\r\n")
	if line, _ := reader.ReadString('\n'); !strings.Contains(line, "A001 OK") {
		t.Fatalf("IMAP LOGIN response = %q", line)
	}
	event := waitType(t, events, "imap.credential")
	if event.Payload["username"] != "backup user" || event.Payload["password"] != "secret value" {
		t.Fatalf("IMAP credential = %#v", event.Payload)
	}
}

func TestLDAPServiceCapturesSimpleBind(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &LDAPService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("ldap", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	bind := append(berElement(0x02, []byte{3}), berElement(0x04, []byte("cn=admin,dc=corp,dc=local"))...)
	bind = append(bind, berElement(0x80, []byte("secret"))...)
	message := append(berElement(0x02, []byte{1}), berElement(0x60, bind)...)
	_, _ = conn.Write(berElement(0x30, message))
	tag, response, err := readBERPacket(conn)
	if err != nil || tag != 0x30 || len(response) == 0 {
		t.Fatalf("LDAP bind response = %x, %v", response, err)
	}
	event := waitType(t, events, "ldap.credential")
	if event.Payload["username"] != "cn=admin,dc=corp,dc=local" || event.Payload["password"] != "secret" {
		t.Fatalf("LDAP credential = %#v", event.Payload)
	}
}

func TestSNMPServiceCapturesCommunityAndAnswers(t *testing.T) {
	events := make(chan protocol.Event, 8)
	packetConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := packetConn.LocalAddr().(*net.UDPAddr).Port
	_ = packetConn.Close()
	service := &SNMPService{}
	if err := service.Start(context.Background(), target("snmp", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	client, err := net.Dial("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(time.Second))
	varBind := append(berElement(0x06, encodeBEROID("1.3.6.1.2.1.1.1.0")), berElement(0x05, nil)...)
	pdu := append(berElement(0x02, []byte{42}), berElement(0x02, []byte{0})...)
	pdu = append(pdu, berElement(0x02, []byte{0})...)
	pdu = append(pdu, berElement(0x30, berElement(0x30, varBind))...)
	message := append(berElement(0x02, []byte{1}), berElement(0x04, []byte("public"))...)
	message = append(message, berElement(0xa0, pdu)...)
	_, _ = client.Write(berElement(0x30, message))
	response := make([]byte, 2048)
	count, err := client.Read(response)
	if err != nil || !bytes.Contains(response[:count], []byte("Linux mail-gateway")) {
		t.Fatalf("SNMP response = %x, %v", response[:count], err)
	}
	event := waitType(t, events, "snmp.community")
	if event.Payload["community"] != "public" {
		t.Fatalf("SNMP community = %#v", event.Payload)
	}
}

func TestRDPServiceCapturesCookieAndNegotiates(t *testing.T) {
	events := make(chan protocol.Event, 8)
	service := &RDPService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("rdp", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	payload := []byte{0, 0xe0, 0, 0, 0x12, 0x34, 0}
	payload = append(payload, []byte("Cookie: mstshash=administrator\r\n")...)
	payload = append(payload, 0x01, 0x00, 0x08, 0x00, 0x03, 0x00, 0x00, 0x00)
	payload[0] = byte(len(payload) - 1)
	if err := writeTPKT(conn, payload); err != nil {
		t.Fatal(err)
	}
	response, err := readTPKT(conn)
	if err != nil || len(response) != 15 || response[1] != 0xd0 || response[7] != 0x02 {
		t.Fatalf("RDP response = %x, %v", response, err)
	}
	event := waitType(t, events, "rdp.username")
	if event.Payload["username"] != "administrator" {
		t.Fatalf("RDP username = %#v", event.Payload)
	}
}

func TestSMBServiceNegotiatesAndCapturesNTLM(t *testing.T) {
	events := make(chan protocol.Event, 16)
	service := &SMBService{}
	port := freePort(t)
	if err := service.Start(context.Background(), target("smb", port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	defer service.Stop()
	conn := dialTCP(t, port)
	defer conn.Close()
	negotiate := smb2TestHeader(0, 1)
	body := make([]byte, 38)
	binary.LittleEndian.PutUint16(body[0:2], 36)
	binary.LittleEndian.PutUint16(body[2:4], 1)
	binary.LittleEndian.PutUint16(body[36:38], 0x0210)
	if err := writeNetBIOSFrame(conn, append(negotiate, body...)); err != nil {
		t.Fatal(err)
	}
	response, err := readNetBIOSFrame(conn)
	if err != nil || len(response) < 128 || binary.LittleEndian.Uint32(response[8:12]) != 0 {
		t.Fatalf("SMB negotiate response = %x, %v", response, err)
	}
	ntlm := testNTLMType3("CORP", "backup", "WORKSTATION")
	session := smb2TestHeader(1, 2)
	sessionBody := make([]byte, 24)
	binary.LittleEndian.PutUint16(sessionBody[0:2], 25)
	binary.LittleEndian.PutUint16(sessionBody[12:14], 88)
	binary.LittleEndian.PutUint16(sessionBody[14:16], uint16(len(ntlm)))
	if err := writeNetBIOSFrame(conn, append(append(session, sessionBody...), ntlm...)); err != nil {
		t.Fatal(err)
	}
	_, _ = readNetBIOSFrame(conn)
	event := waitType(t, events, "smb.authentication")
	if event.Payload["domain"] != "CORP" || event.Payload["username"] != "backup" || event.Payload["workstation"] != "WORKSTATION" {
		t.Fatalf("SMB authentication = %#v", event.Payload)
	}
}

func smb2TestHeader(command uint16, messageID uint64) []byte {
	header := make([]byte, 64)
	copy(header[:4], "\xfeSMB")
	binary.LittleEndian.PutUint16(header[4:6], 64)
	binary.LittleEndian.PutUint16(header[12:14], command)
	binary.LittleEndian.PutUint16(header[14:16], 1)
	binary.LittleEndian.PutUint64(header[24:32], messageID)
	return header
}

func testNTLMType3(domain, username, workstation string) []byte {
	message := make([]byte, 64)
	copy(message, "NTLMSSP\x00")
	binary.LittleEndian.PutUint32(message[8:12], 3)
	binary.LittleEndian.PutUint32(message[60:64], 1)
	appendField := func(descriptor int, raw []byte) {
		binary.LittleEndian.PutUint16(message[descriptor:descriptor+2], uint16(len(raw)))
		binary.LittleEndian.PutUint16(message[descriptor+2:descriptor+4], uint16(len(raw)))
		binary.LittleEndian.PutUint32(message[descriptor+4:descriptor+8], uint32(len(message)))
		message = append(message, raw...)
	}
	appendField(20, []byte{1, 2, 3, 4})
	appendField(28, encodeUTF16LE(domain))
	appendField(36, encodeUTF16LE(username))
	appendField(44, encodeUTF16LE(workstation))
	return message
}
