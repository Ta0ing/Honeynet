package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadYAMLAndEnvironmentOverride(t *testing.T) {
	clearEnvironment(t)
	path := filepath.Join(t.TempDir(), "server.yaml")
	data := []byte(`
server:
  addr: ":9090"
  public_url: "https://honeynet.example/"
  tls_enabled: true
  web_dist: "/opt/honeynet/web/dist"
  downloads_dir: "/opt/honeynet/downloads"
database:
  dsn: "user:pass@tcp(mysql:3306)/honeynet?parseTime=true"
auth:
  jwt_secret: "file-secret"
  jwt_expires: "12h"
  admin_username: "secadmin"
  admin_password: "strong-password"
builtin_agent:
  token: "builtin-token"
agent:
  addr: ":9443"
  public_url: "https://agents.honeynet.example:9443"
  pki_dir: "/var/lib/honeynet/pki"
  tls_names: ["agents.honeynet.example", "127.0.0.1"]
  certificate_validity: "1440h"
  renew_before: "240h"
geoip:
  ipip_db_path: "/var/lib/honeynet/ipip.ipdb"
  language: "CN"
threat_intelligence:
  enabled: true
  database_path: "/var/lib/honeynet/threat-intelligence.db"
  download_url: "https://intelligence.example/download/database.zip"
  update_interval: "12h"
detection:
  rules_dir: "/opt/honeynet/rules/builtin"
security:
  redact_sensitive_events: true
ai:
  enabled: true
  provider: "deepseek"
  base_url: "https://api.example/v1/"
  api_key: "test-key"
  model: "deepseek-chat"
  timeout: "20s"
  send_raw_packet: true
cors:
  origins:
    - "https://honeynet.example"
`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HONEYPOT_ADDR", ":8088")
	t.Setenv("HONEYPOT_JWT_EXPIRES", "30m")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8088" || cfg.PublicURL != "https://honeynet.example" {
		t.Fatalf("unexpected server config: %#v", cfg)
	}
	if !cfg.TLSEnabled {
		t.Fatal("expected native console TLS to be enabled")
	}
	foundConsoleName := false
	foundIPv6Loopback := false
	for _, name := range cfg.AgentTLSNames {
		if name == "honeynet.example" {
			foundConsoleName = true
		}
		if name == "::1" {
			foundIPv6Loopback = true
		}
	}
	if !foundConsoleName || !foundIPv6Loopback {
		t.Fatalf("console host was not added to server certificate names: %#v", cfg.AgentTLSNames)
	}
	if cfg.JWTExpires != 30*time.Minute || cfg.AdminUsername != "secadmin" {
		t.Fatalf("unexpected auth config: %#v", cfg)
	}
	if len(cfg.CORSOrigins) != 1 || cfg.CORSOrigins[0] != "https://honeynet.example" {
		t.Fatalf("unexpected CORS config: %#v", cfg.CORSOrigins)
	}
	if cfg.AgentAddr != ":9443" || cfg.AgentPublicURL != "https://agents.honeynet.example:9443" || cfg.AgentCertValidity != 1440*time.Hour || cfg.AgentRenewBefore != 240*time.Hour {
		t.Fatalf("unexpected Agent gateway config: %#v", cfg)
	}
	if cfg.IPIPDBPath != "/var/lib/honeynet/ipip.ipdb" || cfg.IPIPLanguage != "CN" {
		t.Fatalf("unexpected GeoIP config: %#v", cfg)
	}
	if !cfg.ThreatIntelEnabled || cfg.ThreatIntelDBPath != "/var/lib/honeynet/threat-intelligence.db" || cfg.ThreatIntelDownloadURL != "https://intelligence.example/download/database.zip" || cfg.ThreatIntelUpdateInterval != 12*time.Hour {
		t.Fatalf("unexpected threat intelligence config: %#v", cfg)
	}
	if cfg.DetectionRulesDir != "/opt/honeynet/rules/builtin" || !cfg.RedactSensitiveEvents || !cfg.AIEnabled || cfg.AIProvider != "deepseek" || cfg.AIBaseURL != "https://api.example/v1" || cfg.AIAPIKey != "test-key" || cfg.AIModel != "deepseek-chat" || cfg.AITimeout != 20*time.Second || !cfg.AISendRawPacket {
		t.Fatalf("unexpected detection/AI config: %#v", cfg)
	}
}

