package pots

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPRequestSnapshotReadableAndBinarySafe(t *testing.T) {
	request := httptest.NewRequest("POST", "http://example.test/login?source=pot", strings.NewReader("username=admin&password=test"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	snapshot, encoded := httpRequestSnapshot(request, []byte("username=admin&password=test"), false)
	if encoded != "" || !strings.Contains(snapshot, "POST /login?source=pot HTTP/1.1\r\n") || !strings.Contains(snapshot, "Content-Type: application/x-www-form-urlencoded\r\n") || !strings.HasSuffix(snapshot, "username=admin&password=test") {
		t.Fatalf("unexpected HTTP snapshot: %q, base64=%q", snapshot, encoded)
	}
	binarySnapshot, encoded := httpRequestSnapshot(request, []byte{0xff, 0x00, 0x01}, false)
	if encoded != "/wAB" || !strings.Contains(binarySnapshot, "retained as payload.body_base64") {
		t.Fatalf("binary request body was not retained safely: %q, %q", binarySnapshot, encoded)
	}
}
