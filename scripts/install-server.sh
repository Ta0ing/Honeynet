#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
NETWORK_LIB=$SCRIPT_DIR/lib/network.sh
[ -r "$NETWORK_LIB" ] || { echo "Incomplete release package: missing $NETWORK_LIB" >&2; exit 1; }
. "$NETWORK_LIB"

PREFIX=/opt/honeynet
CONFIG_DIR=/etc/honeynet
CONFIG_PATH=$CONFIG_DIR/server.yaml
AGENT_CONFIG=$CONFIG_DIR/builtin-agent.json
AGENT_STATE=/var/lib/honeynet-agent
BUILTIN_AGENT_BIN=$PREFIX/agent/bin/honeynet-agent
BUILTIN_AGENT_GUARD=$PREFIX/agent/libexec/honeynet-agent-guard
LISTEN=:8080
AGENT_LISTEN=:8443
AGENT_PUBLIC_URL=${HONEYPOT_AGENT_PUBLIC_URL:-}
PKI_DIR=/var/lib/honeynet/pki
MYSQL_DSN=${HONEYPOT_DATABASE_DSN:-}
PUBLIC_URL=${HONEYPOT_PUBLIC_URL:-}
CONSOLE_TLS_ENABLED=${HONEYPOT_TLS_ENABLED:-false}
ADMIN_USER=${HONEYPOT_ADMIN_USERNAME:-admin}
ADMIN_PASSWORD=${HONEYPOT_ADMIN_PASSWORD:-}
JWT_SECRET=${HONEYPOT_JWT_SECRET:-}
BUILTIN_TOKEN=${HONEYPOT_BUILTIN_AGENT_TOKEN:-}
IPIP_DB=${HONEYPOT_IPIP_DB_PATH:-}
IPIP_LANGUAGE=${HONEYPOT_IPIP_LANGUAGE:-CN}
INSTALLED_IPIP_DB=/var/lib/honeynet/ipip.ipdb
THREAT_INTEL_ENABLED=${HONEYPOT_THREAT_INTEL_ENABLED:-false}
THREAT_INTEL_STATE=/var/lib/honeynet/threat-intelligence
THREAT_INTEL_DB=${HONEYPOT_THREAT_INTEL_DB_PATH:-$THREAT_INTEL_STATE/intelligence.db}
THREAT_INTEL_URL=${HONEYPOT_THREAT_INTEL_DOWNLOAD_URL:-}
THREAT_INTEL_INTERVAL=${HONEYPOT_THREAT_INTEL_UPDATE_INTERVAL:-24h}
THREAT_INTEL_PASSWORD=${HONEYPOT_THREAT_INTEL_ARCHIVE_PASSWORD:-}
THREAT_INTEL_ENV=$CONFIG_DIR/threat-intelligence.env
START_SERVICES=1
ENABLE_BUILTIN=1
ENABLE_CLICKHOUSE=1
CLICKHOUSE_PREVIOUSLY_RUNNING=0
CLICKHOUSE_BACKUP=
CLICKHOUSE_HAD_DEPLOY=0
CLICKHOUSE_HAD_MIGRATIONS=0
CLICKHOUSE_ENV_EXISTED=0
ANALYTICS_CONFIG_EXISTED=0

usage() {
  echo "Usage: $0 --mysql-dsn DSN [--public-url URL] [--listen :8080|[::]:8080] [--console-tls|--no-console-tls] [--agent-listen :8443|[::]:8443] [--agent-public-url URL] [--admin-user USER] [--admin-password PASSWORD] [--jwt-secret SECRET] [--ipip-db FILE] [--server-only|--skip-builtin-agent] [--skip-clickhouse] [--no-start]"
}

random_hex() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  else
    od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
  fi
}

yaml_quote() {
  escaped=$(printf '%s' "$1" | sed "s/'/''/g")
  printf "'%s'" "$escaped"
}

server_tls_enabled_from_config() {
  awk '
    /^[[:space:]]*#/ { next }
    /^[^[:space:]]/ { in_server = ($0 ~ /^server:[[:space:]]*($|#)/) }
    in_server && $0 ~ /^[[:space:]]+tls_enabled[[:space:]]*:/ {
      value = $0
      sub(/^[^:]*:/, "", value)
      sub(/#.*/, "", value)
      gsub(/[[:space:]"'\'' ]/, "", value)
      print tolower(value)
      found = 1
      exit
    }
    END { if (!found) print "false" }
  ' "$1"
}

# Read the small set of scalar network values that the installer itself needs
# from an existing strict YAML configuration. Upgrades preserve server.yaml,
# so health probes and the final status message must follow the addresses the
# new Server will actually bind instead of falling back to installer defaults.
yaml_scalar_from_config() {
  section=$1
  key=$2
  file=$3
  awk -v wanted_section="$section" -v wanted_key="$key" '
    /^[[:space:]]*#/ || /^[[:space:]]*$/ { next }
    /^[^[:space:]][^:]*:[[:space:]]*($|#)/ {
      current = $0
      sub(/:.*/, "", current)
      next
    }
    current == wanted_section && $0 ~ "^[[:space:]]+" wanted_key "[[:space:]]*:" {
      value = $0
      sub(/^[^:]*:/, "", value)
      sub(/[[:space:]]+#.*/, "", value)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      first = substr(value, 1, 1)
      last = substr(value, length(value), 1)
      if (length(value) >= 2 && ((first == "\"" && last == "\"") || (first == "\047" && last == "\047"))) {
        value = substr(value, 2, length(value) - 2)
      }
      print value
      exit
    }
  ' "$file"
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --mysql-dsn) MYSQL_DSN=$2; shift 2 ;;
    --public-url) PUBLIC_URL=$2; shift 2 ;;
    --listen) LISTEN=$2; shift 2 ;;
    --console-tls) CONSOLE_TLS_ENABLED=true; shift ;;
    --no-console-tls) CONSOLE_TLS_ENABLED=false; shift ;;
    --agent-listen) AGENT_LISTEN=$2; shift 2 ;;
    --agent-public-url) AGENT_PUBLIC_URL=$2; shift 2 ;;
    --admin-user) ADMIN_USER=$2; shift 2 ;;
    --admin-password) ADMIN_PASSWORD=$2; shift 2 ;;
    --jwt-secret) JWT_SECRET=$2; shift 2 ;;
    --ipip-db) IPIP_DB=$2; shift 2 ;;
    --server-only|--skip-builtin-agent) ENABLE_BUILTIN=0; shift ;;
    --skip-clickhouse) ENABLE_CLICKHOUSE=0; shift ;;
    --no-start) START_SERVICES=0; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

