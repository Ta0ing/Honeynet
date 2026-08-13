package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type schemaConn struct {
	driver.Conn
	version       uint32
	versionErr    error
	engine        string
	engineErr     error
	columns       map[string]string
	columnsErr    error
	columnScanErr error
	rowsErr       error
	pingErr       error
}

func validSchemaConn() *schemaConn {
	columns := make(map[string]string, len(requiredEventColumns))
	for name, columnType := range requiredEventColumns {
		columns[name] = columnType
	}
	return &schemaConn{version: minimumSchemaVersion, engine: requiredTableEngine, columns: columns}
}

func (c *schemaConn) Ping(context.Context) error { return c.pingErr }

func (c *schemaConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }

func (c *schemaConn) QueryRow(_ context.Context, query string, _ ...any) driver.Row {
	switch {
	case strings.Contains(query, "schema_migrations"):
		return schemaRow{value: c.version, err: c.versionErr}
	case strings.Contains(query, "system.tables"):
		return schemaRow{value: c.engine, err: c.engineErr}
	default:
		return schemaRow{err: errors.New("unexpected query")}
	}
}

func (c *schemaConn) Query(_ context.Context, query string, _ ...any) (driver.Rows, error) {
	if !strings.Contains(query, "system.columns") {
		return nil, errors.New("unexpected query")
	}
	if c.columnsErr != nil {
		return nil, c.columnsErr
	}
	values := make([][2]string, 0, len(c.columns))
	for name, columnType := range c.columns {
		values = append(values, [2]string{name, columnType})
	}
	return &schemaRows{values: values, scanErr: c.columnScanErr, rowsErr: c.rowsErr}, nil
}

type schemaRow struct {
	value any
	err   error
}

func (r schemaRow) Err() error           { return r.err }
func (r schemaRow) ScanStruct(any) error { return errors.New("not implemented") }
func (r schemaRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != 1 {
		return errors.New("unexpected scan destination count")
	}
	switch target := dest[0].(type) {
	case *uint32:
		value, ok := r.value.(uint32)
		if !ok {
			return errors.New("unexpected uint32 scan value")
		}
		*target = value
	case *string:
		value, ok := r.value.(string)
		if !ok {
			return errors.New("unexpected string scan value")
		}
		*target = value
	default:
		return errors.New("unexpected scan destination")
	}
	return nil
}

type schemaRows struct {
	values  [][2]string
	index   int
	scanErr error
	rowsErr error
}

func (r *schemaRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++
	return true
}

func (r *schemaRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	if r.index == 0 || r.index > len(r.values) || len(dest) != 2 {
		return errors.New("unexpected row scan")
	}
	name, nameOK := dest[0].(*string)
	columnType, typeOK := dest[1].(*string)
	if !nameOK || !typeOK {
		return errors.New("unexpected row destinations")
	}
	*name = r.values[r.index-1][0]
	*columnType = r.values[r.index-1][1]
	return nil
}

func (r *schemaRows) ScanStruct(any) error             { return errors.New("not implemented") }
func (r *schemaRows) ColumnTypes() []driver.ColumnType { return nil }
func (r *schemaRows) Totals(...any) error              { return nil }
func (r *schemaRows) Columns() []string                { return []string{"name", "type"} }
func (r *schemaRows) Close() error                     { return nil }
func (r *schemaRows) Err() error                       { return r.rowsErr }

func TestValidateSchemaSuccessUsesDatabaseVersion(t *testing.T) {
	if got := len(requiredEventColumns); got != 27 {
		t.Fatalf("required event column count = %d, want 27", got)
	}
	repository, err := New(validSchemaConn(), defaultTable)
	if err != nil {
		t.Fatal(err)
	}
	if repository.currentSchemaVersion() != 0 {
		t.Fatal("constructor must not hard-code a schema version")
	}
	if err := repository.ValidateSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := repository.currentSchemaVersion(); got != minimumSchemaVersion {
		t.Fatalf("schema version = %d, want %d", got, minimumSchemaVersion)
	}
}

func TestValidateSchemaRejectsOldMigrationVersion(t *testing.T) {
	connection := validSchemaConn()
	connection.version = minimumSchemaVersion - 1
	repository, _ := New(connection, defaultTable)
	if err := repository.ValidateSchema(context.Background()); err == nil || !strings.Contains(err.Error(), "below required version") {
		t.Fatalf("old-schema error = %v", err)
	}
	if got := repository.currentSchemaVersion(); got != minimumSchemaVersion-1 {
		t.Fatalf("observed schema version = %d", got)
	}
}

func TestValidateSchemaRejectsWrongEngine(t *testing.T) {
	connection := validSchemaConn()
	connection.engine = "MergeTree"
	repository, _ := New(connection, defaultTable)
	if err := repository.ValidateSchema(context.Background()); err == nil || !strings.Contains(err.Error(), `expected "ReplacingMergeTree"`) {
		t.Fatalf("wrong-engine error = %v", err)
	}
}

func TestValidateSchemaRejectsMissingTable(t *testing.T) {
	connection := validSchemaConn()
	connection.engineErr = sql.ErrNoRows
	repository, _ := New(connection, defaultTable)
	if err := repository.ValidateSchema(context.Background()); err == nil || !strings.Contains(err.Error(), "is missing") {
		t.Fatalf("missing-table error = %v", err)
	}
}

func TestValidateSchemaRejectsMissingColumn(t *testing.T) {
	connection := validSchemaConn()
	delete(connection.columns, "server_rule_revision")
	repository, _ := New(connection, defaultTable)
	if err := repository.ValidateSchema(context.Background()); err == nil || !strings.Contains(err.Error(), `column "server_rule_revision" is missing`) {
		t.Fatalf("missing-column error = %v", err)
	}
}

func TestValidateSchemaRejectsWrongCriticalColumnType(t *testing.T) {
	connection := validSchemaConn()
	connection.columns["record_version"] = "Int64"
	repository, _ := New(connection, defaultTable)
	if err := repository.ValidateSchema(context.Background()); err == nil || !strings.Contains(err.Error(), `expected "UInt64"`) {
		t.Fatalf("wrong-column-type error = %v", err)
	}
}

func TestValidateSchemaPropagatesMetadataErrors(t *testing.T) {
	want := errors.New("metadata unavailable")
	connection := validSchemaConn()
	connection.columnsErr = want
	repository, _ := New(connection, defaultTable)
	if err := repository.ValidateSchema(context.Background()); !errors.Is(err, want) {
		t.Fatalf("metadata error = %v, want wrapped %v", err, want)
	}
}

func TestStatusRequiresValidSchemaAndSanitizesError(t *testing.T) {
	connection := validSchemaConn()
	connection.columnsErr = errors.New("dial tcp clickhouse://secret@example.invalid: leaked DSN")
	repository, _ := New(connection, defaultTable)
	status := repository.Status(context.Background())
	if status.Healthy {
		t.Fatal("invalid schema was reported healthy")
	}
	if status.Error != "安全分析引擎数据结构校验失败" {
		t.Fatalf("status error was not sanitized: %q", status.Error)
	}
	if strings.Contains(status.Error, "secret") {
		t.Fatal("status leaked connection details")
	}
	if status.SchemaVersion != minimumSchemaVersion {
		t.Fatalf("status schema version = %d", status.SchemaVersion)
	}
}

var _ driver.Conn = (*schemaConn)(nil)
var _ driver.Row = schemaRow{}
var _ driver.Rows = (*schemaRows)(nil)
