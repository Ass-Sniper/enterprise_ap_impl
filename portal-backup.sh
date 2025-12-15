#!/bin/sh
# =========================================================
# Backup current iptables & ipset state before portal test
# =========================================================

BACKUP_DIR="/tmp/portal-backup"
TAG="portal-backup"

log() {
	logger -t "$TAG" "$*"
}

log "event=backup_start dir=$BACKUP_DIR"

mkdir -p "$BACKUP_DIR" || {
	log "event=backup_failed reason=mkdir"
	exit 1
}

# Backup iptables
iptables-save -t filter > "$BACKUP_DIR/iptables.filter" || exit 1
iptables-save -t nat	> "$BACKUP_DIR/iptables.nat"	|| exit 1

# Backup ipset
ipset save > "$BACKUP_DIR/ipset.save" || exit 1

# Meta info
cat > "$BACKUP_DIR/meta.info" <<EOF
time=$(date -Iseconds)
hostname=$(cat /proc/sys/kernel/hostname 2>/dev/null || echo unknown)
kernel=$(uname -r)
EOF

log "event=backup_done"