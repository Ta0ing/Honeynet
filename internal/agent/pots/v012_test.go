package pots

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/agent/potcert"
	"github.com/honeynet/honeynet/internal/agent/protocol"
)

func testTLSProvider(t *testing.T) *potcert.Manager {
	t.Helper()
	manager, err := potcert.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func startTLSService(t *testing.T, code string, events chan<- protocol.Event) (Service, int) {
	t.Helper()
	service, err := New(code, testTLSProvider(t))
	if err != nil {
		t.Fatal(err)
	}
	port := freePort(t)
	if err = service.Start(context.Background(), target(code, port), func(event protocol.Event) { events <- event }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = service.Stop() })
	return service, port
}

func dialTestTLS(t *testing.T, port int) *tls.Conn {
	t.Helper()
	conn, err := tls.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port), &tls.Config{InsecureSkipVerify: true}) // test-only self-signed decoy certificate
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestHTTPSServiceUsesNodeCertificateAndCapturesCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	_, port := startTLSService(t, "https", events)
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} // test-only self-signed decoy certificate
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	t.Cleanup(transport.CloseIdleConnections)
	form := url.Values{"username": {"tls-admin"}, "password": {"secret"}}
	response, err := client.PostForm("https://127.0.0.1:"+strconv.Itoa(port)+"/login", form)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	request := waitType(t, events, "web.request")
	if request.Payload["scheme"] != "https" {
		t.Fatalf("HTTPS event = %#v", request.Payload)
	}
	credential := waitType(t, events, "web.credential")
	if credential.Payload["username"] != "tls-admin" || credential.Payload["password"] != "secret" {
		t.Fatalf("HTTPS credential = %#v", credential.Payload)
	}
}

func TestSMTPSServiceCapturesCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	_, port := startTLSService(t, "smtps", events)
	conn := dialTestTLS(t, port)
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
		t.Fatalf("SMTPS credential = %#v", event.Payload)
	}
}

func TestPOP3SServiceCapturesCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	_, port := startTLSService(t, "pop3s", events)
	conn := dialTestTLS(t, port)
	reader := bufio.NewReader(conn)
	_, _ = reader.ReadString('\n')
	_, _ = io.WriteString(conn, "USER backup\r\n")
	_, _ = reader.ReadString('\n')
	_, _ = io.WriteString(conn, "PASS secret\r\n")
	_, _ = reader.ReadString('\n')
	event := waitType(t, events, "pop3.credential")
	if event.Payload["username"] != "backup" || event.Payload["password"] != "secret" {
		t.Fatalf("POP3S credential = %#v", event.Payload)
	}
}

func TestIMAPSServiceCapturesCredential(t *testing.T) {
	events := make(chan protocol.Event, 8)
	_, port := startTLSService(t, "imaps", events)
	conn := dialTestTLS(t, port)
	reader := bufio.NewReader(conn)
	_, _ = reader.ReadString('\n')
	_, _ = io.WriteString(conn, "A1 LOGIN backup secret\r\n")
	_, _ = reader.ReadString('\n')
	event := waitType(t, events, "imap.credential")
	if event.Payload["username"] != "backup" || event.Payload["password"] != "secret" {
		t.Fatalf("IMAPS credential = %#v", event.Payload)
	}
}

func TestLDAPSServiceCapturesSimpleBind(t *testing.T) {
	events := make(chan protocol.Event, 8)
	_, port := startTLSService(t, "ldaps", events)
	conn := dialTestTLS(t, port)
	bind := append(berElement(0x02, []byte{3}), berElement(0x04, []byte("cn=backup,dc=corp,dc=local"))...)
	bind = append(bind, berElement(0x80, []byte("secret"))...)
	message := append(berElement(0x02, []byte{1}), berElement(0x60, bind)...)
	if _, err := conn.Write(berElement(0x30, message)); err != nil {
		t.Fatal(err)
	}
	event := waitType(t, events, "ldap.credential")
	if event.Payload["username"] != "cn=backup,dc=corp,dc=local" || event.Payload["password"] != "secret" {
		t.Fatalf("LDAPS credential = %#v", event.Payload)
	}
}

func TestTLSServiceRequiresCertificateProvider(t *testing.T) {
	service, err := New("https")
	if err != nil {
		t.Fatal(err)
	}
	startErr := service.Start(context.Background(), target("https", freePort(t)), func(protocol.Event) {})
	if startErr == nil || !strings.Contains(startErr.Error(), "certificate manager") {
		t.Fatalf("HTTPS without provider error = %v", startErr)
	}
}