case "$(printf '%s' "$CONSOLE_TLS_ENABLED" | tr '[:upper:]' '[:lower:]')" in
  1|true|yes|on) CONSOLE_TLS_ENABLED=1 ;;
  0|false|no|off|'') CONSOLE_TLS_ENABLED=0 ;;
  *) echo "HONEYPOT_TLS_ENABLED must be true or false." >&2; exit 2 ;;
esac

# Upgrades preserve server.yaml, so startup verification must follow the
# installed configuration rather than a newly supplied command-line switch.
if [ -f "$CONFIG_PATH" ]; then
  configured_value=$(yaml_scalar_from_config server addr "$CONFIG_PATH")
  [ -z "$configured_value" ] || LISTEN=$configured_value
  configured_value=$(yaml_scalar_from_config server public_url "$CONFIG_PATH")
  [ -z "$configured_value" ] || PUBLIC_URL=$configured_value
  configured_value=$(yaml_scalar_from_config agent addr "$CONFIG_PATH")
  [ -z "$configured_value" ] || AGENT_LISTEN=$configured_value
  configured_value=$(yaml_scalar_from_config agent public_url "$CONFIG_PATH")
  [ -z "$configured_value" ] || AGENT_PUBLIC_URL=$configured_value
  case "$(server_tls_enabled_from_config "$CONFIG_PATH")" in
    1|true|yes|on) CONSOLE_TLS_ENABLED=1 ;;
    *) CONSOLE_TLS_ENABLED=0 ;;
  esac
fi

LISTEN_PORT=$(honeynet_listen_port "$LISTEN") || { echo "--listen must use host:PORT syntax; bracket IPv6, for example [::]:8080" >&2; exit 2; }
LISTEN_HOST=$(honeynet_listen_host "$LISTEN") || { echo "Invalid --listen address: $LISTEN" >&2; exit 2; }
case "$LISTEN_PORT" in
  ''|*[!0-9]*) echo "Invalid listen port: $LISTEN_PORT" >&2; exit 2 ;;
esac
if [ "$LISTEN_PORT" -lt 1 ] || [ "$LISTEN_PORT" -gt 65535 ]; then
  echo "Listen port must be between 1 and 65535" >&2
  exit 2
fi
AGENT_PORT=$(honeynet_listen_port "$AGENT_LISTEN") || { echo "--agent-listen must use host:PORT syntax; bracket IPv6, for example [::]:8443" >&2; exit 2; }
AGENT_LISTEN_HOST=$(honeynet_listen_host "$AGENT_LISTEN") || { echo "Invalid --agent-listen address: $AGENT_LISTEN" >&2; exit 2; }
case "$AGENT_PORT" in
  ''|*[!0-9]*) echo "Invalid Agent listen port: $AGENT_PORT" >&2; exit 2 ;;
esac
if [ "$AGENT_PORT" -lt 1 ] || [ "$AGENT_PORT" -gt 65535 ]; then
  echo "Agent listen port must be between 1 and 65535" >&2
  exit 2
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "This installer must run as root." >&2
  exit 1
fi
if ! command -v systemctl >/dev/null 2>&1; then
  echo "systemd is required by this installer." >&2
  exit 1
fi
if [ "$START_SERVICES" -eq 1 ] && ! command -v curl >/dev/null 2>&1; then
  echo "curl is required to verify the Server health endpoint before committing an installation." >&2
  exit 1
fi

PACKAGE_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
if [ -f "$PACKAGE_ROOT/SERVER_ONLY" ]; then
  ENABLE_BUILTIN=0
fi
for required in "$PACKAGE_ROOT/bin/honeynet-server" "$PACKAGE_ROOT/web/dist/index.html"; do
  if [ ! -f "$required" ]; then
    echo "Incomplete release package: missing $required" >&2
    exit 1
  fi
