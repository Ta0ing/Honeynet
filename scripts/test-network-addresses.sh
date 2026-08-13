#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
. "$SCRIPT_DIR/lib/network.sh"

assert_equal() {
  if [ "$1" != "$2" ]; then
    echo "address helper mismatch: got '$1', want '$2'" >&2
    exit 1
  fi
}

assert_equal "$(honeynet_listen_host ':8080')" ""
assert_equal "$(honeynet_listen_port ':8080')" "8080"
assert_equal "$(honeynet_listen_host '[::]:8080')" "::"
assert_equal "$(honeynet_listen_port '[::]:8080')" "8080"
assert_equal "$(honeynet_listen_host '[2001:db8::10]:8443')" "2001:db8::10"
assert_equal "$(honeynet_listen_port '[2001:db8::10]:8443')" "8443"
assert_equal "$(honeynet_url_host 'https://[2001:db8::10]:8443')" "2001:db8::10"
assert_equal "$(honeynet_url_host 'http://192.0.2.10:8080')" "192.0.2.10"
assert_equal "$(honeynet_format_url_host '2001:db8::10')" "[2001:db8::10]"
assert_equal "$(honeynet_probe_host '::')" "::1"

if honeynet_listen_port '2001:db8::10:8443' >/dev/null 2>&1; then
  echo "unbracketed IPv6 listener was accepted" >&2
  exit 1
fi
if honeynet_url_host 'https://2001:db8::10:8443' >/dev/null 2>&1; then
  echo "unbracketed IPv6 URL was accepted" >&2
  exit 1
fi

echo "IPv4/IPv6 deployment address helpers passed"