func TestSensitiveEventRedactionIsOptInAndEnvironmentCanDisableFilePolicy(t *testing.T) {
	clearEnvironment(t)
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedactSensitiveEvents {
		t.Fatal("sensitive event redaction must be disabled when it is not configured")
	}
	t.Setenv("HONEYPOT_REDACT_SENSITIVE_EVENTS", "true")
	cfg, err = Load("")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RedactSensitiveEvents {
		t.Fatal("environment did not enable sensitive event redaction")
	}
	t.Setenv("HONEYPOT_REDACT_SENSITIVE_EVENTS", "")

	path := filepath.Join(t.TempDir(), "server.yaml")
	if err := os.WriteFile(path, []byte("security:\n  redact_sensitive_events: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err = Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RedactSensitiveEvents {
		t.Fatal("YAML did not enable sensitive event redaction")
	}

	t.Setenv("HONEYPOT_REDACT_SENSITIVE_EVENTS", "false")
	cfg, err = Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedactSensitiveEvents {
		t.Fatal("explicit environment false did not override the YAML redaction policy")
	}
}

func TestLoadRejectsInvalidDurationAndUnknownFields(t *testing.T) {
	clearEnvironment(t)
	tests := []string{
		"auth:\n  jwt_expires: never\n",
		"server:\n  mystery: true\n",
	}
	for _, contents := range tests {
		path := filepath.Join(t.TempDir(), "server.yaml")
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatalf("expected configuration error for %q", contents)
		}
	}
}

func TestDefaultNodeCertificateLifetimeExceedsOneYear(t *testing.T) {
	clearEnvironment(t)
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentCertValidity < 366*24*time.Hour {
		t.Fatalf("default Agent certificate validity = %s, want at least 366 days", cfg.AgentCertValidity)
	}
	if cfg.AgentRenewBefore != 30*24*time.Hour {
		t.Fatalf("default Agent renewal window = %s, want 30 days", cfg.AgentRenewBefore)
	}
}