done
if [ "$ENABLE_CLICKHOUSE" -eq 1 ]; then
  for required in \
    "$PACKAGE_ROOT/deploy/clickhouse/compose.yaml" \
    "$PACKAGE_ROOT/deploy/clickhouse/init/bootstrap.sh" \
    "$PACKAGE_ROOT/migrations/clickhouse/001_security_events.sql" \
    "$PACKAGE_ROOT/migrations/clickhouse/002_event_rollups.sql" \
    "$PACKAGE_ROOT/scripts/install-clickhouse.sh" \
    "$PACKAGE_ROOT/scripts/migrate-clickhouse.sh" \
    "$PACKAGE_ROOT/scripts/smoke-clickhouse.sh" \
    "$PACKAGE_ROOT/scripts/uninstall-clickhouse.sh"; do
    if [ ! -f "$required" ]; then
      echo "Incomplete ClickHouse deployment assets: missing $required (use --skip-clickhouse only when an external analytics engine is managed separately)" >&2
      exit 1
    fi
  done
  command -v docker >/dev/null 2>&1 || {
    echo "Docker is required for the default ClickHouse deployment. Install Docker Engine with Compose v2, or rerun with --skip-clickhouse." >&2
    exit 1
  }
  docker compose version >/dev/null 2>&1 || {
    echo "Docker Compose v2 is required for the default ClickHouse deployment, or rerun with --skip-clickhouse." >&2
    exit 1
  }
  docker info >/dev/null 2>&1 || {
    echo "Docker daemon is unavailable. Start Docker and retry, or rerun with --skip-clickhouse." >&2
    exit 1
  }
fi
if [ "$ENABLE_BUILTIN" -eq 1 ]; then
  for required in "$PACKAGE_ROOT/bin/honeynet-agent" "$PACKAGE_ROOT/bin/honeynet-agent-guard" "$PACKAGE_ROOT/templates/web/services/config.json"; do
    if [ ! -f "$required" ]; then
      echo "Incomplete built-in Agent package: missing $required (use --server-only to install only Server)" >&2
      exit 1
    fi
  done
fi
if [ -z "$IPIP_DB" ] && [ -f "$PACKAGE_ROOT/geoip/ipip.ipdb" ]; then
  IPIP_DB=$PACKAGE_ROOT/geoip/ipip.ipdb
fi
if [ -n "$IPIP_DB" ] && [ ! -f "$IPIP_DB" ]; then
  echo "IPIP database does not exist: $IPIP_DB" >&2
  exit 2
fi

HOST_IP=$(honeynet_detect_host_ip)
HOST_URL=$(honeynet_format_url_host "${HOST_IP:-127.0.0.1}")
if [ -z "$PUBLIC_URL" ]; then
  PUBLIC_SCHEME=http
  [ "$CONSOLE_TLS_ENABLED" -eq 0 ] || PUBLIC_SCHEME=https
  case "$LISTEN_PORT" in
    80|443) PUBLIC_URL=$PUBLIC_SCHEME://$HOST_URL ;;
    *) PUBLIC_URL=$PUBLIC_SCHEME://$HOST_URL:$LISTEN_PORT ;;
  esac
fi
case "$PUBLIC_URL" in http://*|https://*) ;; *) echo "--public-url must use http:// or https://" >&2; exit 2 ;; esac
if [ "$CONSOLE_TLS_ENABLED" -eq 1 ]; then
  case "$PUBLIC_URL" in https://*) ;; *) echo "--public-url must use https:// when --console-tls is enabled" >&2; exit 2 ;; esac
fi
PUBLIC_HOST=$(honeynet_url_host "$PUBLIC_URL") || { echo "Invalid --public-url; bracket IPv6 hosts, for example http://[2001:db8::10]:8080" >&2; exit 2; }
if [ -z "$AGENT_PUBLIC_URL" ]; then
  AGENT_PUBLIC_HOST=$(honeynet_format_url_host "${PUBLIC_HOST:-${HOST_IP:-127.0.0.1}}")
  AGENT_PUBLIC_URL=https://$AGENT_PUBLIC_HOST:$AGENT_PORT
fi
case "$AGENT_PUBLIC_URL" in https://*) ;; *) echo "--agent-public-url must use https://" >&2; exit 2 ;; esac
AGENT_HOST=$(honeynet_url_host "$AGENT_PUBLIC_URL") || { echo "Invalid --agent-public-url; bracket IPv6 hosts, for example https://[2001:db8::10]:8443" >&2; exit 2; }

CONSOLE_PROBE_HOST=$(honeynet_probe_host "$LISTEN_HOST")
AGENT_PROBE_HOST=$(honeynet_probe_host "$AGENT_LISTEN_HOST")
CONSOLE_PROBE_SCHEME=http
[ "$CONSOLE_TLS_ENABLED" -eq 0 ] || CONSOLE_PROBE_SCHEME=https
CONSOLE_PROBE_URL=$CONSOLE_PROBE_SCHEME://$(honeynet_format_url_host "$CONSOLE_PROBE_HOST"):$LISTEN_PORT
AGENT_PROBE_URL=https://$(honeynet_format_url_host "$AGENT_PROBE_HOST"):$AGENT_PORT

NEW_CONFIG=0
GENERATED_PASSWORD=0
if [ ! -f "$CONFIG_PATH" ]; then
  if [ -z "$MYSQL_DSN" ]; then
    echo "--mysql-dsn is required for the first installation." >&2
    exit 2
  fi
  if [ -z "$ADMIN_PASSWORD" ]; then
    ADMIN_PASSWORD=$(random_hex)
    GENERATED_PASSWORD=1
  fi
  [ -n "$JWT_SECRET" ] || JWT_SECRET=$(random_hex)
  [ -n "$BUILTIN_TOKEN" ] || BUILTIN_TOKEN=$(random_hex)
  NEW_CONFIG=1
