package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const (
	minimumSchemaVersion = 2
	requiredTableEngine  = "ReplacingMergeTree"
)

// requiredEventColumns is deliberately exhaustive. An application binary
// must not start against a partially applied or manually altered event table:
// accepting events in that state could acknowledge evidence that cannot be
// queried back with its original forensic fields.
var requiredEventColumns = map[string]string{
	"event_id":                 "String",
	"node_id":                  "String",
	"pot_id":                   "String",
	"decoy_id":                 "String",
	"service":                  "LowCardinality(String)",
	"event_type":               "LowCardinality(String)",
	"event_time":               "DateTime64(3, 'UTC')",
	"ingested_at":              "DateTime64(3, 'UTC')",
	"src_ip":                   "String",
	"src_port":                 "UInt16",
	"dst_ip":                   "String",
	"dst_port":                 "UInt16",
	"geo":                      "LowCardinality(String)",
	"asn":                      "LowCardinality(String)",
	"raw_packet":               "String",
	"payload":                  "String",
	"tags":                     "String",
	"detections":               "String",
	"agent_rule_revision":      "Int64",
	"server_rule_revision":     "Int64",
	"session_id":               "String",
	"has_credential":           "UInt8",
	"credential_username":      "String",
	"credential_password":      "String",
	"credential_auth_response": "String",
	"credential_mechanism":     "LowCardinality(String)",
	"record_version":           "UInt64",
}

// ValidateSchema verifies the runtime schema used by all event reads and
// writes. It also records the migration version observed from ClickHouse, so
// status responses never report a compile-time constant as database state.
func (r *Repository) ValidateSchema(ctx context.Context) error {
	migrationsTable, err := qualifiedTable(r.database, "schema_migrations")
	if err != nil {
		return fmt.Errorf("resolve ClickHouse migration table: %w", err)
	}

	var version uint32
	versionQuery := "SELECT toUInt32(ifNull(max(version), 0)) FROM " + migrationsTable + " FINAL"
	if err := r.conn.QueryRow(ctx, versionQuery).Scan(&version); err != nil {
		return fmt.Errorf("read ClickHouse schema migration version: %w", err)
	}
	r.setSchemaVersion(int(version))
	if version < minimumSchemaVersion {
		return fmt.Errorf("ClickHouse schema version %d is below required version %d", version, minimumSchemaVersion)
	}

	var engine string
	if err := r.conn.QueryRow(ctx,
		"SELECT engine FROM system.tables WHERE database = ? AND name = ? LIMIT 1",
		r.database, r.table,
	).Scan(&engine); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("ClickHouse event table %s.%s is missing", r.database, r.table)
		}
		return fmt.Errorf("read ClickHouse event table engine: %w", err)
	}
	if engine != requiredTableEngine {
		return fmt.Errorf("ClickHouse event table engine is %q, expected %q", engine, requiredTableEngine)
	}

	rows, err := r.conn.Query(ctx,
		"SELECT name, type FROM system.columns WHERE database = ? AND table = ? ORDER BY position",
		r.database, r.table,
	)
	if err != nil {
		return fmt.Errorf("read ClickHouse event table columns: %w", err)
	}
	defer rows.Close()

	found := make(map[string]string, len(requiredEventColumns))
	for rows.Next() {
		var name, columnType string
		if err := rows.Scan(&name, &columnType); err != nil {
			return fmt.Errorf("scan ClickHouse event table column: %w", err)
		}
		found[name] = columnType
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate ClickHouse event table columns: %w", err)
	}
	for name, expectedType := range requiredEventColumns {
		actualType, exists := found[name]
		if !exists {
			return fmt.Errorf("ClickHouse event table required column %q is missing", name)
		}
		if actualType != expectedType {
			return fmt.Errorf("ClickHouse event table column %q has type %q, expected %q", name, actualType, expectedType)
		}
	}
	return nil
}

func (r *Repository) setSchemaVersion(version int) {
	r.mu.Lock()
	r.schemaVersion = version
	r.mu.Unlock()
}

func (r *Repository) currentSchemaVersion() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.schemaVersion
}
