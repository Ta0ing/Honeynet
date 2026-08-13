#!/bin/bash
set -euo pipefail

database=${CLICKHOUSE_DB:-honeynet_analytics}
migration_root=${HONEYPOT_CLICKHOUSE_MIGRATIONS_DIR:-/opt/honeynet-clickhouse/migrations}

case "$database" in
  ''|*[!A-Za-z0-9_]*) echo "Invalid ClickHouse database identifier: $database" >&2; exit 2 ;;
esac

for variable in \
  HONEYPOT_CLICKHOUSE_APP_PASSWORD_SHA256 \
  HONEYPOT_CLICKHOUSE_INGEST_PASSWORD_SHA256 \
  HONEYPOT_CLICKHOUSE_QUERY_PASSWORD_SHA256; do
  value=${!variable:-}
  case "$value" in
    ''|*[!0-9a-f]*) echo "$variable must be a lowercase SHA-256 digest" >&2; exit 2 ;;
  esac
  if [ "${#value}" -ne 64 ]; then
    echo "$variable must contain exactly 64 hexadecimal characters" >&2
    exit 2
  fi
done

client=(clickhouse-client --host 127.0.0.1 --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$database")

"${client[@]}" --multiquery <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations
(
    version UInt32,
    name String,
    applied_at DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = ReplacingMergeTree(applied_at)
ORDER BY version;
SQL

shopt -s nullglob
migrations=("$migration_root"/[0-9][0-9][0-9]_*.sql)
if [ "${#migrations[@]}" -eq 0 ]; then
  echo "No ClickHouse migration files found in $migration_root" >&2
  exit 1
fi

for migration in "${migrations[@]}"; do
  filename=$(basename "$migration")
  version=${filename%%_*}
  version=$((10#$version))
  applied=$("${client[@]}" --query "SELECT count() FROM schema_migrations FINAL WHERE version = $version")
  if [ "$applied" -gt 0 ]; then
    echo "ClickHouse migration already applied: $filename"
    continue
  fi
  echo "Applying ClickHouse migration: $filename"
  "${client[@]}" --multiquery < "$migration"
  "${client[@]}" --query "INSERT INTO schema_migrations (version, name) VALUES ($version, {name:String})" --param_name "$filename"
done

"${client[@]}" --multiquery <<SQL
CREATE ROLE IF NOT EXISTS honeynet_ingest_role;
CREATE ROLE IF NOT EXISTS honeynet_query_role;
CREATE ROLE IF NOT EXISTS honeynet_app_role;

REVOKE ALL ON *.* FROM honeynet_ingest_role;
REVOKE ALL ON *.* FROM honeynet_query_role;
REVOKE ALL ON *.* FROM honeynet_app_role;

GRANT INSERT ON ${database}.security_events TO honeynet_ingest_role;
GRANT SELECT ON ${database}.security_events TO honeynet_query_role;
GRANT SELECT ON ${database}.security_events_daily_mv TO honeynet_query_role;
GRANT INSERT ON ${database}.security_events TO honeynet_app_role;
GRANT SELECT ON ${database}.security_events TO honeynet_app_role;
GRANT SELECT ON ${database}.security_events_daily_mv TO honeynet_app_role;
GRANT SELECT ON ${database}.schema_migrations TO honeynet_app_role;

CREATE USER IF NOT EXISTS honeynet_ingest IDENTIFIED WITH sha256_hash BY '${HONEYPOT_CLICKHOUSE_INGEST_PASSWORD_SHA256}';
ALTER USER honeynet_ingest IDENTIFIED WITH sha256_hash BY '${HONEYPOT_CLICKHOUSE_INGEST_PASSWORD_SHA256}';
REVOKE ALL FROM honeynet_ingest;
GRANT honeynet_ingest_role TO honeynet_ingest;
SET DEFAULT ROLE honeynet_ingest_role TO honeynet_ingest;

CREATE USER IF NOT EXISTS honeynet_query IDENTIFIED WITH sha256_hash BY '${HONEYPOT_CLICKHOUSE_QUERY_PASSWORD_SHA256}';
ALTER USER honeynet_query IDENTIFIED WITH sha256_hash BY '${HONEYPOT_CLICKHOUSE_QUERY_PASSWORD_SHA256}';
REVOKE ALL FROM honeynet_query;
GRANT honeynet_query_role TO honeynet_query;
SET DEFAULT ROLE honeynet_query_role TO honeynet_query;

CREATE USER IF NOT EXISTS honeynet_app IDENTIFIED WITH sha256_hash BY '${HONEYPOT_CLICKHOUSE_APP_PASSWORD_SHA256}';
ALTER USER honeynet_app IDENTIFIED WITH sha256_hash BY '${HONEYPOT_CLICKHOUSE_APP_PASSWORD_SHA256}';
REVOKE ALL FROM honeynet_app;
GRANT honeynet_app_role TO honeynet_app;
SET DEFAULT ROLE honeynet_app_role TO honeynet_app;
SQL

echo "ClickHouse schema and least-privilege users are ready."