fi

if ! getent group honeynet >/dev/null 2>&1; then
  groupadd --system honeynet
fi
if ! id honeynet >/dev/null 2>&1; then
  useradd --system --gid honeynet --home-dir "$PREFIX" --shell /sbin/nologin honeynet
fi

# Bring the independent analytics engine up before interrupting an existing
# Server installation. A ClickHouse bootstrap failure therefore leaves the
# currently installed Server running and available for diagnosis. Stage these
# assets under /opt first so Compose never keeps bind mounts into a temporary
# release extraction directory that an administrator may remove after install.
if [ "$ENABLE_CLICKHOUSE" -eq 1 ]; then
  CLICKHOUSE_BACKUP=$(mktemp -d /var/tmp/honeynet-clickhouse-upgrade.XXXXXX)
  if docker ps -q \
    --filter label=com.docker.compose.project=honeynet-analytics \
    --filter label=com.docker.compose.service=clickhouse | grep -q .; then
    CLICKHOUSE_PREVIOUSLY_RUNNING=1
  fi
  restore_clickhouse_install() {
    [ -d "$CLICKHOUSE_BACKUP" ] || return 0
    if [ -f "$PREFIX/deploy/clickhouse/compose.yaml" ] && [ -f /etc/honeynet/clickhouse.env ]; then
      COMPOSE_PROJECT_NAME=honeynet-analytics docker compose --env-file /etc/honeynet/clickhouse.env -f "$PREFIX/deploy/clickhouse/compose.yaml" down --remove-orphans >&2 || true
    fi
    rm -rf "$PREFIX/deploy/clickhouse" "$PREFIX/migrations/clickhouse"
    for clickhouse_script in install-clickhouse.sh migrate-clickhouse.sh smoke-clickhouse.sh uninstall-clickhouse.sh; do
      rm -f "$PREFIX/scripts/$clickhouse_script"
    done
    [ "$CLICKHOUSE_HAD_DEPLOY" -eq 0 ] || cp -R "$CLICKHOUSE_BACKUP/deploy-clickhouse" "$PREFIX/deploy/clickhouse"
    [ "$CLICKHOUSE_HAD_MIGRATIONS" -eq 0 ] || cp -R "$CLICKHOUSE_BACKUP/migrations-clickhouse" "$PREFIX/migrations/clickhouse"
    if [ -d "$CLICKHOUSE_BACKUP/scripts" ]; then
      install -d -m 0755 "$PREFIX/scripts"
      cp -R "$CLICKHOUSE_BACKUP/scripts/." "$PREFIX/scripts/"
    fi
    if [ "$CLICKHOUSE_ENV_EXISTED" -eq 1 ]; then
      install -d -m 0750 -o root -g honeynet /etc/honeynet
      cp -a "$CLICKHOUSE_BACKUP/clickhouse.env" /etc/honeynet/clickhouse.env
    else
      rm -f /etc/honeynet/clickhouse.env
    fi
    if [ "$ANALYTICS_CONFIG_EXISTED" -eq 1 ]; then
      install -d -m 0750 -o root -g honeynet /etc/honeynet
      cp -a "$CLICKHOUSE_BACKUP/analytics.yaml" /etc/honeynet/analytics.yaml
    else
      rm -f /etc/honeynet/analytics.yaml
    fi
    if [ "$CLICKHOUSE_PREVIOUSLY_RUNNING" -eq 1 ] && [ -f /etc/honeynet/clickhouse.env ] && [ -f "$PREFIX/deploy/clickhouse/compose.yaml" ]; then
      COMPOSE_PROJECT_NAME=honeynet-analytics docker compose --env-file /etc/honeynet/clickhouse.env -f "$PREFIX/deploy/clickhouse/compose.yaml" up -d >&2 || true
    fi
  }
  rollback_clickhouse_assets() {
    status=$?
    trap - EXIT INT TERM HUP
    echo "ClickHouse upgrade failed; restoring the previously installed analytics deployment." >&2
    restore_clickhouse_install
    rm -rf "$CLICKHOUSE_BACKUP"
    exit "$status"
  }
  if [ -d "$PREFIX/deploy/clickhouse" ]; then
    CLICKHOUSE_HAD_DEPLOY=1
    cp -R "$PREFIX/deploy/clickhouse" "$CLICKHOUSE_BACKUP/deploy-clickhouse"
  fi
  if [ -d "$PREFIX/migrations/clickhouse" ]; then
    CLICKHOUSE_HAD_MIGRATIONS=1
    cp -R "$PREFIX/migrations/clickhouse" "$CLICKHOUSE_BACKUP/migrations-clickhouse"
  fi
  install -d -m 0700 "$CLICKHOUSE_BACKUP/scripts"
  for clickhouse_script in install-clickhouse.sh migrate-clickhouse.sh smoke-clickhouse.sh uninstall-clickhouse.sh; do
    [ ! -f "$PREFIX/scripts/$clickhouse_script" ] || cp "$PREFIX/scripts/$clickhouse_script" "$CLICKHOUSE_BACKUP/scripts/$clickhouse_script"
  done
  if [ -f /etc/honeynet/clickhouse.env ]; then
    CLICKHOUSE_ENV_EXISTED=1
    cp -a /etc/honeynet/clickhouse.env "$CLICKHOUSE_BACKUP/clickhouse.env"
  fi
  if [ -f /etc/honeynet/analytics.yaml ]; then
    ANALYTICS_CONFIG_EXISTED=1
    cp -a /etc/honeynet/analytics.yaml "$CLICKHOUSE_BACKUP/analytics.yaml"
  fi
  trap rollback_clickhouse_assets EXIT INT TERM HUP
  install -d -m 0755 "$PREFIX/deploy/clickhouse/init" "$PREFIX/migrations/clickhouse" "$PREFIX/scripts"
  install -m 0644 "$PACKAGE_ROOT/deploy/clickhouse/compose.yaml" "$PREFIX/deploy/clickhouse/compose.yaml"
  install -m 0644 "$PACKAGE_ROOT/deploy/clickhouse/init/bootstrap.sh" "$PREFIX/deploy/clickhouse/init/bootstrap.sh"
  # Keep the installed migration directory an exact copy of the release. A
  # migration renamed or removed from a later release must not remain mounted
  # into ClickHouse and be executed accidentally during the next bootstrap.
  find "$PREFIX/migrations/clickhouse" -maxdepth 1 -type f -name '[0-9][0-9][0-9]_*.sql' -delete
  cp -R "$PACKAGE_ROOT/migrations/clickhouse/." "$PREFIX/migrations/clickhouse/"
  find "$PREFIX/migrations/clickhouse" -type f -exec chmod 0644 {} +
  for clickhouse_script in install-clickhouse.sh migrate-clickhouse.sh smoke-clickhouse.sh uninstall-clickhouse.sh; do
    install -m 0755 "$PACKAGE_ROOT/scripts/$clickhouse_script" "$PREFIX/scripts/$clickhouse_script"
  done
  "$PREFIX/scripts/install-clickhouse.sh"
  chown -R root:root "$PREFIX/deploy/clickhouse" "$PREFIX/migrations/clickhouse" "$PREFIX/scripts"
  trap - EXIT INT TERM HUP
