package webtemplate

import (
	"net/http"
	"strings"
	"testing"
)

const validTemplate = `name: fake-oa
listen:
  port: 8080
pages:
  - path: /login
    method: post
    capture:
      fields: [username, password, username]
    response:
      status: 302
      headers:
        location: /index
  - path: /index
    response:
      body: Welcome
`

func TestParseNormalizesTemplate(t *testing.T) {
	document, err := Parse(validTemplate)
	if err != nil {
		t.Fatal(err)
	}
	if document.Pages[0].Method != http.MethodPost || document.Pages[0].Capture.EventType != "web.credential" {
		t.Fatalf("unexpected normalized route: %#v", document.Pages[0])
	}
	if len(document.Pages[0].Capture.Fields) != 2 || document.Pages[0].Response.Headers["Location"] != "/index" {
		t.Fatalf("unexpected fields or headers: %#v", document.Pages[0])
	}
	if document.Pages[1].Method != http.MethodGet || document.Pages[1].Response.Status != http.StatusOK {
		t.Fatalf("defaults not applied: %#v", document.Pages[1])
	}
}

func TestParseRejectsUnsafeTemplates(t *testing.T) {
	tests := map[string]string{
		"unknown field":   strings.Replace(validTemplate, "name: fake-oa", "unknown: true\nname: fake-oa", 1),
		"duplicate route": strings.Replace(validTemplate, "  - path: /index", "  - path: /login\n    method: POST", 1),
		"header newline":  strings.Replace(validTemplate, "location: /index", "location: \"/index\\nInjected: yes\"", 1),
		"hop header":      strings.Replace(validTemplate, "location: /index", "content-length: \"99\"", 1),
		"event namespace": strings.Replace(validTemplate, "fields: [username, password, username]", "fields: [username]\n      event_type: ssh.credential", 1),
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(content); err == nil {
				t.Fatal("Parse accepted an unsafe template")
			}
		})
	}
}
