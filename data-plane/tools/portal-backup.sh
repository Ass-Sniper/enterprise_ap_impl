#!/bin/sh
# =========================================================
# ImmortalWRT Captive Portal backup script
#
# Runs on ImmortalWRT (data-plane).
# - Backs up iptables (nat/filter) and ipset state
# - Also backs up portal runtime env (/tmp/portal-runtime.env) if present
#
# Logs:
#   logread -e portal-backup
# =========================================================

set -eu

BACKUP_DIR="${BACKUP_DIR:-/tmp/portal-backup}"
TAG="${TAG:-portal-backup}"
RUNTIME_ENV="${RUNTIME_ENV:-/tmp/portal-runtime.env}"

log() { logger -t "$TAG" "$*"; }

log "event=backup_start dir=${BACKUP_DIR}"

mkdir -p "$BACKUP_DIR" || {
  log "event=backup_failed reason=mkdir"
  exit 1
}

# Save iptables tables
iptables-save -t nat    > "${BACKUP_DIR}/iptables.nat"    2>/dev/null || true
iptables-save -t filter > "${BACKUP_DIR}/iptables.filter" 2>/dev/null || true

# Save ipset
ipset save > "${BACKUP_DIR}/ipset.save" 2>/dev/null || true

# Save runtime env (optional)
if [ -f "$RUNTIME_ENV" ]; then
  cp -f "$RUNTIME_ENV" "${BACKUP_DIR}/portal-runtime.env" 2>/dev/null || true
fi

# Meta info
HOSTNAME="$(cat /proc/sys/kernel/hostname 2>/dev/null || echo unknown)"
KERNEL="$(uname -r 2>/dev/null || echo unknown)"
TS="$(date -Iseconds 2>/dev/null || date)"

POLICY_VERSION=""
CTRL_BASE=""
if [ -f "$RUNTIME_ENV" ]; then
  # shellcheck disable=SC1090
  . "$RUNTIME_ENV" 2>/dev/null || true
  POLICY_VERSION="${POLICY_VERSION:-}"
  CTRL_BASE="${CTRL_BASE:-}"
fi

{
  echo "time=${TS}"
  echo "hostname=${HOSTNAME}"
  echo "kernel=${KERNEL}"
  echo "policy_version=${POLICY_VERSION}"
  echo "ctrl_base=${CTRL_BASE}"
} > "${BACKUP_DIR}/meta.info" 2>/dev/null || true

log "event=backup_done"
exit 0