package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version                    string
	Addr                       string
	DatabaseDSN                string
	JWTSecret                  string
	JWTExpires                 time.Duration
	AdminUsername              string
	AdminPassword              string
	PublicURL                  string
	TLSEnabled                 bool
	TLSCertFile                string
	TLSKeyFile                 string
	WebDist                    string
	DownloadsDir               string
	BuiltinAgentToken          string
	AgentAddr                  string
	AgentPublicURL             string
	PKIDir                     string
	AgentTLSNames              []string
	AgentCertValidity          time.Duration
	AgentRenewBefore           time.Duration
	IPIPDBPath                 string
	IPIPLanguage               string
	ThreatIntelEnabled         bool
	ThreatIntelDBPath          string
	ThreatIntelDownloadURL     string
	ThreatIntelArchivePassword string
	ThreatIntelUpdateInterval  time.Duration
	DetectionRulesDir          string
	// RedactSensitiveEvents controls the default disclosure policy for attack
	// event API and WebSocket responses. It is intentionally opt-in: persisted
	// evidence is always kept verbatim, while operators that require masked
	// console views can enable this policy explicitly.
	RedactSensitiveEvents bool
	AIEnabled             bool
	AIProvider            string
	AIBaseURL             string
	AIAPIKey              string
	AIModel               string
	AITimeout             time.Duration
	AISendRawPacket       bool
	CORSOrigins           []string
}

