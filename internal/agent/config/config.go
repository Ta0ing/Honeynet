package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerURL         string        `json:"server_url"`
	AgentURL          string        `json:"agent_url"`
	NodeID            string        `json:"node_id"`
	RegistrationToken string        `json:"registration_token,omitempty"`
	AgentToken        string        `json:"agent_token,omitempty"` // Legacy field retained only for configuration migration.
	CAFingerprint     string        `json:"ca_sha256,omitempty"`
	CACertPath        string        `json:"ca_cert_path,omitempty"`
	ClientCertPath    string        `json:"client_cert_path,omitempty"`
	ClientKeyPath     string        `json:"client_key_path,omitempty"`
	TLSServerName     string        `json:"tls_server_name,omitempty"`
	CertificateExpiry time.Time     `json:"certificate_expires_at,omitempty"`
	RenewBefore       time.Duration `json:"renew_before,omitempty"`
	UpdatePublicKey   string        `json:"update_public_key,omitempty"`
	UpdateKeyID       string        `json:"update_key_id,omitempty"`
	TemplateRoot      string        `json:"template_root,omitempty"`
	StateDir          string        `json:"state_dir"`
	ConfigPath        string        `json:"-"`
	InsecureTLS       bool          `json:"insecure_tls,omitempty"`
}

func DefaultPath() string {
	if runtime.GOOS == "windows" {
		if base := os.Getenv("ProgramData"); base != "" {
			return filepath.Join(base, "Honeynet", "agent.json")
		}
	}
	return "/etc/honeynet/agent.json"
}

func Load(path string) (Config, error) {
	cfg := Config{ConfigPath: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	cfg.ConfigPath = path
	return cfg, nil
}

func (c *Config) Normalize() error {
	c.ServerURL = strings.TrimRight(strings.TrimSpace(c.ServerURL), "/")
	if c.ServerURL == "" || c.NodeID == "" {
		return errors.New("server_url and node_id are required")
	}
	if err := validateHTTPURL(c.ServerURL, false); err != nil {
		return fmt.Errorf("server_url %w", err)
	}
	if c.ConfigPath == "" {
		c.ConfigPath = DefaultPath()
	}
	if c.StateDir == "" {
		c.StateDir = filepath.Join(filepath.Dir(c.ConfigPath), "state")
	}
	c.TemplateRoot = strings.TrimSpace(c.TemplateRoot)
	if c.TemplateRoot == "" {
		c.TemplateRoot = DefaultTemplateRoot()
	}
	c.AgentURL = strings.TrimRight(strings.TrimSpace(c.AgentURL), "/")
	if c.AgentURL == "" {
		c.AgentURL = deriveAgentURL(c.ServerURL)
	}
	if err := validateHTTPURL(c.AgentURL, true); err != nil {
		return fmt.Errorf("agent_url %w", err)
	}
	c.CAFingerprint = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(c.CAFingerprint), ":", ""))
	if c.CAFingerprint != "" && len(c.CAFingerprint) != 64 {
		return errors.New("ca_sha256 must be a SHA-256 hex digest")
	}
	if c.RenewBefore <= 0 {
		c.RenewBefore = 7 * 24 * time.Hour
	}
	return nil
}

// DefaultTemplateRoot locates the platform-neutral Web resource
// pack. Source checkouts are supported for native development; installed
// Agents use the fixed product directory and do not require a container.
func DefaultTemplateRoot() string {
	if configured := strings.TrimSpace(os.Getenv("HONEYPOT_TEMPLATE_ROOT")); configured != "" {
		return configured
	}
	if executable, err := os.Executable(); err == nil {
		candidate := filepath.Clean(filepath.Join(filepath.Dir(executable), "..", "templates", "web", "services"))
		if templateManifestExists(candidate) {
			return candidate
		}
	}
	if working, err := os.Getwd(); err == nil {
		current := working
		for range 8 {
			candidate := filepath.Join(current, "honeypot-templates-server", "services")
			if templateManifestExists(candidate) {
				return candidate
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}
	if runtime.GOOS == "windows" {
		base := os.Getenv("ProgramFiles")
		if base == "" {
			base = `C:\Program Files`
		}
		return filepath.Join(base, "Honeynet", "templates", "web", "services")
	}
	return "/opt/honeynet-agent/templates/web/services"
}

func templateManifestExists(root string) bool {
	info, err := os.Stat(filepath.Join(root, "config.json"))
	return err == nil && !info.IsDir()
}

func deriveAgentURL(serverURL string) string {
	parsed, err := url.Parse(serverURL)
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	parsed.Scheme = "https"
	parsed.Host = net.JoinHostPort(parsed.Hostname(), "8443")
	parsed.Path, parsed.RawPath, parsed.RawQuery, parsed.Fragment = "", "", "", ""
	return parsed.String()
}

func validateHTTPURL(value string, httpsOnly bool) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil {
		return errors.New("must be an absolute HTTP URL without user information")
	}
	if httpsOnly && parsed.Scheme != "https" {
		return errors.New("must use https")
	}
	if !httpsOnly && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("must use http or https")
	}
	if portText := parsed.Port(); portText != "" {
		port, err := strconv.Atoi(portText)
		if err != nil || port < 1 || port > 65535 {
			return errors.New("port must be between 1 and 65535")
		}
	}
	if strings.Contains(parsed.Hostname(), ":") && !strings.HasPrefix(parsed.Host, "[") {
		return errors.New("IPv6 host must be enclosed in brackets")
	}
	return nil
}

func (c *Config) HasCertificate() bool {
	return c.CACertPath != "" && c.ClientCertPath != "" && c.ClientKeyPath != "" && !c.CertificateExpiry.IsZero()
}

func (c *Config) SaveEnrollment(agentURL string, caCertificate, clientCertificate, clientKey []byte, expiresAt time.Time, renewBefore time.Duration) error {
	if agentURL = strings.TrimRight(strings.TrimSpace(agentURL), "/"); agentURL != "" {
		c.AgentURL = agentURL
	}
	pkiDir := filepath.Join(c.StateDir, "pki")
	if err := os.MkdirAll(pkiDir, 0700); err != nil {
		return err
	}
	caPath := filepath.Join(pkiDir, "ca.crt")
	certPath := filepath.Join(pkiDir, "client.crt")
	keyPath := filepath.Join(pkiDir, "client.key")
	for _, file := range []struct {
		path string
		data []byte
		mode os.FileMode
	}{{caPath, caCertificate, 0644}, {certPath, clientCertificate, 0644}, {keyPath, clientKey, 0600}} {
		if err := writeAtomic(file.path, file.data, file.mode); err != nil {
			return err
		}
	}
	c.CACertPath = caPath
	c.ClientCertPath = certPath
	c.ClientKeyPath = keyPath
	c.CertificateExpiry = expiresAt
	if renewBefore > 0 {
		c.RenewBefore = renewBefore
	}
	c.RegistrationToken = ""
	c.AgentToken = ""
	return c.Save()
}

func (c *Config) Save() error {
	if err := c.Normalize(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.ConfigPath), 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(c.StateDir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := c.ConfigPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, c.ConfigPath); err != nil {
		return err
	}
	return os.Chmod(c.ConfigPath, 0600)
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
