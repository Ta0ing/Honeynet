#!/bin/sh
set -eu

CONFIG_DIR=/etc/honeynet
ENV_FILE=$CONFIG_DIR/clickhouse.env
ANALYTICS_CONFIG=$CONFIG_DIR/analytics.yaml
SKIP_SMOKE=0
PROJECT_NAME=honeynet-analytics

usage() {
  echo "Usage: $0 [--env-file FILE] [--analytics-config FILE] [--skip-smoke]"
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --env-file) ENV_FILE=$2; shift 2 ;;
    --analytics-config) ANALYTICS_CONFIG=$2; shift 2 ;;
    --skip-smoke) SKIP_SMOKE=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [ "$(id -u)" -ne 0 ]; then
  echo "This installer must run as root." >&2
  exit 1
fi
command -v docker >/dev/null 2>&1 || {
  echo "Docker is required for the default ClickHouse deployment. Install Docker Engine with the Compose plugin, or rerun the Server installer with --skip-clickhouse." >&2
  exit 1
}
docker compose version >/dev/null 2>&1 || {
  echo "Docker Compose v2 is required. Install the docker-compose-plugin package, or rerun the Server installer with --skip-clickhouse." >&2
  exit 1
}
docker info >/dev/null 2>&1 || {
  echo "Docker daemon is unavailable. Start Docker and retry, or rerun the Server installer with --skip-clickhouse." >&2
  exit 1
}

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PACKAGE_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
COMPOSE_FILE=$PACKAGE_ROOT/deploy/clickhouse/compose.yaml
MIGRATIONS_DIR=$PACKAGE_ROOT/migrations/clickhouse
for required in \
  "$COMPOSE_FILE" \
  "$PACKAGE_ROOT/deploy/clickhouse/init/bootstrap.sh" \
  "$MIGRATIONS_DIR/001_security_events.sql" \
  "$MIGRATIONS_DIR/002_event_rollups.sql"; do
  if [ ! -f "$required" ]; then
    echo "Incomplete ClickHouse deployment assets: missing $required" >&2
    exit 1
  fi
done

random_hex() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  else
    od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
  fi
}

sha256_text() {
  if command -v sha256sum >/dev/null 2>&1; then
    printf '%s' "$1" | sha256sum | awk '{print $1}'
  else
    printf '%s' "$1" | shasum -a 256 | awk '{print $1}'
  fi
}

ENV_DIR=$(dirname "$ENV_FILE")
ANALYTICS_DIR=$(dirname "$ANALYTICS_CONFIG")
if ! getent group honeynet >/dev/null 2>&1; then
  groupadd --system honeynet
fi
if [ ! -d "$CONFIG_DIR" ]; then
  install -d -m 0750 -o root -g honeynet "$CONFIG_DIR"
fi
if [ "$ENV_DIR" != "$CONFIG_DIR" ] && [ "$ENV_DIR" != "$ANALYTICS_DIR" ] && [ ! -d "$ENV_DIR" ]; then
  install -d -m 0700 -o root -g root "$ENV_DIR"
fi
if [ "$ANALYTICS_DIR" != "$CONFIG_DIR" ] && [ ! -d "$ANALYTICS_DIR" ]; then
  install -d -m 0750 -o root -g honeynet "$ANALYTICS_DIR"
fi

if [ ! -f "$ENV_FILE" ]; then
  migrate_password=$(random_hex)
  app_password=$(random_hex)
  ingest_password=$(random_hex)
  query_password=$(random_hex)
  umask 077
  {
    echo "HONEYPOT_CLICKHOUSE_DATABASE=honeynet_analytics"
    echo "HONEYPOT_CLICKHOUSE_HTTP_PORT=8123"
    echo "HONEYPOT_CLICKHOUSE_NATIVE_PORT=9000"
    printf 'HONEYPOT_CLICKHOUSE_MIGRATE_PASSWORD=%s\n' "$migrate_password"
    printf 'HONEYPOT_CLICKHOUSE_APP_PASSWORD=%s\n' "$app_password"
    printf 'HONEYPOT_CLICKHOUSE_APP_PASSWORD_SHA256=%s\n' "$(sha256_text "$app_password")"
    printf 'HONEYPOT_CLICKHOUSE_INGEST_PASSWORD=%s\n' "$ingest_password"
    printf 'HONEYPOT_CLICKHOUSE_INGEST_PASSWORD_SHA256=%s\n' "$(sha256_text "$ingest_password")"
    printf 'HONEYPOT_CLICKHOUSE_QUERY_PASSWORD=%s\n' "$query_password"
    printf 'HONEYPOT_CLICKHOUSE_QUERY_PASSWORD_SHA256=%s\n' "$(sha256_text "$query_password")"
  } > "$ENV_FILE"
  chmod 0600 "$ENV_FILE"
  echo "Generated protected ClickHouse credentials: $ENV_FILE"
else
  chown root:root "$ENV_FILE"
  chmod 0600 "$ENV_FILE"
  echo "Preserving existing ClickHouse credentials: $ENV_FILE"
fi

