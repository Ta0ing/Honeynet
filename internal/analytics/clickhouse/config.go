package clickhouse

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

const defaultTable = "security_events"

var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type Config struct {
	DSN             string
	Database        string
	Table           string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
}

func Open(config Config) (*Repository, error) {
	if strings.TrimSpace(config.DSN) == "" {
		return nil, errors.New("clickhouse DSN is required")
	}
	options, err := clickhouse.ParseDSN(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse clickhouse DSN: %w", err)
	}
	if database := strings.TrimSpace(config.Database); database != "" {
		options.Auth.Database = database
	}
	if config.MaxOpenConns > 0 {
		options.MaxOpenConns = config.MaxOpenConns
	}
	if config.MaxIdleConns > 0 {
		options.MaxIdleConns = config.MaxIdleConns
	}
	if config.ConnMaxLifetime > 0 {
		options.ConnMaxLifetime = config.ConnMaxLifetime
	}
	if config.DialTimeout > 0 {
		options.DialTimeout = config.DialTimeout
	}
	if config.ReadTimeout > 0 {
		options.ReadTimeout = config.ReadTimeout
	}
	options.Compression = &clickhouse.Compression{Method: clickhouse.CompressionLZ4}
	connection, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}
	repository, err := New(connection, config.Table)
	if err != nil {
		_ = connection.Close()
		return nil, err
	}
	if options.Auth.Database != "" {
		repository.database = options.Auth.Database
	}
	return repository, nil
}

// New accepts the official driver's Conn interface so production code uses
// clickhouse-go/v2 while unit tests can embed and override only used methods.
func New(connection driver.Conn, table string) (*Repository, error) {
	if connection == nil {
		return nil, errors.New("clickhouse connection is required")
	}
	table = strings.TrimSpace(table)
	if table == "" {
		table = defaultTable
	}
	if !identifierPattern.MatchString(table) {
		return nil, fmt.Errorf("invalid clickhouse table name %q", table)
	}
	return &Repository{conn: connection, table: table, database: "default", now: time.Now}, nil
}

func qualifiedTable(database, table string) (string, error) {
	database, table = strings.TrimSpace(database), strings.TrimSpace(table)
	if database == "" {
		database = "default"
	}
	if table == "" {
		table = defaultTable
	}
	if !identifierPattern.MatchString(database) || !identifierPattern.MatchString(table) {
		return "", errors.New("invalid ClickHouse database or table identifier")
	}
	return "`" + database + "`.`" + table + "`", nil
}
