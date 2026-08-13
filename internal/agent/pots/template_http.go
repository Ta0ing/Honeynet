package pots

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
	"github.com/honeynet/honeynet/internal/webtemplate"
)

const maxTemplateRequestBody = 64 << 10

type TemplateHTTPService struct {
	listener net.Listener
	server   *http.Server
}

func (s *TemplateHTTPService) Start(_ context.Context, target protocol.PotTarget, sink Sink) error {
	if target.Template == nil || target.Template.YAML == "" {
		return errors.New("web template configuration is required")
	}
	document, err := webtemplate.Parse(target.Template.YAML)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	s.listener = listener
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handle(w, r, target, document, sink)
	})
	s.server = &http.Server{
		Handler:           http.MaxBytesHandler(handler, 128<<10),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    64 << 10,
		ConnContext: func(ctx context.Context, connection net.Conn) context.Context {
			return context.WithValue(ctx, localAddressKey{}, connection.LocalAddr())
		},
	}
	go func() { _ = s.server.Serve(listener) }()
	return nil
}

func (s *TemplateHTTPService) handle(w http.ResponseWriter, request *http.Request, target protocol.PotTarget, document webtemplate.Document, sink Sink) {
	defer request.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(request.Body, maxTemplateRequestBody+1))
	truncated := len(raw) > maxTemplateRequestBody
	if truncated {
		raw = raw[:maxTemplateRequestBody]
	}
	page := matchTemplatePage(document, request.Method, request.URL.Path)
	rawRequest, bodyBase64 := httpRequestSnapshot(request, raw, truncated)
	payload := map[string]any{
		"method": request.Method, "path": request.URL.RequestURI(), "host": request.Host,
		"user_agent": request.UserAgent(), "headers": request.Header, "body": string(raw),
		"body_truncated": truncated, "matched": page != nil,
		"template_id": target.Template.ID, "template_name": target.Template.Name,
		"template_version": target.Template.Version,
		"raw_request":      rawRequest,
	}
	if bodyBase64 != "" {
		payload["body"] = "[binary body]"
		payload["body_base64"] = bodyBase64
	}
	sink(protocol.NewEvent("web.request", endpoint(remoteAddr(request.RemoteAddr)), endpoint(localAddr(request.Context())), payload, "http", "template"))

	if page != nil && len(page.Capture.Fields) > 0 {
		values := requestValues(request, raw)
		captured := make(map[string]string, len(page.Capture.Fields))
		for _, field := range page.Capture.Fields {
			if value, exists := lookupValue(values, field); exists {
				captured[field] = value
			}
		}
		if len(captured) > 0 {
			credential := map[string]any{
				"fields": captured, "path": request.URL.Path, "method": request.Method,
				"template_id": target.Template.ID, "template_name": target.Template.Name,
				"template_version": target.Template.Version,
			}
			for key, value := range captured {
				credential[key] = value
			}
			sink(protocol.NewEvent(page.Capture.EventType, endpoint(remoteAddr(request.RemoteAddr)), endpoint(localAddr(request.Context())), credential, "credential", "template"))
		}
	}

	w.Header().Set("Server", "nginx")
	if page == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		if request.Method != http.MethodHead {
			_, _ = io.WriteString(w, "<!doctype html><title>404 Not Found</title><h1>404 Not Found</h1>")
		}
		return
	}
	for name, value := range page.Response.Headers {
		w.Header().Set(name, value)
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	w.WriteHeader(page.Response.Status)
	if request.Method != http.MethodHead {
		_, _ = io.WriteString(w, page.Response.Body)
	}
}

func (s *TemplateHTTPService) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func matchTemplatePage(document webtemplate.Document, method, requestPath string) *webtemplate.Page {
	for index := range document.Pages {
		page := &document.Pages[index]
		if page.Path == requestPath && (page.Method == "*" || page.Method == method || method == http.MethodHead && page.Method == http.MethodGet) {
			return page
		}
	}
	return nil
}

func requestValues(request *http.Request, raw []byte) map[string][]string {
	values := make(map[string][]string, len(request.URL.Query()))
	for key, items := range request.URL.Query() {
		values[key] = append([]string(nil), items...)
	}
	contentType := strings.ToLower(request.Header.Get("Content-Type"))
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if form, err := url.ParseQuery(string(raw)); err == nil {
			for key, items := range form {
				values[key] = append(values[key], items...)
			}
		}
	}
	if strings.Contains(contentType, "application/json") {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var object map[string]any
		if decoder.Decode(&object) == nil {
			for key, value := range object {
				switch item := value.(type) {
				case string:
					values[key] = append(values[key], item)
				case json.Number:
					values[key] = append(values[key], item.String())
				case bool:
					if item {
						values[key] = append(values[key], "true")
					} else {
						values[key] = append(values[key], "false")
					}
				}
			}
		}
	}
	return values
}

func lookupValue(values map[string][]string, field string) (string, bool) {
	for key, items := range values {
		if strings.EqualFold(key, field) && len(items) > 0 {
			return items[len(items)-1], true
		}
	}
	return "", false
}
