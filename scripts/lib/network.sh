#!/bin/sh
# POSIX address helpers shared by native deployment scripts. Callers retain
# validation policy; these functions only parse/format unambiguous endpoints.

honeynet_listen_port() {
  honeynet_value=$1
  case "$honeynet_value" in
    \[*\]:*) honeynet_port=${honeynet_value##*:} ;;
    :*) honeynet_port=${honeynet_value#:} ;;
    *:*)
      honeynet_host_part=${honeynet_value%:*}
      case "$honeynet_host_part" in *:*) return 1 ;; esac
      honeynet_port=${honeynet_value##*:}
      ;;
    *) return 1 ;;
  esac
  [ -n "$honeynet_port" ] || return 1
  printf '%s\n' "$honeynet_port"
}

honeynet_listen_host() {
  honeynet_value=$1
  case "$honeynet_value" in
    \[*\]:*)
      honeynet_host=${honeynet_value#\[}
      honeynet_host=${honeynet_host%%\]*}
      ;;
    :*) honeynet_host= ;;
    *:*)
      honeynet_host=${honeynet_value%:*}
      case "$honeynet_host" in *:*) return 1 ;; esac
      ;;
    *) return 1 ;;
  esac
  printf '%s\n' "$honeynet_host"
}

honeynet_url_host() {
  honeynet_value=$1
  case "$honeynet_value" in *://*) ;; *) return 1 ;; esac
  honeynet_authority=${honeynet_value#*://}
  honeynet_authority=${honeynet_authority%%/*}
  case "$honeynet_authority" in *@*) return 1 ;; esac
  case "$honeynet_authority" in
    \[*\]*)
      honeynet_host=${honeynet_authority#\[}
      honeynet_host=${honeynet_host%%\]*}
      ;;
    *:*)
      honeynet_host=${honeynet_authority%%:*}
      honeynet_remainder=${honeynet_authority#*:}
      case "$honeynet_remainder" in *:*) return 1 ;; esac
      ;;
    *) honeynet_host=$honeynet_authority ;;
  esac
  [ -n "$honeynet_host" ] || return 1
  printf '%s\n' "$honeynet_host"
}

honeynet_format_url_host() {
  honeynet_value=$1
  case "$honeynet_value" in
    \[*\]) printf '%s\n' "$honeynet_value" ;;
    *:*) printf '[%s]\n' "$honeynet_value" ;;
    *) printf '%s\n' "$honeynet_value" ;;
  esac
}

honeynet_probe_host() {
  case "$1" in
    ''|0.0.0.0) printf '%s\n' '127.0.0.1' ;;
    ::) printf '%s\n' '::1' ;;
    *) printf '%s\n' "$1" ;;
  esac
}

honeynet_detect_host_ip() {
  honeynet_ipv6=
  for honeynet_candidate in $(hostname -I 2>/dev/null || true); do
    case "$honeynet_candidate" in
      *:*) [ -n "$honeynet_ipv6" ] || honeynet_ipv6=$honeynet_candidate ;;
      *) printf '%s\n' "$honeynet_candidate"; return 0 ;;
    esac
  done
  printf '%s\n' "${honeynet_ipv6:-127.0.0.1}"
}
