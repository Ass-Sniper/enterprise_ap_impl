#!/bin/sh
# =========================================================
# ImmortalWRT Captive Portal data-plane restore script
#
# Purpose:
#   Restore iptables and ipset state from a previous backup,
#   and cleanly remove all Captive Portal related rules.
#
# Features:
#   - Idempotent (safe to run multiple times)
#   - Proper cleanup before restore
#   - Syslog integration (logread compatible)
# =========================================================

TAG="portal-restore"
BACKUP_DIR="/tmp/portal-backup"
IPSET_NAME="portal_allow"

log() {
    logger -t "$TAG" "$*"
}

log "event=restore_start dir=$BACKUP_DIR"

# ---------------------------------------------------------
# 0. Validate backup directory
# ---------------------------------------------------------
if [ ! -d "$BACKUP_DIR" ]; then
    log "event=restore_abort reason=no_backup_dir"
    exit 1
fi

# ---------------------------------------------------------
# 1. Cleanup Captive Portal data-plane
#    - Remove custom iptables chains
#    - Destroy portal ipset if exists
# ---------------------------------------------------------
log "event=cleanup_start"

# NAT table cleanup (DNS hijack chain)
iptables -t nat -D PREROUTING -j PORTAL_DNS 2>/dev/null
iptables -t nat -F PORTAL_DNS 2>/dev/null
iptables -t nat -X PORTAL_DNS 2>/dev/null

# FILTER table cleanup (forward control chain)
iptables -D forwarding_lan_rule -j PORTAL_FWD 2>/dev/null
iptables -F PORTAL_FWD 2>/dev/null
iptables -X PORTAL_FWD 2>/dev/null

# ipset cleanup
if ipset list "$IPSET_NAME" >/dev/null 2>&1; then
    ipset destroy "$IPSET_NAME"
    log "event=ipset_destroyed name=$IPSET_NAME"
fi

log "event=cleanup_done"

# ---------------------------------------------------------
# 2. Restore iptables state
# ---------------------------------------------------------
if [ -f "$BACKUP_DIR/iptables.nat" ]; then
    iptables-restore < "$BACKUP_DIR/iptables.nat"
    log "event=iptables_nat_restored"
fi

if [ -f "$BACKUP_DIR/iptables.filter" ]; then
    iptables-restore < "$BACKUP_DIR/iptables.filter"
    log "event=iptables_filter_restored"
fi

# ---------------------------------------------------------
# 3. Restore ipset state
# ---------------------------------------------------------
if [ -f "$BACKUP_DIR/ipset.save" ]; then
    ipset restore < "$BACKUP_DIR/ipset.save"
    log "event=ipset_restored"
fi

# ---------------------------------------------------------
# 4. Done
# ---------------------------------------------------------
log "event=restore_done"
exit 0