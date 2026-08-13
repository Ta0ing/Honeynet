package main

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/honeynet/honeynet/internal/nodepki"
)

func TestConsoleTLSAllowsBrowserClientAndEnforcesTLS13(t *testing.T) {
	authority, err := nodepki.LoadOrCreate(t.TempDir(), []string{"127.0.0.1"}, 400*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	tlsConfig := consoleTLSConfig(authority)
	if tlsConfig.ClientAuth != tls.NoClientCert || tlsConfig.ClientCAs != nil {
		t.Fatal("console TLS must not request an Agent client certificate")
	}

	listener, err := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/healthz" {
			http.NotFound(writer, request)
			return
		}
		_, _ = io.WriteString(writer, "ok")
	})}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		_ = server.Close()
		<-done
	})

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(authority.CACertificatePEM()) {
		t.Fatal("failed to load Honeynet CA")
	}
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			RootCAs:    roots,
			ServerName: "127.0.0.1",
		},
		DialContext: (&net.Dialer{Timeout: 2 * time.Second}).DialContext,
	}}
	response, err := client.Get("https://" + listener.Addr().String() + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health response status = %d", response.StatusCode)
	}
	if response.TLS == nil || response.TLS.Version != tls.VersionTLS13 {
		t.Fatalf("negotiated TLS version = %#v, want TLS 1.3", response.TLS)
	}
}

func TestExternalConsoleTLSIsIndependentOfAgentPKI(t *testing.T) {
	tlsConfig := externalConsoleTLSConfig()
	if tlsConfig.MinVersion != tls.VersionTLS13 {
		t.Fatalf("minimum TLS version = %#v, want TLS 1.3", tlsConfig.MinVersion)
	}
	if tlsConfig.ClientAuth != tls.NoClientCert || tlsConfig.ClientCAs != nil {
		t.Fatal("external Console TLS must not request an Agent client certificate")
	}
	if len(tlsConfig.Certificates) != 0 || tlsConfig.GetCertificate != nil {
		t.Fatal("external Console TLS config unexpectedly reuses an Agent PKI certificate")
	}
}