fi

# Snapshot the working Server installation only after ClickHouse has passed its
# migration and write/read smoke. From this point onward any file-copy,
# configuration, systemd, startup, or health-check failure restores the old
# native Server/Agent tree and their previous running state.
INSTALL_BACKUP=$(mktemp -d /var/tmp/honeynet-server-upgrade.XXXXXX)
PREFIX_EXISTED=0
SERVER_WAS_RUNNING=0
AGENT_WAS_RUNNING=0
SERVER_UNIT_EXISTED=0
AGENT_UNIT_EXISTED=0
SERVER_WAS_ENABLED=0
AGENT_WAS_ENABLED=0
CONFIG_EXISTED=0
AGENT_CONFIG_EXISTED=0
PKI_EXISTED=0
IPIP_EXISTED=0
THREAT_INTEL_ENV_EXISTED=0
THREAT_INTEL_STATE_EXISTED=0
if [ -d "$PREFIX" ]; then
  PREFIX_EXISTED=1
  cp -a "$PREFIX" "$INSTALL_BACKUP/prefix"
fi
if [ -f /etc/systemd/system/honeynet-server.service ]; then
  SERVER_UNIT_EXISTED=1
  cp -a /etc/systemd/system/honeynet-server.service "$INSTALL_BACKUP/honeynet-server.service"
fi
if [ -f /etc/systemd/system/honeynet-agent.service ]; then
  AGENT_UNIT_EXISTED=1
  cp -a /etc/systemd/system/honeynet-agent.service "$INSTALL_BACKUP/honeynet-agent.service"
fi
if systemctl is-active --quiet honeynet-server.service 2>/dev/null; then SERVER_WAS_RUNNING=1; fi
if systemctl is-active --quiet honeynet-agent.service 2>/dev/null; then AGENT_WAS_RUNNING=1; fi
if systemctl is-enabled --quiet honeynet-server.service 2>/dev/null; then SERVER_WAS_ENABLED=1; fi
if systemctl is-enabled --quiet honeynet-agent.service 2>/dev/null; then AGENT_WAS_ENABLED=1; fi
if [ -f "$CONFIG_PATH" ]; then
  CONFIG_EXISTED=1
  cp -a "$CONFIG_PATH" "$INSTALL_BACKUP/server.yaml"
fi
if [ -f "$AGENT_CONFIG" ]; then
  AGENT_CONFIG_EXISTED=1
  cp -a "$AGENT_CONFIG" "$INSTALL_BACKUP/builtin-agent.json"
fi
if [ -d "$PKI_DIR" ]; then
  PKI_EXISTED=1
  cp -a "$PKI_DIR" "$INSTALL_BACKUP/pki"
fi
if [ -f "$INSTALLED_IPIP_DB" ]; then
  IPIP_EXISTED=1
  cp -a "$INSTALLED_IPIP_DB" "$INSTALL_BACKUP/ipip.ipdb"
fi
if [ -f "$THREAT_INTEL_ENV" ]; then
  THREAT_INTEL_ENV_EXISTED=1
  cp -a "$THREAT_INTEL_ENV" "$INSTALL_BACKUP/threat-intelligence.env"
fi
if [ -d "$THREAT_INTEL_STATE" ]; then
  THREAT_INTEL_STATE_EXISTED=1
  cp -a "$THREAT_INTEL_STATE" "$INSTALL_BACKUP/threat-intelligence-state"
fi