func TestLoadSupportsIPv6ListenersAndPublicURLs(t *testing.T) {
	clearEnvironment(t)
	path := filepath.Join(t.TempDir(), "server.yaml")
	data := []byte(`
server:
  addr: "[::]:8080"
  public_url: "http://[2001:4860:4860::8888]:8080"
database:
  dsn: "user:pass@tcp(127.0.0.1:3306)/honeynet?parseTime=true"
auth:
  jwt_secret: "ipv6-test-secret"
agent:
  addr: "[::]:8443"
  pki_dir: "/var/lib/honeynet/pki"
cors:
  origins: ["http://[2001:4860:4860::8888]:8080"]
`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "[::]:8080" || cfg.AgentAddr != "[::]:8443" {
		t.Fatalf("IPv6 listeners = %q, %q", cfg.Addr, cfg.AgentAddr)
	}
	if cfg.AgentPublicURL != "https://[2001:4860:4860::8888]:8443" {
		t.Fatalf("derived Agent URL = %q", cfg.AgentPublicURL)
	}
	found := false
	for _, name := range cfg.AgentTLSNames {
		if name == "2001:4860:4860::8888" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Agent TLS names do not contain IPv6 host: %#v", cfg.AgentTLSNames)
	}
}

func TestLoadRejectsMalformedIPv6AddressConfiguration(t *testing.T) {
	clearEnvironment(t)
	for name, environment := range map[string]map[string]string{
		"listen without brackets":     {"HONEYPOT_ADDR": "::1:8080"},
		"public URL without brackets": {"HONEYPOT_PUBLIC_URL": "http://2001:4860:4860::8888:8080"},
		"agent URL without brackets":  {"HONEYPOT_AGENT_PUBLIC_URL": "https://2001:4860:4860::8888:8443"},
	} {
		t.Run(name, func(t *testing.T) {
			clearEnvironment(t)
			for key, value := range environment {
				t.Setenv(key, value)
			}
			if _, err := Load(""); err == nil {
				t.Fatal("Load accepted malformed IPv6 configuration")
			}
		})
	}
}

func TestLoadConsoleTLSDefaultsOffAndRequiresHTTPSWhenEnabled(t *testing.T) {
	clearEnvironment(t)
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TLSEnabled {
		t.Fatal("native console TLS must remain opt-in for reverse-proxy and legacy deployments")
	}

	t.Setenv("HONEYPOT_TLS_ENABLED", "true")
	if _, err := Load(""); err == nil {
		t.Fatal("expected HTTP public URL to be rejected when native console TLS is enabled")
	}
	t.Setenv("HONEYPOT_PUBLIC_URL", "https://[2001:db8::20]:8080")
	cfg, err = Load("")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.TLSEnabled {
		t.Fatal("environment override did not enable native console TLS")
	}
	found := false
	for _, name := range cfg.AgentTLSNames {
		if name == "2001:db8::20" {
			found = true
		}
	}
	if !found {
		t.Fatalf("IPv6 console host missing from certificate names: %#v", cfg.AgentTLSNames)
	}

	t.Setenv("HONEYPOT_TLS_ENABLED", "false")
	cfg, err = Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TLSEnabled {
		t.Fatal("explicit false environment override did not disable native TLS")
	}
}

func TestLoadExternalConsoleCertificateFromYAMLAndEnvironment(t *testing.T) {
	clearEnvironment(t)
	certFile, keyFile := writeTestKeyPair(t)
	path := filepath.Join(t.TempDir(), "server.yaml")
	data := []byte(`
server:
  public_url: "https://console.example:8080"
  tls_enabled: true
  tls_cert_file: "/does/not/exist/cert.pem"
  tls_key_file: "/does/not/exist/key.pem"
agent:
  public_url: "https://agents.example:8443"
`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HONEYPOT_TLS_CERT_FILE", certFile)
	t.Setenv("HONEYPOT_TLS_KEY_FILE", keyFile)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.UsesExternalConsoleCertificate() || cfg.TLSCertFile != certFile || cfg.TLSKeyFile != keyFile {
		t.Fatalf("external console certificate config = %#v", cfg)
	}
	foundAgentName := false
	for _, name := range cfg.AgentTLSNames {
		foundAgentName = foundAgentName || name == "agents.example"
	}
	if !foundAgentName {
		t.Fatalf("Agent gateway name missing from Agent PKI names: %#v", cfg.AgentTLSNames)
	}
}

func TestExternalConsoleCertificateKeepsAgentPKINameSetStable(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("HONEYPOT_TLS_ENABLED", "true")
	t.Setenv("HONEYPOT_PUBLIC_URL", "https://console.example:8080")
	privateCfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	certFile, keyFile := writeTestKeyPair(t)
	t.Setenv("HONEYPOT_TLS_CERT_FILE", certFile)
	t.Setenv("HONEYPOT_TLS_KEY_FILE", keyFile)
	externalCfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if len(privateCfg.AgentTLSNames) != len(externalCfg.AgentTLSNames) {
		t.Fatalf("Agent PKI names changed with Console certificate mode: %#v != %#v", privateCfg.AgentTLSNames, externalCfg.AgentTLSNames)
	}
	for index := range privateCfg.AgentTLSNames {
		if privateCfg.AgentTLSNames[index] != externalCfg.AgentTLSNames[index] {
			t.Fatalf("Agent PKI names changed with Console certificate mode: %#v != %#v", privateCfg.AgentTLSNames, externalCfg.AgentTLSNames)
		}
	}
}

func TestLoadRejectsIncompleteOrInvalidExternalConsoleCertificate(t *testing.T) {
	for name, setup := range map[string]func(*testing.T){
		"certificate only": func(t *testing.T) {
			t.Setenv("HONEYPOT_TLS_CERT_FILE", "/tmp/cert.pem")
		},
		"key only": func(t *testing.T) {
			t.Setenv("HONEYPOT_TLS_KEY_FILE", "/tmp/key.pem")
		},
		"TLS disabled": func(t *testing.T) {
			certFile, keyFile := writeTestKeyPair(t)
			t.Setenv("HONEYPOT_TLS_CERT_FILE", certFile)
			t.Setenv("HONEYPOT_TLS_KEY_FILE", keyFile)
		},
		"invalid pair": func(t *testing.T) {
			t.Setenv("HONEYPOT_TLS_ENABLED", "true")
			t.Setenv("HONEYPOT_PUBLIC_URL", "https://console.example:8080")
			t.Setenv("HONEYPOT_TLS_CERT_FILE", "/does/not/exist/cert.pem")
			t.Setenv("HONEYPOT_TLS_KEY_FILE", "/does/not/exist/key.pem")
		},
	} {
		t.Run(name, func(t *testing.T) {
			clearEnvironment(t)
			setup(t)
			if _, err := Load(""); err == nil {
				t.Fatal("Load accepted invalid external Console certificate configuration")
			}
		})
	}
}

func writeTestKeyPair(t *testing.T) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "console.example"},
		DNSNames: []string{"console.example"}, NotBefore: time.Now().Add(-time.Minute),
		NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certFile, keyFile := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0600); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}

