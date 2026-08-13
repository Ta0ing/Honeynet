#!/bin/sh
set -eu

VERSION=${1:-0.24.0-dev}
TARGET_OS=${2:-linux}
TARGET_ARCH=${3:-amd64}
case "$TARGET_OS/$TARGET_ARCH" in
  linux/386|linux/amd64|linux/arm|linux/arm64|linux/loong64|windows/386|windows/amd64) ;;
  *) echo "Supported Agent targets: linux/386, linux/amd64, linux/arm, linux/arm64, linux/loong64, windows/386, windows/amd64." >&2; exit 2 ;;
esac

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
: "${GOTOOLCHAIN:=go1.26.5}"
export GOTOOLCHAIN
. "$SCRIPT_DIR/lib/go-version.sh"
honeynet_require_go_1265
OUTPUT=$ROOT/dist
NAME=honeynet-agent-$VERSION-$TARGET_OS-$TARGET_ARCH
STAGE=$OUTPUT/$NAME
LDFLAGS="-s -w -X main.version=$VERSION"

rm -rf "$STAGE"
mkdir -p "$STAGE/bin" "$STAGE/scripts" "$STAGE/templates/web"
[ -f "$ROOT/honeypot-templates-server/services/config.json" ] || { echo "Missing honeypot-templates-server/services/config.json" >&2; exit 1; }

suffix=
if [ "$TARGET_OS" = windows ]; then suffix=.exe; fi
cd "$ROOT"
if [ "$TARGET_ARCH" = arm ]; then
  CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH GOARM=7 go build -trimpath -ldflags "$LDFLAGS" -o "$STAGE/bin/honeynet-agent$suffix" ./cmd/agent
else
  CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -trimpath -ldflags "$LDFLAGS" -o "$STAGE/bin/honeynet-agent$suffix" ./cmd/agent
fi
if [ "$TARGET_OS" = linux ]; then
  if [ "$TARGET_ARCH" = arm ]; then
    CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH GOARM=7 go build -trimpath -ldflags "$LDFLAGS" -o "$STAGE/bin/honeynet-agent-guard" ./cmd/agent-guard
  else
    CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -trimpath -ldflags "$LDFLAGS" -o "$STAGE/bin/honeynet-agent-guard" ./cmd/agent-guard
  fi
fi

cp -R "$ROOT/honeypot-templates-server/services" "$STAGE/templates/web/"
find "$STAGE/templates" -name .DS_Store -type f -delete
cp "$ROOT/README.md" "$STAGE/"
cp "$ROOT/LICENSE" "$STAGE/"
cp "$ROOT/THIRD_PARTY_NOTICES.md" "$STAGE/"
if [ "$TARGET_OS" = linux ]; then
  mkdir -p "$STAGE/deploy/systemd"
  cp "$ROOT/scripts/install-agent.sh" "$STAGE/scripts/"
  cp "$ROOT/deploy/systemd/honeynet-node-agent.service" "$STAGE/deploy/systemd/"
  chmod 0755 "$STAGE/scripts/install-agent.sh"
else
  cp "$ROOT/scripts/install-agent.ps1" "$STAGE/scripts/"
fi

cd "$OUTPUT"
if [ "$TARGET_OS" = windows ]; then
  command -v zip >/dev/null 2>&1 || { echo "zip is required to build a Windows Agent release" >&2; exit 1; }
  archive=$NAME.zip
  zip -qr -FS "$archive" "$NAME"
else
  archive=$NAME.tar.gz
  tar -czf "$archive" "$NAME"
fi
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "$archive" > "$archive.sha256"
else
  shasum -a 256 "$archive" > "$archive.sha256"
fi
echo "$OUTPUT/$archive"
