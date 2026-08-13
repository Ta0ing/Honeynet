package pots

import (
	"context"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/honeynet/honeynet/internal/agent/protocol"
)

type HTTPService struct {
	listener    net.Listener
	server      *http.Server
	secure      bool
	tlsProvider TLSProvider
}

func (s *HTTPService) Start(_ context.Context, target protocol.PotTarget, sink Sink) error {
	listener, err := net.Listen("tcp", listenAddress(target))
	if err != nil {
		return err
	}
	if s.secure {
		listener, err = wrapTLSListener(listener, s.tlsProvider)
		if err != nil {
			return err
		}
	}
	s.listener = listener
	title := configString(target.Config, "title", "Enterprise Management Portal")
	body := configString(target.Config, "body", "")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		raw, _ := io.ReadAll(io.LimitReader(r.Body, 64<<10))
		form := r.URL.Query()
		if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			if parsed, err := url.ParseQuery(string(raw)); err == nil {
				for key, values := range parsed {
					form[key] = values
				}
			}
		}
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		rawRequest, bodyBase64 := httpRequestSnapshot(r, raw, false)
		payload := map[string]any{"method": r.Method, "path": r.URL.RequestURI(), "host": r.Host, "scheme": scheme, "user_agent": r.UserAgent(), "headers": r.Header, "body": string(raw), "raw_request": rawRequest}
		if bodyBase64 != "" {
			payload["body"] = "[binary body]"
			payload["body_base64"] = bodyBase64
		}
		sink(protocol.NewEvent("web.request", endpoint(remoteAddr(r.RemoteAddr)), endpoint(localAddr(r.Context())), payload, "http"))
		username := first(form, "username", "user", "login", "email")
		password := first(form, "password", "pass", "passwd", "pwd")
		if username != "" || password != "" {
			sink(protocol.NewEvent("web.credential", endpoint(remoteAddr(r.RemoteAddr)), endpoint(localAddr(r.Context())), map[string]any{"username": username, "password": password, "path": r.URL.Path}, "credential"))
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Server", "nginx")
		if body != "" {
			_, _ = io.WriteString(w, body)
			return
		}
		_, _ = fmt.Fprintf(w, "<!doctype html><html><head><title>%s</title><style>body{font-family:Arial;background:#f4f6f8}.card{width:360px;margin:12vh auto;background:white;padding:32px;box-shadow:0 8px 30px #ccd}input{width:100%%;padding:10px;margin:8px 0;box-sizing:border-box}button{width:100%%;padding:11px;background:#177ddc;color:white;border:0}</style></head><body><div class=card><h2>%s</h2><form method=post action=/login><input name=username placeholder=Username><input name=password type=password placeholder=Password><button>Sign in</button></form><small>Secure enterprise access</small></div></body></html>", html.EscapeString(title), html.EscapeString(title))
	})
	s.server = &http.Server{Handler: http.MaxBytesHandler(handler, 128<<10), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second, ConnContext: func(ctx context.Context, c net.Conn) context.Context {
		return context.WithValue(ctx, localAddressKey{}, c.LocalAddr())
	}}
	go func() { _ = s.server.Serve(listener) }()
	return nil
}

type localAddressKey struct{}

func localAddr(ctx context.Context) net.Addr {
	value, _ := ctx.Value(localAddressKey{}).(net.Addr)
	return value
}
func remoteAddr(value string) net.Addr { addr, _ := net.ResolveTCPAddr("tcp", value); return addr }
func first(values map[string][]string, keys ...string) string {
	for _, key := range keys {
		for actual, list := range values {
			if strings.EqualFold(actual, key) && len(list) > 0 {
				return list[0]
			}
		}
	}
	return ""
}
func (s *HTTPService) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