type fileConfig struct {
	Server struct {
		Addr         string `yaml:"addr"`
		PublicURL    string `yaml:"public_url"`
		TLSEnabled   bool   `yaml:"tls_enabled"`
		TLSCertFile  string `yaml:"tls_cert_file"`
		TLSKeyFile   string `yaml:"tls_key_file"`
		WebDist      string `yaml:"web_dist"`
		DownloadsDir string `yaml:"downloads_dir"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	Auth struct {
		JWTSecret     string `yaml:"jwt_secret"`
		JWTExpires    string `yaml:"jwt_expires"`
		AdminUsername string `yaml:"admin_username"`
		AdminPassword string `yaml:"admin_password"`
	} `yaml:"auth"`
	BuiltinAgent struct {
		Token string `yaml:"token"`
	} `yaml:"builtin_agent"`
	Agent struct {
		Addr                string   `yaml:"addr"`
		PublicURL           string   `yaml:"public_url"`
		PKIDir              string   `yaml:"pki_dir"`
		TLSNames            []string `yaml:"tls_names"`
		CertificateValidity string   `yaml:"certificate_validity"`
		RenewBefore         string   `yaml:"renew_before"`
	} `yaml:"agent"`
	GeoIP struct {
		IPIPDBPath string `yaml:"ipip_db_path"`
		Language   string `yaml:"language"`
	} `yaml:"geoip"`
	ThreatIntelligence struct {
		Enabled        bool   `yaml:"enabled"`
		DatabasePath   string `yaml:"database_path"`
		DownloadURL    string `yaml:"download_url"`
		UpdateInterval string `yaml:"update_interval"`
	} `yaml:"threat_intelligence"`
	Detection struct {
		RulesDir string `yaml:"rules_dir"`
	} `yaml:"detection"`
	Security struct {
		RedactSensitiveEvents bool `yaml:"redact_sensitive_events"`
	} `yaml:"security"`
	AI struct {
		Enabled       bool   `yaml:"enabled"`
		Provider      string `yaml:"provider"`
		BaseURL       string `yaml:"base_url"`
		APIKey        string `yaml:"api_key"`
		Model         string `yaml:"model"`
		Timeout       string `yaml:"timeout"`
		SendRawPacket bool   `yaml:"send_raw_packet"`
	} `yaml:"ai"`
	CORS struct {
		Origins []string `yaml:"origins"`
	} `yaml:"cors"`
}

// Load reads an optional YAML configuration file and then applies environment
// variables as the final override. An empty path intentionally selects the
// environment-only development mode.
func Load(path string) (Config, error) {
	cfg := Config{
		Addr:              env("HONEYPOT_ADDR", ":8080"),
		DatabaseDSN:       env("HONEYPOT_DATABASE_DSN", "honeynet:honeynet@tcp(127.0.0.1:3306)/honeynet?charset=utf8mb4&parseTime=True&loc=Local"),
		JWTSecret:         env("HONEYPOT_JWT_SECRET", "dev-only-change-this-secret"),
		JWTExpires:        8 * time.Hour,
		AdminUsername:     env("HONEYPOT_ADMIN_USERNAME", "admin"),
		AdminPassword:     env("HONEYPOT_ADMIN_PASSWORD", "Honeynet@123"),
		PublicURL:         strings.TrimRight(env("HONEYPOT_PUBLIC_URL", "http://localhost:8080"), "/"),
		TLSEnabled:        envBool("HONEYPOT_TLS_ENABLED", false),
		TLSCertFile:       strings.TrimSpace(os.Getenv("HONEYPOT_TLS_CERT_FILE")),
		TLSKeyFile:        strings.TrimSpace(os.Getenv("HONEYPOT_TLS_KEY_FILE")),
		WebDist:           env("HONEYPOT_WEB_DIST", "web/dist"),
		DownloadsDir:      env("HONEYPOT_DOWNLOADS_DIR", "downloads"),
		BuiltinAgentToken: env("HONEYPOT_BUILTIN_AGENT_TOKEN", ""),
		AgentAddr:         env("HONEYPOT_AGENT_ADDR", ":8443"),
		AgentPublicURL:    strings.TrimRight(strings.TrimSpace(os.Getenv("HONEYPOT_AGENT_PUBLIC_URL")), "/"),
		PKIDir:            env("HONEYPOT_PKI_DIR", filepath.Join("data", "pki")),
		// Node mTLS certificates default to 400 days. This is deliberately
		// longer than one year while still rotating automatically well before
		// expiry.
		AgentCertValidity:          400 * 24 * time.Hour,
		AgentRenewBefore:           30 * 24 * time.Hour,
		IPIPDBPath:                 strings.TrimSpace(os.Getenv("HONEYPOT_IPIP_DB_PATH")),
		IPIPLanguage:               env("HONEYPOT_IPIP_LANGUAGE", "CN"),
		ThreatIntelEnabled:         envBool("HONEYPOT_THREAT_INTEL_ENABLED", false),
		ThreatIntelDBPath:          env("HONEYPOT_THREAT_INTEL_DB_PATH", filepath.Join("data", "threat-intelligence.db")),
		ThreatIntelDownloadURL:     strings.TrimSpace(os.Getenv("HONEYPOT_THREAT_INTEL_DOWNLOAD_URL")),
		ThreatIntelArchivePassword: os.Getenv("HONEYPOT_THREAT_INTEL_ARCHIVE_PASSWORD"),
		ThreatIntelUpdateInterval:  24 * time.Hour,
		DetectionRulesDir:          env("HONEYPOT_DETECTION_RULES_DIR", defaultDetectionRulesDir()),
		RedactSensitiveEvents:      envBool("HONEYPOT_REDACT_SENSITIVE_EVENTS", false),
		AIEnabled:                  envBool("HONEYPOT_AI_ENABLED", false),
		AIProvider:                 env("HONEYPOT_AI_PROVIDER", "openai-compatible"),
		AIBaseURL:                  strings.TrimRight(strings.TrimSpace(os.Getenv("HONEYPOT_AI_BASE_URL")), "/"),
		AIAPIKey:                   strings.TrimSpace(os.Getenv("HONEYPOT_AI_API_KEY")),
		AIModel:                    strings.TrimSpace(os.Getenv("HONEYPOT_AI_MODEL")),
		AITimeout:                  45 * time.Second,
		AISendRawPacket:            envBool("HONEYPOT_AI_SEND_RAW_PACKET", false),
		CORSOrigins:                split(env("HONEYPOT_CORS_ORIGINS", "http://localhost:5173")),
	}

	if strings.TrimSpace(path) != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read server config %q: %w", path, err)
		}
		var raw fileConfig
		decoder := yaml.NewDecoder(strings.NewReader(string(data)))
		decoder.KnownFields(true)
		if err := decoder.Decode(&raw); err != nil {
			return Config{}, fmt.Errorf("parse server config %q: %w", path, err)
		}
		applyFile(&cfg, raw)
	}

	applyEnvironment(&cfg)
	cfg.PublicURL = strings.TrimRight(cfg.PublicURL, "/")
	cfg.AIBaseURL = strings.TrimRight(strings.TrimSpace(cfg.AIBaseURL), "/")
	if cfg.AgentPublicURL == "" {
		cfg.AgentPublicURL = deriveAgentURL(cfg.PublicURL, cfg.AgentAddr)
	}
	cfg.AgentPublicURL = strings.TrimRight(cfg.AgentPublicURL, "/")
	if len(cfg.AgentTLSNames) == 0 {
		cfg.AgentTLSNames = defaultTLSNames(cfg.AgentPublicURL)
	}
	// Local installer probes use loopback even when an existing configuration
	// supplied an explicit public-only SAN list. Keeping both loopbacks present
	// also makes dual-stack health checks deterministic after an upgrade.
	cfg.AgentTLSNames = appendTLSNameValue(cfg.AgentTLSNames, "localhost")
	cfg.AgentTLSNames = appendTLSNameValue(cfg.AgentTLSNames, "127.0.0.1")
	cfg.AgentTLSNames = appendTLSNameValue(cfg.AgentTLSNames, "::1")
	if cfg.TLSEnabled {
		cfg.AgentTLSNames = appendTLSName(cfg.AgentTLSNames, cfg.PublicURL)
	}
	if cfg.Addr == "" || cfg.DatabaseDSN == "" || cfg.JWTSecret == "" || cfg.PublicURL == "" || cfg.AgentAddr == "" || cfg.AgentPublicURL == "" || cfg.PKIDir == "" {
		return Config{}, errors.New("server, database, auth and agent listener settings are required")
	}
	if err := validateListenAddress(cfg.Addr); err != nil {
		return Config{}, fmt.Errorf("server.addr: %w", err)
	}
	if err := validateListenAddress(cfg.AgentAddr); err != nil {
		return Config{}, fmt.Errorf("agent.addr: %w", err)
	}
	if err := validateHTTPURL(cfg.PublicURL, false); err != nil {
		return Config{}, fmt.Errorf("server.public_url: %w", err)
	}
	if cfg.TLSEnabled {
		publicURL, _ := url.Parse(cfg.PublicURL)
		if publicURL.Scheme != "https" {
			return Config{}, errors.New("server.public_url must use https when server.tls_enabled is true")
		}
	}
	if (cfg.TLSCertFile == "") != (cfg.TLSKeyFile == "") {
		return Config{}, errors.New("server.tls_cert_file and server.tls_key_file must be configured together")
	}
	if cfg.UsesExternalConsoleCertificate() {
		if !cfg.TLSEnabled {
			return Config{}, errors.New("server.tls_enabled must be true when an external console certificate is configured")
		}
		if _, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
			return Config{}, fmt.Errorf("load external console TLS certificate: %w", err)
		}
	}
	if cfg.JWTExpires <= 0 {
		return Config{}, errors.New("auth.jwt_expires must be a positive Go duration such as 8h")
	}
	if len(cfg.CORSOrigins) == 0 {
		return Config{}, errors.New("at least one cors origin is required")
	}
	if cfg.AgentCertValidity < 24*time.Hour {
		return Config{}, errors.New("agent.certificate_validity must be at least 24h")
	}
	if cfg.AgentRenewBefore <= 0 || cfg.AgentRenewBefore >= cfg.AgentCertValidity {
		return Config{}, errors.New("agent.renew_before must be positive and shorter than certificate_validity")
	}
	if cfg.ThreatIntelEnabled && strings.TrimSpace(cfg.ThreatIntelDBPath) == "" {
		return Config{}, errors.New("threat_intelligence.database_path is required when threat intelligence is enabled")
	}
	if cfg.ThreatIntelUpdateInterval < time.Hour || cfg.ThreatIntelUpdateInterval > 30*24*time.Hour {
		return Config{}, errors.New("threat_intelligence.update_interval must be between 1h and 720h")
	}
	if cfg.ThreatIntelDownloadURL != "" {
		if err := validateHTTPURL(cfg.ThreatIntelDownloadURL, true); err != nil {
			return Config{}, fmt.Errorf("threat_intelligence.download_url: %w", err)
		}
	}
	if cfg.AITimeout <= 0 || cfg.AITimeout > 5*time.Minute {
		return Config{}, errors.New("ai.timeout must be positive and no longer than 5m")
	}
	if cfg.AIEnabled && (cfg.AIBaseURL == "" || cfg.AIAPIKey == "" || cfg.AIModel == "") {
		return Config{}, errors.New("ai.base_url, ai.api_key and ai.model are required when AI is enabled")
	}
	if err := validateHTTPURL(cfg.AgentPublicURL, true); err != nil {
		return Config{}, fmt.Errorf("agent.public_url: %w", err)
	}
	return cfg, nil
}

func applyFile(cfg *Config, raw fileConfig) {
	assign(&cfg.Addr, raw.Server.Addr)
	assign(&cfg.PublicURL, raw.Server.PublicURL)
	if raw.Server.TLSEnabled {
		cfg.TLSEnabled = true
	}
	assign(&cfg.TLSCertFile, raw.Server.TLSCertFile)
	assign(&cfg.TLSKeyFile, raw.Server.TLSKeyFile)
	assign(&cfg.WebDist, raw.Server.WebDist)
	assign(&cfg.DownloadsDir, raw.Server.DownloadsDir)
	assign(&cfg.DatabaseDSN, raw.Database.DSN)
	assign(&cfg.JWTSecret, raw.Auth.JWTSecret)
	assign(&cfg.AdminUsername, raw.Auth.AdminUsername)
	assign(&cfg.AdminPassword, raw.Auth.AdminPassword)
	assign(&cfg.BuiltinAgentToken, raw.BuiltinAgent.Token)
	assign(&cfg.AgentAddr, raw.Agent.Addr)
	assign(&cfg.AgentPublicURL, raw.Agent.PublicURL)
	assign(&cfg.PKIDir, raw.Agent.PKIDir)
	if len(raw.Agent.TLSNames) > 0 {
		cfg.AgentTLSNames = raw.Agent.TLSNames
	}
	applyDuration(&cfg.AgentCertValidity, raw.Agent.CertificateValidity)
	applyDuration(&cfg.AgentRenewBefore, raw.Agent.RenewBefore)
	assign(&cfg.IPIPDBPath, raw.GeoIP.IPIPDBPath)
	assign(&cfg.IPIPLanguage, raw.GeoIP.Language)
	if raw.ThreatIntelligence.Enabled {
		cfg.ThreatIntelEnabled = true
	}
	assign(&cfg.ThreatIntelDBPath, raw.ThreatIntelligence.DatabasePath)
	assign(&cfg.ThreatIntelDownloadURL, raw.ThreatIntelligence.DownloadURL)
	applyDuration(&cfg.ThreatIntelUpdateInterval, raw.ThreatIntelligence.UpdateInterval)
	assign(&cfg.DetectionRulesDir, raw.Detection.RulesDir)
	if raw.Security.RedactSensitiveEvents {
		cfg.RedactSensitiveEvents = true
	}
	if raw.AI.Enabled {
		cfg.AIEnabled = true
	}
	assign(&cfg.AIProvider, raw.AI.Provider)
	assign(&cfg.AIBaseURL, raw.AI.BaseURL)
	assign(&cfg.AIAPIKey, raw.AI.APIKey)
	assign(&cfg.AIModel, raw.AI.Model)
	applyDuration(&cfg.AITimeout, raw.AI.Timeout)
	if raw.AI.SendRawPacket {
		cfg.AISendRawPacket = true
	}
	if raw.Auth.JWTExpires != "" {
		duration, err := time.ParseDuration(raw.Auth.JWTExpires)
		if err == nil {
			cfg.JWTExpires = duration
		} else {
			cfg.JWTExpires = 0
		}
	}
	if len(raw.CORS.Origins) > 0 {
		cfg.CORSOrigins = raw.CORS.Origins
	}
}

func applyEnvironment(cfg *Config) {
	assignEnv(&cfg.Addr, "HONEYPOT_ADDR")
	assignEnv(&cfg.DatabaseDSN, "HONEYPOT_DATABASE_DSN")
	assignEnv(&cfg.JWTSecret, "HONEYPOT_JWT_SECRET")
	assignEnv(&cfg.AdminUsername, "HONEYPOT_ADMIN_USERNAME")
	assignEnv(&cfg.AdminPassword, "HONEYPOT_ADMIN_PASSWORD")
	assignEnv(&cfg.PublicURL, "HONEYPOT_PUBLIC_URL")
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_TLS_ENABLED")); value != "" {
		cfg.TLSEnabled = envBool("HONEYPOT_TLS_ENABLED", cfg.TLSEnabled)
	}
	assignEnv(&cfg.TLSCertFile, "HONEYPOT_TLS_CERT_FILE")
	assignEnv(&cfg.TLSKeyFile, "HONEYPOT_TLS_KEY_FILE")
	assignEnv(&cfg.WebDist, "HONEYPOT_WEB_DIST")
	assignEnv(&cfg.DownloadsDir, "HONEYPOT_DOWNLOADS_DIR")
	assignEnv(&cfg.BuiltinAgentToken, "HONEYPOT_BUILTIN_AGENT_TOKEN")
	assignEnv(&cfg.AgentAddr, "HONEYPOT_AGENT_ADDR")
	assignEnv(&cfg.AgentPublicURL, "HONEYPOT_AGENT_PUBLIC_URL")
	assignEnv(&cfg.PKIDir, "HONEYPOT_PKI_DIR")
	assignEnv(&cfg.IPIPDBPath, "HONEYPOT_IPIP_DB_PATH")
	assignEnv(&cfg.IPIPLanguage, "HONEYPOT_IPIP_LANGUAGE")
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_THREAT_INTEL_ENABLED")); value != "" {
		cfg.ThreatIntelEnabled = envBool("HONEYPOT_THREAT_INTEL_ENABLED", cfg.ThreatIntelEnabled)
	}
	assignEnv(&cfg.ThreatIntelDBPath, "HONEYPOT_THREAT_INTEL_DB_PATH")
	assignEnv(&cfg.ThreatIntelDownloadURL, "HONEYPOT_THREAT_INTEL_DOWNLOAD_URL")
	if value, exists := os.LookupEnv("HONEYPOT_THREAT_INTEL_ARCHIVE_PASSWORD"); exists {
		cfg.ThreatIntelArchivePassword = value
	}
	applyEnvDuration(&cfg.ThreatIntelUpdateInterval, "HONEYPOT_THREAT_INTEL_UPDATE_INTERVAL")
	assignEnv(&cfg.DetectionRulesDir, "HONEYPOT_DETECTION_RULES_DIR")
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_REDACT_SENSITIVE_EVENTS")); value != "" {
		cfg.RedactSensitiveEvents = envBool("HONEYPOT_REDACT_SENSITIVE_EVENTS", cfg.RedactSensitiveEvents)
	}
	assignEnv(&cfg.AIProvider, "HONEYPOT_AI_PROVIDER")
	assignEnv(&cfg.AIBaseURL, "HONEYPOT_AI_BASE_URL")
	assignEnv(&cfg.AIAPIKey, "HONEYPOT_AI_API_KEY")
	assignEnv(&cfg.AIModel, "HONEYPOT_AI_MODEL")
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_AI_ENABLED")); value != "" {
		cfg.AIEnabled = envBool("HONEYPOT_AI_ENABLED", cfg.AIEnabled)
	}
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_AI_SEND_RAW_PACKET")); value != "" {
		cfg.AISendRawPacket = envBool("HONEYPOT_AI_SEND_RAW_PACKET", cfg.AISendRawPacket)
	}
	applyEnvDuration(&cfg.AITimeout, "HONEYPOT_AI_TIMEOUT")
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_AGENT_TLS_NAMES")); value != "" {
		cfg.AgentTLSNames = split(value)
	}
	applyEnvDuration(&cfg.AgentCertValidity, "HONEYPOT_AGENT_CERT_VALIDITY")
	applyEnvDuration(&cfg.AgentRenewBefore, "HONEYPOT_AGENT_RENEW_BEFORE")
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_JWT_EXPIRES")); value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil {
			cfg.JWTExpires = duration
		} else {
			cfg.JWTExpires = 0
		}
	}
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_CORS_ORIGINS")); value != "" {
		cfg.CORSOrigins = split(value)
	}
}

// UsesExternalConsoleCertificate reports whether the management console uses
// an operator-provided certificate instead of the private Agent PKI. The Agent
// gateway always keeps using its independent mTLS certificate and CA.
func (cfg Config) UsesExternalConsoleCertificate() bool {
	return cfg.TLSCertFile != "" && cfg.TLSKeyFile != ""
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func defaultDetectionRulesDir() string {
	for _, candidate := range []string{filepath.Join("rules", "builtin"), filepath.Join("cve-rules-decrypted", "Yara")} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return filepath.Join("rules", "builtin")
}

func applyDuration(target *time.Duration, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		*target = 0
		return
	}
	*target = duration
}

func applyEnvDuration(target *time.Duration, key string) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		applyDuration(target, value)
	}
}

func deriveAgentURL(publicURL, listen string) string {
	parsed, err := url.Parse(publicURL)
	if err != nil || parsed.Hostname() == "" {
		return ""
	}
	port := strings.TrimPrefix(strings.TrimSpace(listen), ":")
	if host, _, err := net.SplitHostPort(listen); err == nil {
		_ = host
		_, port, _ = net.SplitHostPort(listen)
	}
	if port == "" {
		port = "8443"
	}
	parsed.Scheme = "https"
	parsed.Host = net.JoinHostPort(parsed.Hostname(), port)
	parsed.Path, parsed.RawPath, parsed.RawQuery, parsed.Fragment = "", "", "", ""
	return parsed.String()
}

func validateListenAddress(value string) error {
	host, portText, err := net.SplitHostPort(strings.TrimSpace(value))
	if err != nil {
		return errors.New("must use host:port syntax; bracket IPv6 addresses, for example [::]:8080")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if strings.Contains(host, "[") || strings.Contains(host, "]") {
		return errors.New("host contains invalid brackets")
	}
	return nil
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

func defaultTLSNames(agentURL string) []string {
	items := []string{"localhost", "127.0.0.1", "::1"}
	if parsed, err := url.Parse(agentURL); err == nil && parsed.Hostname() != "" {
		items = append([]string{parsed.Hostname()}, items...)
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" && !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func appendTLSName(names []string, rawURL string) []string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Hostname() == "" {
		return names
	}
	return appendTLSNameValue(names, parsed.Hostname())
}

func appendTLSNameValue(names []string, host string) []string {
	host = strings.TrimSpace(host)
	if host == "" {
		return names
	}
	for _, name := range names {
		if strings.EqualFold(strings.TrimSpace(name), host) {
			return names
		}
	}
	return append(names, host)
}

func assign(target *string, value string) {
	if value = strings.TrimSpace(value); value != "" {
		*target = value
	}
}

func assignEnv(target *string, key string) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		*target = value
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func split(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
