#!/bin/sh
# =========================================================
# ImmortalWRT Captive Portal restore script
#
# Runs on ImmortalWRT (data-plane).
# - Cleans up portal-related chains/ipsets first
# - Restores iptables (nat/filter) + ipset from backup
# - Restores /tmp/portal-runtime.env if present in backup
#
# Logs:
#   logread -e portal-restore
# =========================================================

set -eu

TAG="${TAG:-portal-restore}"
BACKUP_DIR="${BACKUP_DIR:-/tmp/portal-backup}"
RUNTIME_ENV="${RUNTIME_ENV:-/tmp/portal-runtime.env}"

# Portal chain/ipset naming conventions (keep consistent with portal-fw.sh)
CHAIN_DNS="${CHAIN_DNS:-PORTAL_DNS}"
CHAIN_FWD="${CHAIN_FWD:-PORTAL_FWD}"
IPSET_PREFIX="${IPSET_PREFIX:-portal_}"   # destroy ipsets starting with this prefix

log() { logger -t "$TAG" "$*"; }

log "event=restore_start dir=${BACKUP_DIR}"

[ -d "$BACKUP_DIR" ] || {
  log "event=restore_failed reason=backup_dir_missing"
  exit 1
}

# -----------------------
# Cleanup portal rules
# -----------------------
log "event=cleanup_start"

# Unhook chains if present
iptables -t nat    -D PREROUTING -j "$CHAIN_DNS" 2>/dev/null || true
iptables           -D forwarding_lan_rule -j "$CHAIN_FWD" 2>/dev/null || true

# Flush/delete chains
iptables -t nat    -F "$CHAIN_DNS" 2>/dev/null || true
iptables -t nat    -X "$CHAIN_DNS" 2>/dev/null || true
iptables           -F "$CHAIN_FWD" 2>/dev/null || true
iptables           -X "$CHAIN_FWD" 2>/dev/null || true

# Destroy portal ipsets (best-effort)
# ipset list -n prints set names
for s in $(ipset list -n 2>/dev/null | awk -v pfx="$IPSET_PREFIX" '$0 ~ "^"pfx {print}'); do
  ipset destroy "$s" 2>/dev/null || true
  log "event=ipset_destroyed name=${s}"
done

log "event=cleanup_done"

# -----------------------
# Restore iptables
# -----------------------
if [ -f "${BACKUP_DIR}/iptables.nat" ]; then
  iptables-restore -T nat < "${BACKUP_DIR}/iptables.nat" 2>/dev/null || {
    log "event=restore_failed target=iptables_nat"
    exit 2
  }
  log "event=iptables_nat_restored"
fi

if [ -f "${BACKUP_DIR}/iptables.filter" ]; then
  iptables-restore -T filter < "${BACKUP_DIR}/iptables.filter" 2>/dev/null || {
    log "event=restore_failed target=iptables_filter"
    exit 3
  }
  log "event=iptables_filter_restored"
fi

# -----------------------
# Restore ipset
# -----------------------
if [ -f "${BACKUP_DIR}/ipset.save" ]; then
  ipset restore < "${BACKUP_DIR}/ipset.save" 2>/dev/null || {
    log "event=restore_failed target=ipset"
    exit 4
  }
  log "event=ipset_restored"
fi

# -----------------------
# Restore runtime env (optional)
# -----------------------
if [ -f "${BACKUP_DIR}/portal-runtime.env" ]; then
  cp -f "${BACKUP_DIR}/portal-runtime.env" "$RUNTIME_ENV" 2>/dev/null || true
  log "event=runtime_env_restored path=${RUNTIME_ENV}"
fi

log "event=restore_done"
exit 0