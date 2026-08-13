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
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PACKAGE_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
COMPOSE_FILE=$PACKAGE_ROOT/deploy/clickhouse/compose.yaml
[ -f "$COMPOSE_FILE" ] || { echo "ClickHouse Compose file not found: $COMPOSE_FILE" >&2; exit 1; }

container_id=$(COMPOSE_PROJECT_NAME=$PROJECT_NAME docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps -q clickhouse)
[ -n "$container_id" ] || { echo "ClickHouse container is not running." >&2; exit 1; }
docker exec "$container_id" bash /opt/honeynet-clickhouse/bootstrap.sh
