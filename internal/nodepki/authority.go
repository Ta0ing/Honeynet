package nodepki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	caCertificateFile     = "ca.crt"
	caPrivateKeyFile      = "ca.key"
	serverCertificateFile = "server.crt"
	serverPrivateKeyFile  = "server.key"
)

type Authority struct {
	ca       *x509.Certificate
	caKey    *ecdsa.PrivateKey
	caPEM    []byte
	server   tls.Certificate
	validity time.Duration
}

type Enrollment struct {
	CertificatePEM []byte
	PrivateKeyPEM  []byte
	CACertificate  []byte
	Serial         string
	ExpiresAt      time.Time
}

func LoadOrCreate(dir string, serverNames []string, nodeValidity time.Duration) (*Authority, error) {
	if nodeValidity < 24*time.Hour {
		return nil, errors.New("node certificate validity must be at least 24h")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create PKI directory: %w", err)
	}
	ca, key, caPEM, err := loadCA(dir)
	if errors.Is(err, os.ErrNotExist) {
		ca, key, caPEM, err = createCA(dir)
	}
	if err != nil {
		return nil, err
	}
	server, err := loadServerCertificate(dir, ca, serverNames)
	if err != nil {
		server, err = createServerCertificate(dir, ca, key, caPEM, serverNames)
	}
	if err != nil {
		return nil, err
	}
	return &Authority{ca: ca, caKey: key, caPEM: caPEM, server: server, validity: nodeValidity}, nil
}

func (a *Authority) IssueNode(nodeID, nodeName string) (Enrollment, error) {
	if strings.TrimSpace(nodeID) == "" {
		return Enrollment{}, errors.New("node ID is required")
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return Enrollment{}, err
	}
	serial, err := randomSerial()
	if err != nil {
		return Enrollment{}, err
	}
	now := time.Now().UTC()
	identity, _ := url.Parse("spiffe://honeynet/node/" + url.PathEscape(nodeID))
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:         nodeID,
			Organization:       []string{"Honeynet Nodes"},
			OrganizationalUnit: []string{strings.TrimSpace(nodeName)},
		},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(a.validity),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		URIs:                  []*url.URL{identity},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, a.ca, &key.PublicKey, a.caKey)
	if err != nil {
		return Enrollment{}, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return Enrollment{}, err
	}
	return Enrollment{
		CertificatePEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		PrivateKeyPEM:  pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
		CACertificate:  append([]byte(nil), a.caPEM...),
		Serial:         serial.Text(16),
		ExpiresAt:      template.NotAfter,
	}, nil
}

func (a *Authority) ServerTLSConfig() *tls.Config {
	pool := x509.NewCertPool()
	pool.AddCert(a.ca)
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{a.server},
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    pool,
	}
}

func (a *Authority) CAFingerprint() string {
	sum := sha256.Sum256(a.ca.Raw)
	return hex.EncodeToString(sum[:])
}

func (a *Authority) CACertificatePEM() []byte {
	return append([]byte(nil), a.caPEM...)
}

// CACertificateDER returns the public CA certificate in the binary form used
// by Windows Certificate APIs. It never exposes the CA private key.
func (a *Authority) CACertificateDER() []byte {
	return append([]byte(nil), a.ca.Raw...)
}

func CertificateNodeID(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	for _, identity := range cert.URIs {
		if identity.Scheme == "spiffe" && identity.Host == "honeynet" && strings.HasPrefix(identity.Path, "/node/") {
			if value, err := url.PathUnescape(strings.TrimPrefix(identity.Path, "/node/")); err == nil {
				return value
			}
		}
	}
	return cert.Subject.CommonName
}

func CertificateSerial(cert *x509.Certificate) string {
	if cert == nil || cert.SerialNumber == nil {
		return ""
	}
	return strings.ToLower(cert.SerialNumber.Text(16))
}

func loadCA(dir string) (*x509.Certificate, *ecdsa.PrivateKey, []byte, error) {
	certPEM, err := os.ReadFile(filepath.Join(dir, caCertificateFile))
	if err != nil {
		return nil, nil, nil, err
	}
	keyPEM, err := os.ReadFile(filepath.Join(dir, caPrivateKeyFile))
	if err != nil {
		return nil, nil, nil, err
	}
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, nil, errors.New("invalid CA certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil || !cert.IsCA {
		return nil, nil, nil, errors.New("invalid CA certificate")
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, nil, errors.New("invalid CA private key PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse CA private key: %w", err)
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, nil, nil, errors.New("CA private key is not ECDSA")
	}
	return cert, key, certPEM, nil
}

func createCA(dir string) (*x509.Certificate, *ecdsa.PrivateKey, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}
	serial, err := randomSerial()
	if err != nil {
		return nil, nil, nil, err
	}
	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "Honeynet Node CA", Organization: []string{"Honeynet"}},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if err := writePrivateFile(filepath.Join(dir, caPrivateKeyFile), keyPEM); err != nil {
		return nil, nil, nil, err
	}
	if err := writeFile(filepath.Join(dir, caCertificateFile), certPEM, 0644); err != nil {
		return nil, nil, nil, err
	}
	return cert, key, certPEM, nil
}

func loadServerCertificate(dir string, ca *x509.Certificate, names []string) (tls.Certificate, error) {
	certificate, err := tls.LoadX509KeyPair(filepath.Join(dir, serverCertificateFile), filepath.Join(dir, serverPrivateKeyFile))
	if err != nil {
		return tls.Certificate{}, err
	}
	if len(certificate.Certificate) == 0 {
		return tls.Certificate{}, errors.New("server certificate chain is empty")
	}
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		return tls.Certificate{}, err
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: pool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}); err != nil {
		return tls.Certificate{}, err
	}
	if time.Until(leaf.NotAfter) < 30*24*time.Hour {
		return tls.Certificate{}, errors.New("server certificate expires soon")
	}
	for _, name := range names {
		if err := leaf.VerifyHostname(strings.TrimSpace(name)); err != nil {
			return tls.Certificate{}, err
		}
	}
	certificate.Leaf = leaf
	return certificate, nil
}

func createServerCertificate(dir string, ca *x509.Certificate, caKey *ecdsa.PrivateKey, caPEM []byte, names []string) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := randomSerial()
	if err != nil {
		return tls.Certificate{}, err
	}
	dnsNames := make([]string, 0, len(names))
	ipAddresses := make([]net.IP, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if ip := net.ParseIP(name); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		} else {
			dnsNames = append(dnsNames, name)
		}
	}
	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "Honeynet Agent Gateway", Organization: []string{"Honeynet"}},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.AddDate(2, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              dnsNames,
		IPAddresses:           ipAddresses,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), caPEM...)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	if err := writePrivateFile(filepath.Join(dir, serverPrivateKeyFile), keyPEM); err != nil {
		return tls.Certificate{}, err
	}
	if err := writeFile(filepath.Join(dir, serverCertificateFile), certPEM, 0644); err != nil {
		return tls.Certificate{}, err
	}
	return loadServerCertificate(dir, ca, names)
}

func randomSerial() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return nil, err
	}
	if serial.Sign() == 0 {
		return big.NewInt(1), nil
	}
	return serial, nil
}

func writePrivateFile(path string, data []byte) error {
	return writeFile(path, data, 0600)
}

func writeFile(path string, data []byte, mode os.FileMode) error {
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, mode); err != nil {
		return err
	}
	if err := os.Chmod(temporary, mode); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
