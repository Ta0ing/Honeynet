package webtemplate

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	MaxTemplateBytes = 1 << 20
	MaxPages         = 128
	MaxBodyBytes     = 512 << 10
)

type Document struct {
	Name   string `yaml:"name" json:"name"`
	Listen struct {
		Port int `yaml:"port" json:"port"`
	} `yaml:"listen" json:"listen"`
	Pages []Page `yaml:"pages" json:"pages"`
}

type Page struct {
	Path     string   `yaml:"path" json:"path"`
	Method   string   `yaml:"method" json:"method"`
	Capture  Capture  `yaml:"capture" json:"capture"`
	Response Response `yaml:"response" json:"response"`
}

type Capture struct {
	Fields    []string `yaml:"fields" json:"fields"`
	EventType string   `yaml:"event_type" json:"event_type"`
}

type Response struct {
	Status  int               `yaml:"status" json:"status"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Body    string            `yaml:"body" json:"body"`
}

func Parse(content string) (Document, error) {
	if len(content) == 0 || len(content) > MaxTemplateBytes {
		return Document{}, fmt.Errorf("template must be between 1 byte and %d bytes", MaxTemplateBytes)
	}
	decoder := yaml.NewDecoder(strings.NewReader(content))
	decoder.KnownFields(true)
	var document Document
	if err := decoder.Decode(&document); err != nil {
		return Document{}, fmt.Errorf("decode template YAML: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Document{}, errors.New("template must contain exactly one YAML document")
		}
		return Document{}, fmt.Errorf("decode trailing YAML: %w", err)
	}
	if err := normalize(&document); err != nil {
		return Document{}, err
	}
	return document, nil
}

func normalize(document *Document) error {
	document.Name = strings.TrimSpace(document.Name)
	if document.Name == "" || len(document.Name) > 128 {
		return errors.New("name must be between 1 and 128 bytes")
	}
	if document.Listen.Port < 1 || document.Listen.Port > 65535 {
		return errors.New("listen.port must be between 1 and 65535")
	}
	if len(document.Pages) == 0 || len(document.Pages) > MaxPages {
		return fmt.Errorf("pages must contain between 1 and %d routes", MaxPages)
	}
	seen := make(map[string]struct{}, len(document.Pages))
	for index := range document.Pages {
		page := &document.Pages[index]
		page.Path = strings.TrimSpace(page.Path)
		if !validPath(page.Path) {
			return fmt.Errorf("pages[%d].path must be an absolute path without a query or fragment", index)
		}
		page.Method = strings.ToUpper(strings.TrimSpace(page.Method))
		if page.Method == "" {
			page.Method = http.MethodGet
		}
		if !validMethod(page.Method) {
			return fmt.Errorf("pages[%d].method is not supported", index)
		}
		key := page.Method + " " + page.Path
		if _, exists := seen[key]; exists {
			return fmt.Errorf("pages[%d] duplicates route %s", index, key)
		}
		seen[key] = struct{}{}
		if page.Response.Status == 0 {
			page.Response.Status = http.StatusOK
		}
		if page.Response.Status < 200 || page.Response.Status > 599 {
			return fmt.Errorf("pages[%d].response.status must be between 200 and 599", index)
		}
		if len(page.Response.Body) > MaxBodyBytes {
			return fmt.Errorf("pages[%d].response.body exceeds %d bytes", index, MaxBodyBytes)
		}
		if len(page.Response.Headers) > 32 {
			return fmt.Errorf("pages[%d].response.headers exceeds 32 entries", index)
		}
		headers := make(map[string]string, len(page.Response.Headers))
		for name, value := range page.Response.Headers {
			canonical := http.CanonicalHeaderKey(strings.TrimSpace(name))
			if !validHeaderName(canonical) || blockedHeader(canonical) || strings.ContainsAny(value, "\r\n") || len(value) > 4096 {
				return fmt.Errorf("pages[%d].response.headers contains an invalid or unsafe header", index)
			}
			headers[canonical] = value
		}
		page.Response.Headers = headers
		if len(page.Capture.Fields) > 32 {
			return fmt.Errorf("pages[%d].capture.fields exceeds 32 entries", index)
		}
		fields := make([]string, 0, len(page.Capture.Fields))
		fieldSet := map[string]struct{}{}
		for _, raw := range page.Capture.Fields {
			field := strings.TrimSpace(raw)
			if field == "" || len(field) > 128 || strings.ContainsAny(field, "\r\n") {
				return fmt.Errorf("pages[%d].capture.fields contains an invalid field", index)
			}
			if _, exists := fieldSet[field]; !exists {
				fieldSet[field] = struct{}{}
				fields = append(fields, field)
			}
		}
		page.Capture.Fields = fields
		page.Capture.EventType = strings.TrimSpace(page.Capture.EventType)
		if len(fields) > 0 && page.Capture.EventType == "" {
			page.Capture.EventType = "web.credential"
		}
		if page.Capture.EventType != "" && !validEventType(page.Capture.EventType) {
			return fmt.Errorf("pages[%d].capture.event_type must start with web.", index)
		}
	}
	return nil
}

func validPath(value string) bool {
	return len(value) > 0 && len(value) <= 512 && strings.HasPrefix(value, "/") && !strings.ContainsAny(value, "?#\x00\r\n")
}

func validMethod(value string) bool {
	switch value {
	case "*", http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

func validEventType(value string) bool {
	if !strings.HasPrefix(value, "web.") || len(value) > 96 {
		return false
	}
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '.' || char == '_' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func validHeaderName(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", char) {
			continue
		}
		return false
	}
	return true
}

func blockedHeader(value string) bool {
	switch value {
	case "Connection", "Content-Length", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade":
		return true
	default:
		return false
	}
}
