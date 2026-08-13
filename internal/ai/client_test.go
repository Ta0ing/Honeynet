package ai

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenAICompatibleAnalysis(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/chat/completions" || request.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"summary\":\"SQL 注入尝试\",\"confidence\":0.9}"}}]}`)
	}))
	defer server.Close()
	client := New(Config{Enabled: true, Provider: "deepseek", BaseURL: server.URL + "/v1", APIKey: "secret", Model: "test", Timeout: time.Second})
	response, err := client.Analyze(context.Background(), Request{Task: "事件分析", Evidence: map[string]any{"payload": "untrusted"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.JSON) == 0 {
		t.Fatalf("expected structured JSON, got %#v", response)
	}
}

func TestAIProviderRedirectCannotLeakBearerKey(t *testing.T) {
	var redirected bool
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		redirected = true
		if request.Header.Get("Authorization") != "" {
			t.Error("redirected request leaked Authorization")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer plain.Close()
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		http.Redirect(w, request, plain.URL+"/stolen", http.StatusTemporaryRedirect)
	}))
	defer tlsServer.Close()

	client := New(Config{Enabled: true, Provider: "test", BaseURL: tlsServer.URL, APIKey: "never-leak", Model: "model", Timeout: time.Second})
	pool := x509.NewCertPool()
	pool.AddCert(tlsServer.Certificate())
	client.http.Transport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}}
	_, err := client.Analyze(context.Background(), Request{Task: "redirect test", Evidence: map[string]any{"safe": true}})
	if err == nil {
		t.Fatal("provider redirect was followed")
	}
	if redirected {
		t.Fatal("redirect target received a request")
	}
}
