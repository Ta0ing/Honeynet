#!/bin/sh
set -eu

ENV_FILE=/etc/honeynet/clickhouse.env
PROJECT_NAME=honeynet-analytics
usage() { echo "Usage: $0 [--env-file FILE]"; }
while [ "$#" -gt 0 ]; do
  case "$1" in
    --env-file) ENV_FILE=$2; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

[ -f "$ENV_FILE" ] || { echo "ClickHouse environment file not found: $ENV_FILE" >&2; exit 1; }
# shellcheck disable=SC1090
set -a
. "$ENV_FILE"
set +a
: "${HONEYPOT_CLICKHOUSE_APP_PASSWORD:?missing HONEYPOT_CLICKHOUSE_APP_PASSWORD in $ENV_FILE}"
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PACKAGE_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
COMPOSE_FILE=$PACKAGE_ROOT/deploy/clickhouse/compose.yaml
container_id=$(COMPOSE_PROJECT_NAME=$PROJECT_NAME docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps -q clickhouse)
[ -n "$container_id" ] || { echo "ClickHouse container is not running." >&2; exit 1; }

smoke_id="smoke-$(date -u +%Y%m%dT%H%M%SZ)-$$"
docker exec -e HONEYPOT_SMOKE_EVENT_ID="$smoke_id" -e HONEYPOT_CLICKHOUSE_APP_PASSWORD="$HONEYPOT_CLICKHOUSE_APP_PASSWORD" "$container_id" bash -euc '
  database=${CLICKHOUSE_DB:-honeynet_analytics}
  app=(clickhouse-client --host 127.0.0.1 --user honeynet_app --password "$HONEYPOT_CLICKHOUSE_APP_PASSWORD" --database "$database")
  migrate=(clickhouse-client --host 127.0.0.1 --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$database")
  cleanup() {
    "${migrate[@]}" --query "ALTER TABLE security_events DELETE WHERE event_id = {event_id:String}" --param_event_id "$HONEYPOT_SMOKE_EVENT_ID" >/dev/null 2>&1 || true
  }
  trap cleanup EXIT
  now=$(date -u "+%Y-%m-%d %H:%M:%S.000")
  printf "%s\n" "{\"event_id\":\"$HONEYPOT_SMOKE_EVENT_ID\",\"node_id\":\"smoke-node\",\"pot_id\":\"\",\"decoy_id\":\"\",\"service\":\"honeynet\",\"event_type\":\"system.smoke\",\"event_time\":\"$now\",\"ingested_at\":\"$now\",\"src_ip\":\"127.0.0.1\",\"src_port\":0,\"dst_ip\":\"127.0.0.1\",\"dst_port\":0,\"geo\":\"本机\",\"asn\":\"\",\"raw_packet\":\"\",\"payload\":\"{}\",\"tags\":\"[\\\"system-smoke\\\"]\",\"detections\":\"[]\",\"agent_rule_revision\":0,\"server_rule_revision\":0,\"session_id\":\"\",\"has_credential\":0,\"credential_username\":\"\",\"credential_password\":\"\",\"credential_auth_response\":\"\",\"credential_mechanism\":\"\",\"record_version\":1}" | "${app[@]}" --query "INSERT INTO security_events FORMAT JSONEachRow"
  count=$("${app[@]}" --query "SELECT uniqExact(event_id) FROM security_events WHERE event_id = {event_id:String}" --param_event_id "$HONEYPOT_SMOKE_EVENT_ID")
  [ "$count" = 1 ] || { echo "ClickHouse smoke query expected one event, got $count" >&2; exit 1; }
  "${app[@]}" --query "SELECT count() FROM security_events_daily_mv WHERE day >= today()" >/dev/null
  schema_version=$("${app[@]}" --query "SELECT toUInt32(ifNull(max(version), 0)) FROM schema_migrations FINAL")
  [ "$schema_version" -ge 2 ] || { echo "ClickHouse smoke expected schema version 2 or newer, got $schema_version" >&2; exit 1; }
'

echo "ClickHouse write/query smoke passed: $smoke_id"
