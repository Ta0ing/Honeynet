#!/bin/sh

honeynet_require_go_1265() {
  command -v go >/dev/null 2>&1 || { echo "Go 1.26.5 or newer is required to build Honeynet release artifacts." >&2; return 1; }
  version=$(go env GOVERSION 2>/dev/null || true)
  case "$version" in
    go*) version=${version#go} ;;
    *) echo "Unable to determine Go toolchain version." >&2; return 1 ;;
  esac
  major=${version%%.*}
  rest=${version#*.}
  minor=${rest%%.*}
  patch=${rest#*.}
  patch=${patch%%[^0-9]*}
  case "$major.$minor.$patch" in
    *[!0-9.]*) echo "Unsupported Go toolchain version: go$version" >&2; return 1 ;;
  esac
  if [ "$major" -lt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -lt 26 ]; } || { [ "$major" -eq 1 ] && [ "$minor" -eq 26 ] && [ "${patch:-0}" -lt 5 ]; }; then
    echo "Go 1.26.5 or newer is required; found go$version." >&2
    return 1
  fi
}