func clearEnvironment(t *testing.T) {
	for _, key := range []string{
		"HONEYPOT_ADDR", "HONEYPOT_DATABASE_DSN", "HONEYPOT_JWT_SECRET",
		"HONEYPOT_JWT_EXPIRES", "HONEYPOT_ADMIN_USERNAME", "HONEYPOT_ADMIN_PASSWORD",
		"HONEYPOT_PUBLIC_URL", "HONEYPOT_WEB_DIST", "HONEYPOT_DOWNLOADS_DIR",
		"HONEYPOT_TLS_ENABLED",
		"HONEYPOT_TLS_CERT_FILE", "HONEYPOT_TLS_KEY_FILE",
		"HONEYPOT_BUILTIN_AGENT_TOKEN", "HONEYPOT_CORS_ORIGINS", "HONEYPOT_AGENT_ADDR",
		"HONEYPOT_AGENT_PUBLIC_URL", "HONEYPOT_AGENT_TLS_NAMES", "HONEYPOT_PKI_DIR",
		"HONEYPOT_AGENT_CERT_VALIDITY", "HONEYPOT_AGENT_RENEW_BEFORE",
		"HONEYPOT_IPIP_DB_PATH", "HONEYPOT_IPIP_LANGUAGE",
		"HONEYPOT_THREAT_INTEL_ENABLED", "HONEYPOT_THREAT_INTEL_DB_PATH", "HONEYPOT_THREAT_INTEL_DOWNLOAD_URL",
		"HONEYPOT_THREAT_INTEL_ARCHIVE_PASSWORD", "HONEYPOT_THREAT_INTEL_UPDATE_INTERVAL",
		"HONEYPOT_DETECTION_RULES_DIR", "HONEYPOT_AI_ENABLED", "HONEYPOT_AI_PROVIDER",
		"HONEYPOT_AI_BASE_URL", "HONEYPOT_AI_API_KEY", "HONEYPOT_AI_MODEL", "HONEYPOT_AI_TIMEOUT",
		"HONEYPOT_AI_SEND_RAW_PACKET",
		"HONEYPOT_REDACT_SENSITIVE_EVENTS",
	} {
		t.Setenv(key, "")
	}
}
