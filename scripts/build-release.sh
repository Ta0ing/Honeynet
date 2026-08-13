#!/bin/sh
set -eu

# BSD tar/cp may preserve macOS resource-fork and provenance xattrs as PAX
# records. Linux extractors ignore them but emit noisy warnings during an
# otherwise clean Server installation. Release artifacts only need portable
# file contents and Unix modes.
export COPYFILE_DISABLE=1
TAR_PORTABLE_FLAGS=
if tar --help 2>&1 | grep -q -- '--no-mac-metadata'; then
  TAR_PORTABLE_FLAGS='--no-xattrs --no-mac-metadata'
fi

VERSION=${1:-0.24.0-dev}
TARGET_OS=${2:-linux}
TARGET_ARCH=${3:-amd64}
PACKAGE_MODE=${4:-full}
case "$TARGET_OS/$TARGET_ARCH" in
  linux/amd64|linux/arm64|windows/386|windows/amd64) ;;
  *) echo "Supported Server targets: linux/amd64, linux/arm64, windows/386, windows/amd64." >&2; exit 2 ;;
esac
case "$PACKAGE_MODE" in
  full|server-only) ;;
  *) echo "Package mode must be full or server-only." >&2; exit 2 ;;
esac

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
: "${GOTOOLCHAIN:=go1.26.5}"
export GOTOOLCHAIN
. "$SCRIPT_DIR/lib/go-version.sh"
honeynet_require_go_1265
OUTPUT=$ROOT/dist
NAME=honeynet-server-$VERSION-$TARGET_OS-$TARGET_ARCH
if [ "$PACKAGE_MODE" = server-only ]; then
  NAME=honeynet-server-only-$VERSION-$TARGET_OS-$TARGET_ARCH
fi
STAGE=$OUTPUT/$NAME
LDFLAGS="-s -w -X main.version=$VERSION"

rm -rf "$STAGE"
mkdir -p "$STAGE/bin" "$STAGE/web" "$STAGE/configs" "$STAGE/scripts" "$STAGE/rules/builtin" "$STAGE/migrations/clickhouse"
[ -d "$ROOT/cve-rules-decrypted/Yara" ] || { echo "Missing cve-rules-decrypted/Yara" >&2; exit 1; }

if [ "$PACKAGE_MODE" = full ]; then
  mkdir -p "$STAGE/downloads" "$STAGE/templates/web"
  [ -f "$ROOT/honeypot-templates-server/services/config.json" ] || { echo "Missing honeypot-templates-server/services/config.json" >&2; exit 1; }
  command -v zip >/dev/null 2>&1 || { echo "zip is required to package the Web templates" >&2; exit 1; }
fi

cd "$ROOT/web"
npm_config_cache="$ROOT/.cache/npm" npm ci
npm run build

cd "$ROOT"
server_suffix=
if [ "$TARGET_OS" = windows ]; then server_suffix=.exe; fi
CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -trimpath -ldflags "$LDFLAGS" -o "$STAGE/bin/honeynet-server$server_suffix" ./cmd/server

build_agent() {
  agent_os=$1
  agent_arch=$2
  suffix=
  extra=
  if [ "$agent_os" = windows ]; then suffix=.exe; fi
  if [ "$agent_arch" = arm ]; then extra=7; fi
  output=$STAGE/downloads/honeynet-agent-$agent_os-$agent_arch$suffix
  if [ -n "$extra" ]; then
    CGO_ENABLED=0 GOOS=$agent_os GOARCH=$agent_arch GOARM=$extra go build -trimpath -ldflags "$LDFLAGS" -o "$output" ./cmd/agent
  else
    CGO_ENABLED=0 GOOS=$agent_os GOARCH=$agent_arch go build -trimpath -ldflags "$LDFLAGS" -o "$output" ./cmd/agent
  fi
  if [ "$agent_os" = linux ]; then
    guard_output=$STAGE/downloads/honeynet-agent-guard-$agent_os-$agent_arch
    if [ -n "$extra" ]; then
      CGO_ENABLED=0 GOOS=$agent_os GOARCH=$agent_arch GOARM=$extra go build -trimpath -ldflags "$LDFLAGS" -o "$guard_output" ./cmd/agent-guard
    else
      CGO_ENABLED=0 GOOS=$agent_os GOARCH=$agent_arch go build -trimpath -ldflags "$LDFLAGS" -o "$guard_output" ./cmd/agent-guard
    fi
  fi
}

