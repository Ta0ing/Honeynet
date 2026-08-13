#!/bin/sh
set -eu

SERVER_URL=${HONEYPOT_PUBLIC_URL:-}
AGENT_URL=${HONEYPOT_AGENT_PUBLIC_URL:-}
CA_SHA256=${HONEYPOT_CA_SHA256:-}
NODE_ID=${HONEYPOT_NODE_ID:-}
REGISTRATION_TOKEN=${HONEYPOT_REGISTRATION_TOKEN:-}
START_SERVICE=1

PREFIX=/opt/honeynet-agent
BIN=$PREFIX/bin/honeynet-agent
GUARD=$PREFIX/libexec/honeynet-agent-guard
CONFIG_DIR=/etc/honeynet
CONFIG_PATH=$CONFIG_DIR/agent.json
STATE_DIR=/var/lib/honeynet-agent
TEMPLATE_BASE=$PREFIX/templates/web

usage() {
  echo "Usage: $0 --server URL --agent-url URL --ca-sha256 SHA256 --node-id ID --token TOKEN [--no-start]"
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --server) SERVER_URL=$2; shift 2 ;;
    --agent-url) AGENT_URL=$2; shift 2 ;;
    --ca-sha256) CA_SHA256=$2; shift 2 ;;
    --node-id) NODE_ID=$2; shift 2 ;;
    --token) REGISTRATION_TOKEN=$2; shift 2 ;;
    --no-start) START_SERVICE=0; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

[ -n "$SERVER_URL" ] && [ -n "$AGENT_URL" ] && [ -n "$CA_SHA256" ] && [ -n "$NODE_ID" ] && [ -n "$REGISTRATION_TOKEN" ] || {
  echo "server, agent-url, ca-sha256, node-id and token are required" >&2
  exit 2
}
[ "$(id -u)" -eq 0 ] || { echo "Honeynet Agent installer must run as root" >&2; exit 1; }
command -v systemctl >/dev/null 2>&1 || { echo "systemd is required" >&2; exit 1; }

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PACKAGE_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
SOURCE_BIN=$PACKAGE_ROOT/bin/honeynet-agent
SOURCE_GUARD=$PACKAGE_ROOT/bin/honeynet-agent-guard
SOURCE_TEMPLATES=$PACKAGE_ROOT/templates/web/services
SERVICE_UNIT=$PACKAGE_ROOT/deploy/systemd/honeynet-node-agent.service
for required in "$SOURCE_BIN" "$SOURCE_GUARD" "$SOURCE_TEMPLATES/config.json" "$SERVICE_UNIT"; do
  [ -f "$required" ] || { echo "Incomplete Agent package: missing $required" >&2; exit 1; }
done

systemctl stop honeynet-agent.service >/dev/null 2>&1 || true
install -d -m 0700 -o root -g root "$CONFIG_DIR" "$STATE_DIR"
install -d -m 0755 "$PREFIX/bin" "$PREFIX/libexec" "$TEMPLATE_BASE"
install -m 0755 "$SOURCE_BIN" "$BIN.new"
mv -f "$BIN.new" "$BIN"
install -m 0755 "$SOURCE_GUARD" "$GUARD"
rm -rf "$TEMPLATE_BASE/services"
cp -R "$SOURCE_TEMPLATES" "$TEMPLATE_BASE/services"
chown -R root:root "$TEMPLATE_BASE"

"$BIN" \
  --config "$CONFIG_PATH" \
  --server "$SERVER_URL" \
  --agent-url "$AGENT_URL" \
  --ca-sha256 "$CA_SHA256" \
  --node-id "$NODE_ID" \
  --registration-token "$REGISTRATION_TOKEN" \
  --force-enroll \
  --state-dir "$STATE_DIR" \
  --template-root "$TEMPLATE_BASE/services" \
  --init-only

# Complete the pinned-CA enrollment before registering the long-running
# service. This is both an end-to-end gateway probe and an early, actionable
# failure for IPv4/IPv6 URL or reachability problems.
"$BIN" --config "$CONFIG_PATH" --enroll-only

install -m 0644 "$SERVICE_UNIT" /etc/systemd/system/honeynet-agent.service
systemctl daemon-reload
systemctl enable honeynet-agent.service >/dev/null
if [ "$START_SERVICE" -eq 1 ]; then
  systemctl start honeynet-agent.service
fi

echo "Honeynet Agent installed for node $NODE_ID"
echo "Status: systemctl status honeynet-agent"
