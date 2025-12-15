#!/bin/sh

TAG="portal-sync"
CTRL_URL="http://192.168.16.118:8443"
LEASE_FILE="/tmp/dhcp.leases"
IPSET_NAME="portal_allow"

SCANNED=0
ALLOWED=0
REFRESHED=0
REMOVED=0

log() {
	logger -t "$TAG" "$*"
}

# Ensure ipset exists (ipset v7 requires timeout value)
if ! ipset list "$IPSET_NAME" >/dev/null 2>&1; then
	ipset create "$IPSET_NAME" hash:mac timeout 0
	log "created ipset $IPSET_NAME"
fi

[ -f "$LEASE_FILE" ] || {
	log "no dhcp lease file"
	exit 0
}

# Read leases WITHOUT pipe (avoid subshell)
while read -r _ MAC _; do
	MAC="$(echo "$MAC" | tr 'A-F' 'a-f')"
	[ -z "$MAC" ] && continue

	SCANNED=$((SCANNED + 1))

	RESP="$(curl -s --max-time 2 "$CTRL_URL/portal/status/$MAC")"
	AUTH="$(echo "$RESP" | jq -r '.authorized // empty')"
	TTL="$(echo "$RESP" | jq -r '.ttl // 0')"

	if [ "$AUTH" != "true" ] || [ "$TTL" -le 0 ]; then
		if ipset test "$IPSET_NAME" "$MAC" 2>/dev/null; then
			ipset del "$IPSET_NAME" "$MAC"
			REMOVED=$((REMOVED + 1))
		fi
		continue
	fi

	ALLOWED=$((ALLOWED + 1))
	ipset -exist add "$IPSET_NAME" "$MAC" timeout "$TTL"
	REFRESHED=$((REFRESHED + 1))

done < "$LEASE_FILE"

log "scanned=$SCANNED allowed=$ALLOWED refreshed=$REFRESHED removed=$REMOVED"