if [ "$PACKAGE_MODE" = full ]; then
  CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -trimpath -ldflags "$LDFLAGS" -o "$STAGE/bin/honeynet-agent$server_suffix" ./cmd/agent
  if [ "$TARGET_OS" = linux ]; then
    CGO_ENABLED=0 GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -trimpath -ldflags "$LDFLAGS" -o "$STAGE/bin/honeynet-agent-guard" ./cmd/agent-guard
  fi
  build_agent linux 386
  build_agent linux amd64
  build_agent linux arm
  build_agent linux arm64
  build_agent linux loong64
  build_agent windows 386
  build_agent windows amd64
fi

cp -R "$ROOT/web/dist" "$STAGE/web/"
if [ "$PACKAGE_MODE" = full ]; then
  cp "$ROOT/LICENSE" "$STAGE/templates/web/"
  cp "$ROOT/THIRD_PARTY_NOTICES.md" "$STAGE/templates/web/"
  cp -R "$ROOT/honeypot-templates-server/services" "$STAGE/templates/web/"
  find "$STAGE/templates" -name .DS_Store -type f -delete
  (cd "$STAGE/templates/web" && tar $TAR_PORTABLE_FLAGS -czf "$STAGE/downloads/honeypot-templates-server.tar.gz" services LICENSE THIRD_PARTY_NOTICES.md)
  (cd "$STAGE/templates/web" && zip -qr "$STAGE/downloads/honeypot-templates-server.zip" services LICENSE THIRD_PARTY_NOTICES.md)
else
  : > "$STAGE/SERVER_ONLY"
fi
cp "$ROOT/configs/server.example.yaml" "$STAGE/configs/"
cp "$ROOT/configs/analytics.example.yaml" "$STAGE/configs/"
cp -R "$ROOT/migrations/clickhouse/." "$STAGE/migrations/clickhouse/"
cp -R "$ROOT/cve-rules-decrypted/Yara/." "$STAGE/rules/builtin/"
cp "$ROOT/README.md" "$STAGE/"
cp "$ROOT/LICENSE" "$STAGE/"
cp "$ROOT/THIRD_PARTY_NOTICES.md" "$STAGE/"
if [ -f "$ROOT/data/ipip.ipdb" ]; then
  mkdir -p "$STAGE/geoip"
  cp "$ROOT/data/ipip.ipdb" "$STAGE/geoip/ipip.ipdb"
fi

if [ "$TARGET_OS" = linux ]; then
  mkdir -p "$STAGE/deploy/systemd" "$STAGE/deploy/clickhouse/init" "$STAGE/scripts/lib"
  cp "$ROOT/scripts/install-server.sh" "$STAGE/scripts/"
  cp "$ROOT/scripts/lib/network.sh" "$STAGE/scripts/lib/"
  cp "$ROOT/deploy/systemd/honeynet-server.service" "$STAGE/deploy/systemd/"
  cp "$ROOT/deploy/clickhouse/compose.yaml" "$STAGE/deploy/clickhouse/"
  cp "$ROOT/deploy/clickhouse/init/bootstrap.sh" "$STAGE/deploy/clickhouse/init/"
  for clickhouse_script in install-clickhouse.sh migrate-clickhouse.sh smoke-clickhouse.sh uninstall-clickhouse.sh; do
    cp "$ROOT/scripts/$clickhouse_script" "$STAGE/scripts/"
    chmod 0755 "$STAGE/scripts/$clickhouse_script"
  done
  if [ "$PACKAGE_MODE" = full ]; then
    cp "$ROOT/deploy/systemd/honeynet-agent.service" "$STAGE/deploy/systemd/"
  fi
  chmod 0755 "$STAGE/scripts/install-server.sh"
else
  cp "$ROOT/scripts/install-server.ps1" "$STAGE/scripts/"
fi

# Remove inherited Finder metadata from the disposable staging tree. Without
# this step BSD tar can emit LIBARCHIVE.xattr PAX records even when portable
# archive flags are selected, causing noisy warnings on Linux extraction.
if command -v xattr >/dev/null 2>&1; then
  xattr -cr "$STAGE"
fi

cd "$OUTPUT"
if [ "$TARGET_OS" = windows ]; then
  command -v zip >/dev/null 2>&1 || { echo "zip is required to build a Windows release" >&2; exit 1; }
  archive=$NAME.zip
  zip -qr -FS "$archive" "$NAME"
else
  archive=$NAME.tar.gz
  tar $TAR_PORTABLE_FLAGS -czf "$archive" "$NAME"
fi
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "$archive" > "$archive.sha256"
else
  shasum -a 256 "$archive" > "$archive.sha256"
fi
echo "$OUTPUT/$archive"
