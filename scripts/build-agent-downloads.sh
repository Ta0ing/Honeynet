#!/bin/sh
set -eu

VERSION=${1:-0.24.0-dev}
OUTPUT_ARG=${2:-downloads}
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
: "${GOTOOLCHAIN:=go1.26.5}"
export GOTOOLCHAIN
. "$SCRIPT_DIR/lib/go-version.sh"
honeynet_require_go_1265
case "$OUTPUT_ARG" in
  /*) OUTPUT=$OUTPUT_ARG ;;
  *) OUTPUT=$ROOT/$OUTPUT_ARG ;;
esac
LDFLAGS="-s -w -X main.version=$VERSION"

[ -f "$ROOT/honeypot-templates-server/services/config.json" ] || { echo "Missing honeypot-templates-server/services/config.json" >&2; exit 1; }
command -v zip >/dev/null 2>&1 || { echo "zip is required" >&2; exit 1; }
mkdir -p "$OUTPUT"
cd "$ROOT"

build_agent() {
  target_os=$1
  target_arch=$2
  suffix=
  if [ "$target_os" = windows ]; then suffix=.exe; fi
  if [ "$target_arch" = arm ]; then
    CGO_ENABLED=0 GOOS=$target_os GOARCH=$target_arch GOARM=7 go build -trimpath -ldflags "$LDFLAGS" -o "$OUTPUT/honeynet-agent-$target_os-$target_arch$suffix" ./cmd/agent
  else
    CGO_ENABLED=0 GOOS=$target_os GOARCH=$target_arch go build -trimpath -ldflags "$LDFLAGS" -o "$OUTPUT/honeynet-agent-$target_os-$target_arch$suffix" ./cmd/agent
  fi
  if [ "$target_os" = linux ]; then
    if [ "$target_arch" = arm ]; then
      CGO_ENABLED=0 GOOS=$target_os GOARCH=$target_arch GOARM=7 go build -trimpath -ldflags "$LDFLAGS" -o "$OUTPUT/honeynet-agent-guard-$target_os-$target_arch" ./cmd/agent-guard
    else
      CGO_ENABLED=0 GOOS=$target_os GOARCH=$target_arch go build -trimpath -ldflags "$LDFLAGS" -o "$OUTPUT/honeynet-agent-guard-$target_os-$target_arch" ./cmd/agent-guard
    fi
  fi
}

build_agent linux 386
build_agent linux amd64
build_agent linux arm
build_agent linux arm64
build_agent linux loong64
build_agent windows 386
build_agent windows amd64

TEMP_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/honeynet-templates.XXXXXX")
trap 'rm -rf "$TEMP_ROOT"' EXIT INT TERM
mkdir -p "$TEMP_ROOT/web"
cp -R "$ROOT/honeypot-templates-server/services" "$TEMP_ROOT/web/"
cp "$ROOT/LICENSE" "$TEMP_ROOT/web/"
cp "$ROOT/THIRD_PARTY_NOTICES.md" "$TEMP_ROOT/web/"
find "$TEMP_ROOT" -name .DS_Store -type f -delete
(cd "$TEMP_ROOT/web" && tar -czf "$OUTPUT/honeypot-templates-server.tar.gz" services LICENSE THIRD_PARTY_NOTICES.md)
(cd "$TEMP_ROOT/web" && zip -qr "$OUTPUT/honeypot-templates-server.zip" services LICENSE THIRD_PARTY_NOTICES.md)

echo "Agent downloads prepared at $OUTPUT"
