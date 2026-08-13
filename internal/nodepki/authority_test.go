package nodepki

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"net"
	"testing"
	"time"
)

func TestAuthorityPersistsCAAndIssuesNodeIdentity(t *testing.T) {
	dir := t.TempDir()
	authority, err := LoadOrCreate(dir, []string{"localhost", "127.0.0.1"}, 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	fingerprint := authority.CAFingerprint()
	if len(authority.CACertificateDER()) == 0 || string(authority.CACertificateDER()) == string(authority.CACertificatePEM()) {
		t.Fatal("CA DER export is empty or not distinct from PEM encoding")
	}
	reloaded, err := LoadOrCreate(dir, []string{"localhost", "127.0.0.1"}, 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.CAFingerprint() != fingerprint {
		t.Fatal("CA changed after reload")
	}
	enrollment, err := authority.IssueNode("node-123", "office node")
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(enrollment.CertificatePEM)
	if block == nil {
		t.Fatal("node certificate is not PEM")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if CertificateNodeID(certificate) != "node-123" || CertificateSerial(certificate) != enrollment.Serial {
		t.Fatalf("unexpected node identity: %q %q", CertificateNodeID(certificate), CertificateSerial(certificate))
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(authority.CACertificatePEM())
	if _, err := certificate.Verify(x509.VerifyOptions{Roots: pool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}); err != nil {
		t.Fatalf("verify node certificate: %v", err)
	}
	caBlock, _ := pem.Decode(authority.CACertificatePEM())
	sum := sha256.Sum256(caBlock.Bytes)
	if hex.EncodeToString(sum[:]) != fingerprint {
		t.Fatal("CA fingerprint does not match certificate")
	}
}

func TestServerTLSRequiresValidPresentedClientCertificate(t *testing.T) {
	authority, err := LoadOrCreate(t.TempDir(), []string{"localhost", "127.0.0.1"}, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	enrollment, err := authority.IssueNode("node-tls", "TLS node")
	if err != nil {
		t.Fatal(err)
	}
	clientCertificate, err := tls.X509KeyPair(enrollment.CertificatePEM, enrollment.PrivateKeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(enrollment.CACertificate)
	listener, err := tls.Listen("tcp", "127.0.0.1:0", authority.ServerTLSConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan error, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			done <- acceptErr
			return
		}
		defer connection.Close()
		done <- connection.(*tls.Conn).Handshake()
	}()
	connection, err := tls.Dial("tcp", listener.Addr().String(), &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: roots, ServerName: "127.0.0.1", Certificates: []tls.Certificate{clientCertificate}})
	if err != nil {
		t.Fatal(err)
	}
	connection.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, _, err := net.SplitHostPort(listener.Addr().String()); err != nil {
		t.Fatal(err)
	}
}

func TestAuthorityIssuesConfiguredLongLivedNodeCertificate(t *testing.T) {
	const validity = 400 * 24 * time.Hour
	authority, err := LoadOrCreate(t.TempDir(), []string{"localhost"}, validity)
	if err != nil {
		t.Fatal(err)
	}
	enrollment, err := authority.IssueNode("node-long-lived", "长期节点")
	if err != nil {
		t.Fatal(err)
	}
	remaining := time.Until(enrollment.ExpiresAt)
	if remaining < 399*24*time.Hour || remaining > validity+time.Hour {
		t.Fatalf("issued certificate lifetime = %s, want about 400 days", remaining)
	}
}
