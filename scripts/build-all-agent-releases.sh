#!/bin/sh
set -eu

VERSION=${1:-0.24.0-dev}
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

for target in linux/386 linux/amd64 linux/arm linux/arm64 linux/loong64 windows/386 windows/amd64; do
  target_os=${target%/*}
  target_arch=${target#*/}
  "$SCRIPT_DIR/build-agent-release.sh" "$VERSION" "$target_os" "$target_arch"
done
