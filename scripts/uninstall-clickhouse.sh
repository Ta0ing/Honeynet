#!/bin/sh
set -eu

ENV_FILE=/etc/honeynet/clickhouse.env
PROJECT_NAME=honeynet-analytics
usage() { echo "Usage: $0 [--env-file FILE] [--project-name NAME]"; }
while [ "$#" -gt 0 ]; do
  case "$1" in
    --env-file) ENV_FILE=$2; shift 2 ;;
    --project-name) PROJECT_NAME=$2; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done
case "$PROJECT_NAME" in
  ''|*[!A-Za-z0-9_.-]*) echo "Invalid Compose project name: $PROJECT_NAME" >&2; exit 2 ;;
esac

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PACKAGE_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
COMPOSE_FILE=$PACKAGE_ROOT/deploy/clickhouse/compose.yaml
[ -f "$COMPOSE_FILE" ] || { echo "ClickHouse Compose file not found: $COMPOSE_FILE" >&2; exit 1; }

if [ -f "$ENV_FILE" ]; then
  COMPOSE_PROJECT_NAME=$PROJECT_NAME docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" down --remove-orphans
else
  echo "Credentials file is absent; stopping Honeynet analytics project $PROJECT_NAME without deleting volumes."
  # Ignore inherited Compose variables here. Without the protected env file we
  # may only target the deployment's fixed default project, never an arbitrary
  # project selected by the caller's shell.
  container_ids=$(docker ps -aq \
    --filter label=com.docker.compose.project="$PROJECT_NAME" \
    --filter label=com.docker.compose.service=clickhouse)
  if [ -n "$container_ids" ]; then
    # The Compose file deliberately requires secrets during interpolation, so
    # remove only containers carrying our fixed project label when the env file
    # has already been lost. Never select or remove a volume here.
    docker rm -f $container_ids >/dev/null
  fi
  network_ids=$(docker network ls -q \
    --filter label=com.docker.compose.project="$PROJECT_NAME" \
    --filter label=com.docker.compose.network=default)
  if [ -n "$network_ids" ]; then
    docker network rm $network_ids >/dev/null
  fi
fi

echo "ClickHouse containers and project network were removed."
echo "Named ClickHouse data volumes and configuration files were preserved."