rollback_server_install() {
  status=$?
  trap - EXIT INT TERM HUP
  echo "Honeynet Server installation failed; restoring the previous native runtime." >&2
  systemctl stop honeynet-agent.service >/dev/null 2>&1 || true
  systemctl stop honeynet-server.service >/dev/null 2>&1 || true
  if [ "$PREFIX_EXISTED" -eq 1 ] && [ -d "$INSTALL_BACKUP/prefix" ]; then
    rm -rf "$PREFIX"
    cp -a "$INSTALL_BACKUP/prefix" "$PREFIX"
  else
    rm -rf "$PREFIX"
  fi
  if [ "$CONFIG_EXISTED" -eq 1 ]; then
    install -d -m 0750 -o root -g honeynet "$CONFIG_DIR"
    cp -a "$INSTALL_BACKUP/server.yaml" "$CONFIG_PATH"
  else
    rm -f "$CONFIG_PATH"
  fi
  if [ "$AGENT_CONFIG_EXISTED" -eq 1 ]; then
    install -d -m 0750 -o root -g honeynet "$CONFIG_DIR"
    cp -a "$INSTALL_BACKUP/builtin-agent.json" "$AGENT_CONFIG"
  else
    rm -f "$AGENT_CONFIG"
  fi
  if [ "$PKI_EXISTED" -eq 1 ]; then
    rm -rf "$PKI_DIR"
    cp -a "$INSTALL_BACKUP/pki" "$PKI_DIR"
  else
    rm -rf "$PKI_DIR"
  fi
  if [ "$IPIP_EXISTED" -eq 1 ]; then
    install -d -m 0755 "$(dirname "$INSTALLED_IPIP_DB")"
    cp -a "$INSTALL_BACKUP/ipip.ipdb" "$INSTALLED_IPIP_DB"
  else
    rm -f "$INSTALLED_IPIP_DB"
  fi
  if [ "$THREAT_INTEL_ENV_EXISTED" -eq 1 ]; then
    cp -a "$INSTALL_BACKUP/threat-intelligence.env" "$THREAT_INTEL_ENV"
  else
    rm -f "$THREAT_INTEL_ENV"
  fi
  if [ "$THREAT_INTEL_STATE_EXISTED" -eq 1 ]; then
    rm -rf "$THREAT_INTEL_STATE"
    cp -a "$INSTALL_BACKUP/threat-intelligence-state" "$THREAT_INTEL_STATE"
  else
    rm -rf "$THREAT_INTEL_STATE"
  fi
  if [ "$ENABLE_CLICKHOUSE" -eq 1 ]; then
    restore_clickhouse_install
    rm -rf "$CLICKHOUSE_BACKUP"
  fi
  if [ "$SERVER_UNIT_EXISTED" -eq 1 ]; then
    cp -a "$INSTALL_BACKUP/honeynet-server.service" /etc/systemd/system/honeynet-server.service
  else
    rm -f /etc/systemd/system/honeynet-server.service
  fi
  if [ "$AGENT_UNIT_EXISTED" -eq 1 ]; then
    cp -a "$INSTALL_BACKUP/honeynet-agent.service" /etc/systemd/system/honeynet-agent.service
  else
    rm -f /etc/systemd/system/honeynet-agent.service
  fi
  systemctl daemon-reload >/dev/null 2>&1 || true
  if [ "$SERVER_WAS_ENABLED" -eq 1 ]; then
    systemctl enable honeynet-server.service >/dev/null 2>&1 || true
  else
    systemctl disable honeynet-server.service >/dev/null 2>&1 || true
  fi
  if [ "$AGENT_WAS_ENABLED" -eq 1 ]; then
    systemctl enable honeynet-agent.service >/dev/null 2>&1 || true
  else
    systemctl disable honeynet-agent.service >/dev/null 2>&1 || true
  fi
  [ "$SERVER_WAS_RUNNING" -eq 0 ] || systemctl start honeynet-server.service >/dev/null 2>&1 || true
  [ "$AGENT_WAS_RUNNING" -eq 0 ] || systemctl start honeynet-agent.service >/dev/null 2>&1 || true
  rm -rf "$INSTALL_BACKUP"
  exit "$status"
}
trap rollback_server_install EXIT INT TERM HUP

if [ "$ENABLE_BUILTIN" -eq 1 ]; then
  systemctl stop honeynet-agent.service >/dev/null 2>&1 || true
fi
systemctl stop honeynet-server.service >/dev/null 2>&1 || true

install -d -m 0755 "$PREFIX/bin" "$PREFIX/web" "$PREFIX/downloads" "$PREFIX/rules/builtin" "$PREFIX/scripts"
install -m 0755 "$PACKAGE_ROOT/bin/honeynet-server" "$PREFIX/bin/honeynet-server"
cp -R "$PACKAGE_ROOT/web/dist" "$PREFIX/web/"
# cp applies the invoking user's umask when it creates files. Root commonly uses
# 0077 on hardened hosts, which would make the console unreadable by the
# unprivileged honeynet-server service and cause GET / to return 403.
find "$PREFIX/web" -type d -exec chmod 0755 {} +
find "$PREFIX/web" -type f -exec chmod 0644 {} +
cp -R "$PACKAGE_ROOT/rules/builtin/." "$PREFIX/rules/builtin/"
find "$PREFIX/rules" -type d -exec chmod 0755 {} +
find "$PREFIX/rules" -type f -exec chmod 0644 {} +
if [ "$ENABLE_BUILTIN" -eq 1 ]; then
  install -d -m 0755 "$PREFIX/templates/web"
  install -d -m 0755 "$PREFIX/agent/bin" "$PREFIX/agent/libexec"
  install -m 0755 "$PACKAGE_ROOT/bin/honeynet-agent" "$BUILTIN_AGENT_BIN"
  install -m 0755 "$PACKAGE_ROOT/bin/honeynet-agent-guard" "$BUILTIN_AGENT_GUARD"
  rm -rf "$PREFIX/templates/web/services"
  cp -R "$PACKAGE_ROOT/templates/web/services" "$PREFIX/templates/web/"
