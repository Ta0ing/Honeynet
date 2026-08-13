package potcert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	certificateLifetime = 30 * 24 * time.Hour
	renewBefore         = 7 * 24 * time.Hour
)

type Manager struct {
	mu       sync.Mutex
	certPath string
	keyPath  string
	hostname string
	now      func() time.Time
	current  *tls.Certificate
}

func New(stateDir string) (*Manager, error) {
	directory := filepath.Join(stateDir, "pot-tls")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	hostname, _ := os.Hostname()
	manager := &Manager{
		certPath: filepath.Join(directory, "server.crt"), keyPath: filepath.Join(directory, "server.key"),
		hostname: hostname, now: time.Now,
	}
	if _, err := manager.certificate(); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *Manager) TLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return m.certificate()
		},
	}
}

func (m *Manager) certificate() (*tls.Certificate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	if m.current != nil && m.current.Leaf != nil && now.Before(m.current.Leaf.NotAfter.Add(-renewBefore)) {
		return m.current, nil
	}
	if loaded, err := m.load(); err == nil && loaded.Leaf != nil && now.After(loaded.Leaf.NotBefore) && now.Before(loaded.Leaf.NotAfter.Add(-renewBefore)) {
		m.current = loaded
		return loaded, nil
	}
	generated, certPEM, keyPEM, err := m.generate(now)
	if err != nil {
		return nil, err
	}
	if err := writeAtomic(m.certPath, certPEM, 0644); err != nil {
		return nil, err
	}
	if err := writeAtomic(m.keyPath, keyPEM, 0600); err != nil {
		return nil, err
	}
	m.current = generated
	return generated, nil
}

func (m *Manager) load() (*tls.Certificate, error) {
	certificate, err := tls.LoadX509KeyPair(m.certPath, m.keyPath)
	if err != nil {
		return nil, err
	}
	if len(certificate.Certificate) == 0 {
		return nil, errors.New("pot TLS certificate chain is empty")
	}
	certificate.Leaf, err = x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		return nil, err
	}
	return &certificate, nil
}

func (m *Manager) generate(now time.Time) (*tls.Certificate, []byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, nil, nil, err
	}
	dnsNames := []string{"localhost"}
	if m.hostname != "" && m.hostname != "localhost" {
		dnsNames = append(dnsNames, m.hostname)
	}
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "Enterprise Secure Gateway", Organization: []string{"Enterprise Infrastructure"}},
		NotBefore:    now.Add(-5 * time.Minute), NotAfter: now.Add(certificateLifetime),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    dnsNames, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, nil, err
	}
	pair.Leaf, err = x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, nil, err
	}
	return &pair, certPEM, keyPEM, nil
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, mode); err != nil {
		return err
	}
	if err := os.Chmod(temporary, mode); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
