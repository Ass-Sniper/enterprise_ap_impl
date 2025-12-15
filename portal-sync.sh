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

# No DHCP lease file
[ -f "$LEASE_FILE" ] || {
	log "no dhcp lease file"
	exit 0
}

# Iterate active DHCP leases
awk '{print tolower($2)}' "$LEASE_FILE" | while read -r MAC; do
	[ -z "$MAC" ] && continue
	SCANNED=$((SCANNED + 1))

	RESP="$(curl -s --max-time 2 "$CTRL_URL/portal/status/$MAC")"
	AUTH="$(echo "$RESP" | jsonfilter -e '@.authorized' 2>/dev/null)"
	TTL="$(echo "$RESP" | jsonfilter -e '@.ttl' 2>/dev/null)"

	# Not authorized or invalid TTL
	if [ "$AUTH" != "true" ] || [ -z "$TTL" ] || [ "$TTL" -le 0 ]; then
		if ipset test "$IPSET_NAME" "$MAC" 2>/dev/null; then
			ipset del "$IPSET_NAME" "$MAC"
			REMOVED=$((REMOVED + 1))
		fi
		continue
	fi

	# Authorized
	ALLOWED=$((ALLOWED + 1))
	ipset -exist add "$IPSET_NAME" "$MAC" timeout "$TTL"
	REFRESHED=$((REFRESHED + 1))
done

log "scanned=$SCANNED allowed=$ALLOWED refreshed=$REFRESHED removed=$REMOVED"