# shellcheck disable=SC1090
set -a
. "$ENV_FILE"
set +a
: "${HONEYPOT_CLICKHOUSE_DATABASE:?missing HONEYPOT_CLICKHOUSE_DATABASE in $ENV_FILE}"
: "${HONEYPOT_CLICKHOUSE_MIGRATE_PASSWORD:?missing HONEYPOT_CLICKHOUSE_MIGRATE_PASSWORD in $ENV_FILE}"
: "${HONEYPOT_CLICKHOUSE_APP_PASSWORD:?missing HONEYPOT_CLICKHOUSE_APP_PASSWORD in $ENV_FILE}"
: "${HONEYPOT_CLICKHOUSE_APP_PASSWORD_SHA256:?missing HONEYPOT_CLICKHOUSE_APP_PASSWORD_SHA256 in $ENV_FILE}"
: "${HONEYPOT_CLICKHOUSE_INGEST_PASSWORD:?missing HONEYPOT_CLICKHOUSE_INGEST_PASSWORD in $ENV_FILE}"
: "${HONEYPOT_CLICKHOUSE_INGEST_PASSWORD_SHA256:?missing HONEYPOT_CLICKHOUSE_INGEST_PASSWORD_SHA256 in $ENV_FILE}"
: "${HONEYPOT_CLICKHOUSE_QUERY_PASSWORD:?missing HONEYPOT_CLICKHOUSE_QUERY_PASSWORD in $ENV_FILE}"
: "${HONEYPOT_CLICKHOUSE_QUERY_PASSWORD_SHA256:?missing HONEYPOT_CLICKHOUSE_QUERY_PASSWORD_SHA256 in $ENV_FILE}"
for digest in \
  "$HONEYPOT_CLICKHOUSE_APP_PASSWORD_SHA256" \
  "$HONEYPOT_CLICKHOUSE_INGEST_PASSWORD_SHA256" \
  "$HONEYPOT_CLICKHOUSE_QUERY_PASSWORD_SHA256"; do
  case "$digest" in
    *[!0-9a-f]*|'') echo "ClickHouse password hashes must be lowercase SHA-256 digests." >&2; exit 2 ;;
  esac
  [ "${#digest}" -eq 64 ] || { echo "ClickHouse password hashes must contain exactly 64 hexadecimal characters." >&2; exit 2; }
done

COMPOSE_PROJECT_NAME=$PROJECT_NAME docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d

container_id=$(COMPOSE_PROJECT_NAME=$PROJECT_NAME docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps -q clickhouse)
if [ -z "$container_id" ]; then
  echo "ClickHouse container was not created." >&2
  exit 1
fi

attempt=0
while [ "$attempt" -lt 60 ]; do
  health=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id" 2>/dev/null || true)
  case "$health" in
    healthy) break ;;
    unhealthy|exited|dead)
      COMPOSE_PROJECT_NAME=$PROJECT_NAME docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" logs --tail=80 clickhouse >&2 || true
      echo "ClickHouse failed its health check: $health" >&2
      exit 1
      ;;
  esac
  attempt=$((attempt + 1))
  sleep 2
done
if [ "$health" != "healthy" ]; then
  COMPOSE_PROJECT_NAME=$PROJECT_NAME docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" logs --tail=80 clickhouse >&2 || true
  echo "Timed out waiting for ClickHouse health." >&2
  exit 1
fi

"$SCRIPT_DIR/migrate-clickhouse.sh" --env-file "$ENV_FILE"
if [ "$SKIP_SMOKE" -eq 0 ]; then
  "$SCRIPT_DIR/smoke-clickhouse.sh" --env-file "$ENV_FILE"
fi

# Enable a newly generated Server configuration only after the engine, schema,
# least-privilege application account, and write/read path have all succeeded.
# Existing settings are never rewritten during an install or upgrade.
if [ ! -f "$ANALYTICS_CONFIG" ]; then
  umask 027
  {
    echo "# Generated by install-clickhouse.sh. MySQL remains the business database."
    echo "analytics:"
    echo "  enabled: true"
    printf '  dsn: "clickhouse://honeynet_app:%s@127.0.0.1:%s/%s"\n' "$HONEYPOT_CLICKHOUSE_APP_PASSWORD" "${HONEYPOT_CLICKHOUSE_NATIVE_PORT:-9000}" "$HONEYPOT_CLICKHOUSE_DATABASE"
    printf '  database: "%s"\n' "$HONEYPOT_CLICKHOUSE_DATABASE"
    echo "  table: \"security_events\""
    echo "  max_open_conns: 10"
    echo "  max_idle_conns: 5"
    echo "  conn_max_lifetime: \"1h\""
    echo "  dial_timeout: \"5s\""
    echo "  read_timeout: \"30s\""
  } > "$ANALYTICS_CONFIG"
  chown root:honeynet "$ANALYTICS_CONFIG"
  chmod 0640 "$ANALYTICS_CONFIG"
  echo "Generated Server analytics configuration: $ANALYTICS_CONFIG"
else
  echo "Preserving existing analytics configuration: $ANALYTICS_CONFIG"
fi

echo "ClickHouse security analytics engine is ready."
echo "HTTP endpoint: 127.0.0.1:${HONEYPOT_CLICKHOUSE_HTTP_PORT:-8123}"
echo "Native endpoint: 127.0.0.1:${HONEYPOT_CLICKHOUSE_NATIVE_PORT:-9000}"
echo "Data volume: ${HONEYPOT_CLICKHOUSE_VOLUME:-honeynet_analytics_clickhouse_data}"