fi
if [ -d "$PACKAGE_ROOT/downloads" ]; then
  cp -R "$PACKAGE_ROOT/downloads/." "$PREFIX/downloads/"
  find "$PREFIX/downloads" -type d -exec chmod 0755 {} +
  find "$PREFIX/downloads" -type f -exec chmod 0644 {} +
fi
chown -R root:root "$PREFIX"

install -d -m 0750 -o root -g honeynet "$CONFIG_DIR"
if [ "$ENABLE_BUILTIN" -eq 1 ]; then
  install -d -m 0700 -o root -g root "$AGENT_STATE"
fi
install -d -m 0700 -o honeynet -g honeynet "$PKI_DIR"
# Keep the downloadable intelligence database in a dedicated writable state
# directory. /var/lib/honeynet itself remains root-owned so the Server cannot
# replace PKI or unrelated state even though systemd permits writes below it.
install -d -m 0750 -o honeynet -g honeynet "$THREAT_INTEL_STATE"
if [ -n "$IPIP_DB" ]; then
  if [ "$IPIP_DB" != "$INSTALLED_IPIP_DB" ]; then
    install -m 0640 -o root -g honeynet "$IPIP_DB" "$INSTALLED_IPIP_DB"
  fi
fi
if [ "$NEW_CONFIG" -eq 1 ]; then
  {
    echo "server:"
    printf '  addr: %s\n' "$(yaml_quote "$LISTEN")"
    printf '  public_url: %s\n' "$(yaml_quote "$PUBLIC_URL")"
    if [ "$CONSOLE_TLS_ENABLED" -eq 1 ]; then
      echo "  tls_enabled: true"
    else
      echo "  tls_enabled: false"
    fi
    printf '  web_dist: %s\n' "$(yaml_quote "$PREFIX/web/dist")"
    printf '  downloads_dir: %s\n' "$(yaml_quote "$PREFIX/downloads")"
    echo "agent:"
    printf '  addr: %s\n' "$(yaml_quote "$AGENT_LISTEN")"
    printf '  public_url: %s\n' "$(yaml_quote "$AGENT_PUBLIC_URL")"
    printf '  pki_dir: %s\n' "$(yaml_quote "$PKI_DIR")"
    echo "  tls_names:"
    printf '    - %s\n' "$(yaml_quote "${AGENT_HOST:-${HOST_IP:-127.0.0.1}}")"
	printf '    - %s\n' "$(yaml_quote "$AGENT_PROBE_HOST")"
	printf '    - %s\n' "$(yaml_quote "$CONSOLE_PROBE_HOST")"
    echo "    - 'localhost'"
    echo "    - '127.0.0.1'"
	echo "    - '::1'"
    echo "  certificate_validity: '9600h'"
    echo "  renew_before: '720h'"
    echo "geoip:"
    if [ -f "$INSTALLED_IPIP_DB" ]; then
      printf '  ipip_db_path: %s\n' "$(yaml_quote "$INSTALLED_IPIP_DB")"
    else
      echo "  ipip_db_path: ''"
    fi
    printf '  language: %s\n' "$(yaml_quote "$IPIP_LANGUAGE")"
    echo "threat_intelligence:"
    printf '  enabled: %s\n' "$THREAT_INTEL_ENABLED"
    printf '  database_path: %s\n' "$(yaml_quote "$THREAT_INTEL_DB")"
    printf '  download_url: %s\n' "$(yaml_quote "$THREAT_INTEL_URL")"
    printf '  update_interval: %s\n' "$(yaml_quote "$THREAT_INTEL_INTERVAL")"
    echo "detection:"
    printf '  rules_dir: %s\n' "$(yaml_quote "$PREFIX/rules/builtin")"
    echo "ai:"
    echo "  enabled: false"
    echo "  provider: 'openai-compatible'"
    echo "  base_url: ''"
    echo "  api_key: ''"
    echo "  model: ''"
    echo "  timeout: '45s'"
    echo "  send_raw_packet: false"
    echo "database:"
    printf '  dsn: %s\n' "$(yaml_quote "$MYSQL_DSN")"
    echo "auth:"
    printf '  jwt_secret: %s\n' "$(yaml_quote "$JWT_SECRET")"
    echo "  jwt_expires: '8h'"
    printf '  admin_username: %s\n' "$(yaml_quote "$ADMIN_USER")"
    printf '  admin_password: %s\n' "$(yaml_quote "$ADMIN_PASSWORD")"
    echo "builtin_agent:"
    if [ "$ENABLE_BUILTIN" -eq 1 ]; then
      printf '  token: %s\n' "$(yaml_quote "$BUILTIN_TOKEN")"
    else
      echo "  token: ''"
    fi
    echo "cors:"
    echo "  origins:"
    printf '    - %s\n' "$(yaml_quote "$PUBLIC_URL")"
  } > "$CONFIG_PATH"
  chown root:honeynet "$CONFIG_PATH"
  chmod 0640 "$CONFIG_PATH"

  if [ "$ENABLE_BUILTIN" -eq 1 ]; then
    "$BUILTIN_AGENT_BIN" \
      --config "$AGENT_CONFIG" \
      --server "$CONSOLE_PROBE_URL" \
      --agent-url "$AGENT_PROBE_URL" \
      --node-id 00000000-0000-4000-8000-000000000001 \
      --registration-token "$BUILTIN_TOKEN" \
      --ca-cert "$PKI_DIR/ca.crt" \
      --state-dir "$AGENT_STATE" \
	  --template-root "$PREFIX/templates/web/services" \
      --init-only
  fi
