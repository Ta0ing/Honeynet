package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultAnalyticsConfigName = "analytics.yaml"

var analyticsIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// AnalyticsConfig is intentionally loaded from a separate file. Older Server
// binaries use strict server.yaml parsing; keeping the analytical engine
// configuration separate makes a binary rollback safe.
type AnalyticsConfig struct {
	Enabled         bool
	DSN             string
	Database        string
	Table           string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	Path            string
}

type analyticsFile struct {
	Analytics struct {
		Enabled         bool   `yaml:"enabled"`
		DSN             string `yaml:"dsn"`
		Database        string `yaml:"database"`
		Table           string `yaml:"table"`
		MaxOpenConns    int    `yaml:"max_open_conns"`
		MaxIdleConns    int    `yaml:"max_idle_conns"`
		ConnMaxLifetime string `yaml:"conn_max_lifetime"`
		DialTimeout     string `yaml:"dial_timeout"`
		ReadTimeout     string `yaml:"read_timeout"`
	} `yaml:"analytics"`
}

func DefaultAnalyticsConfig() AnalyticsConfig {
	return AnalyticsConfig{
		Database: "honeynet_analytics", Table: "security_events",
		MaxOpenConns: 10, MaxIdleConns: 5, ConnMaxLifetime: time.Hour,
		DialTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
	}
}

// LoadAnalytics loads analytics.yaml next to server.yaml. A missing derived
// file means analytics is disabled for developer/backward-compatible runs; an
// explicitly configured missing file is an error.
func LoadAnalytics(serverConfigPath string) (AnalyticsConfig, error) {
	cfg := DefaultAnalyticsConfig()
	explicitPath := strings.TrimSpace(os.Getenv("HONEYPOT_ANALYTICS_CONFIG"))
	path := explicitPath
	if path == "" && strings.TrimSpace(serverConfigPath) != "" {
		path = filepath.Join(filepath.Dir(serverConfigPath), defaultAnalyticsConfigName)
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && explicitPath == "" {
				return applyAnalyticsEnvironment(cfg)
			}
			return AnalyticsConfig{}, fmt.Errorf("read analytics config %q: %w", path, err)
		}
		var raw analyticsFile
		decoder := yaml.NewDecoder(strings.NewReader(string(data)))
		decoder.KnownFields(true)
		if err := decoder.Decode(&raw); err != nil {
			return AnalyticsConfig{}, fmt.Errorf("parse analytics config %q: %w", path, err)
		}
		cfg.Enabled = raw.Analytics.Enabled
		assign(&cfg.DSN, raw.Analytics.DSN)
		assign(&cfg.Database, raw.Analytics.Database)
		assign(&cfg.Table, raw.Analytics.Table)
		if raw.Analytics.MaxOpenConns > 0 {
			cfg.MaxOpenConns = raw.Analytics.MaxOpenConns
		}
		if raw.Analytics.MaxIdleConns >= 0 {
			cfg.MaxIdleConns = raw.Analytics.MaxIdleConns
		}
		if err := parseAnalyticsDuration(&cfg.ConnMaxLifetime, raw.Analytics.ConnMaxLifetime, "conn_max_lifetime"); err != nil {
			return AnalyticsConfig{}, err
		}
		if err := parseAnalyticsDuration(&cfg.DialTimeout, raw.Analytics.DialTimeout, "dial_timeout"); err != nil {
			return AnalyticsConfig{}, err
		}
		if err := parseAnalyticsDuration(&cfg.ReadTimeout, raw.Analytics.ReadTimeout, "read_timeout"); err != nil {
			return AnalyticsConfig{}, err
		}
		cfg.Path = path
	}
	return applyAnalyticsEnvironment(cfg)
}

func applyAnalyticsEnvironment(cfg AnalyticsConfig) (AnalyticsConfig, error) {
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_ANALYTICS_ENABLED")); value != "" {
		cfg.Enabled = strings.EqualFold(value, "true") || value == "1"
	}
	if value := strings.TrimSpace(os.Getenv("HONEYPOT_CLICKHOUSE_DSN")); value != "" {
		cfg.DSN = value
		cfg.Enabled = true
	}
	if cfg.Enabled && strings.TrimSpace(cfg.DSN) == "" {
		return AnalyticsConfig{}, errors.New("analytics.dsn or HONEYPOT_CLICKHOUSE_DSN is required when ClickHouse analytics is enabled")
	}
	if !analyticsIdentifier.MatchString(cfg.Database) || !analyticsIdentifier.MatchString(cfg.Table) {
		return AnalyticsConfig{}, errors.New("analytics.database and analytics.table must be safe ClickHouse identifiers")
	}
	if cfg.MaxOpenConns < 1 || cfg.MaxIdleConns < 0 || cfg.MaxIdleConns > cfg.MaxOpenConns {
		return AnalyticsConfig{}, errors.New("analytics connection pool limits are invalid")
	}
	return cfg, nil
}

func parseAnalyticsDuration(target *time.Duration, value, name string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fmt.Errorf("analytics.%s must be a positive Go duration", name)
	}
	*target = duration
	return nil
}