else
  echo "Preserving existing configuration: $CONFIG_PATH"
fi

# The archive password is intentionally separate from server.yaml so backups,
# screenshots and configuration APIs cannot expose it. Existing secret files
# are preserved on upgrades unless the operator explicitly provides a value.
if [ -n "$THREAT_INTEL_PASSWORD" ]; then
  umask 077
  printf 'HONEYPOT_THREAT_INTEL_ARCHIVE_PASSWORD=%s\n' "$THREAT_INTEL_PASSWORD" > "$THREAT_INTEL_ENV"
  chown root:honeynet "$THREAT_INTEL_ENV"
  chmod 0640 "$THREAT_INTEL_ENV"
elif [ ! -e "$THREAT_INTEL_ENV" ]; then
  install -m 0640 -o root -g honeynet /dev/null "$THREAT_INTEL_ENV"
fi

install -m 0644 "$PACKAGE_ROOT/deploy/systemd/honeynet-server.service" /etc/systemd/system/honeynet-server.service
if [ "$ENABLE_BUILTIN" -eq 1 ]; then
  install -m 0644 "$PACKAGE_ROOT/deploy/systemd/honeynet-agent.service" /etc/systemd/system/honeynet-agent.service
fi
systemctl daemon-reload
systemctl enable honeynet-server.service >/dev/null
if [ "$ENABLE_BUILTIN" -eq 1 ] && [ -f "$AGENT_CONFIG" ]; then
  systemctl enable honeynet-agent.service >/dev/null
fi

if [ "$START_SERVICES" -eq 1 ]; then
  systemctl start honeynet-server.service
  console_healthcheck() {
    if [ "$CONSOLE_TLS_ENABLED" -eq 1 ]; then
      curl --fail --silent --show-error --globoff --noproxy '*' --max-time 2 --cacert "$PKI_DIR/ca.crt" "$CONSOLE_PROBE_URL/healthz" >/dev/null 2>&1
    else
      curl --fail --silent --show-error --globoff --noproxy '*' --max-time 2 "$CONSOLE_PROBE_URL/healthz" >/dev/null 2>&1
    fi
  }
  health_attempt=0
  while [ "$health_attempt" -lt 60 ]; do
    if console_healthcheck && \
       curl --fail --silent --show-error --globoff --noproxy '*' --max-time 2 --cacert "$PKI_DIR/ca.crt" "$AGENT_PROBE_URL/healthz" >/dev/null 2>&1; then
      break
    fi
    if ! systemctl is-active --quiet honeynet-server.service; then
      echo "Honeynet Server stopped before passing its health check." >&2
      exit 1
    fi
    health_attempt=$((health_attempt + 1))
    sleep 2
  done
  if [ "$health_attempt" -ge 60 ]; then
    echo "Honeynet console and Agent gateway did not both pass /healthz within 120 seconds." >&2
    exit 1
  fi
  if [ "$ENABLE_BUILTIN" -eq 1 ] && [ -f "$AGENT_CONFIG" ]; then
    systemctl start honeynet-agent.service
  fi
elif [ "$ENABLE_CLICKHOUSE" -eq 1 ] && [ "$CLICKHOUSE_PREVIOUSLY_RUNNING" -eq 0 ]; then
  # --no-start prepares and validates the schema, then restores the original
  # stopped state instead of leaving a newly-created analytics container up.
  COMPOSE_PROJECT_NAME=honeynet-analytics docker compose --env-file /etc/honeynet/clickhouse.env -f "$PREFIX/deploy/clickhouse/compose.yaml" stop clickhouse >/dev/null
fi

trap - EXIT INT TERM HUP
rm -rf "$INSTALL_BACKUP"
[ ! -n "$CLICKHOUSE_BACKUP" ] || rm -rf "$CLICKHOUSE_BACKUP"

echo "Honeynet Server installed at $PREFIX"
echo "Console: $PUBLIC_URL"
if [ "$CONSOLE_TLS_ENABLED" -eq 1 ]; then
  echo "Console TLS uses the Honeynet CA at $PKI_DIR/ca.crt; import that CA into administrator browsers before opening the console."
fi
echo "Administrator: $ADMIN_USER"
if [ "$GENERATED_PASSWORD" -eq 1 ]; then
  echo "Generated administrator password: $ADMIN_PASSWORD"
  echo "Save this password now; it will not be displayed again."
fi
if [ "$ENABLE_BUILTIN" -eq 1 ]; then
  echo "Status: systemctl status honeynet-server honeynet-agent"
else
  echo "Status: systemctl status honeynet-server"
fi
