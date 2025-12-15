#!/bin/sh
# =========================================================
# ImmortalWRT Captive Portal data-plane init script
# with syslog (logread) support
# =========================================================

PORTAL_IP="192.168.16.118"
LAN_IF="br-lan"
IPSET_NAME="portal_allow"
LOG_TAG="portal-fw"

log() {
	logger -t "$LOG_TAG" "$1"
}

log "[INIT] Initializing Captive Portal data plane"

# ---------------------------------------------------------
# 1. ipset: authorized MAC list
# ---------------------------------------------------------
if ! ipset list "$IPSET_NAME" >/dev/null 2>&1; then
	log "[IPSET] Creating ipset $IPSET_NAME (hash:mac, timeout=3600)"
	ipset create "$IPSET_NAME" hash:mac timeout 3600
else
	log "[IPSET] ipset $IPSET_NAME already exists"
fi

# ---------------------------------------------------------
# 2. NAT table: DNS hijack chain
# ---------------------------------------------------------
if ! iptables -t nat -L PORTAL_DNS >/dev/null 2>&1; then
	log "[NAT] Creating chain PORTAL_DNS"
	iptables -t nat -N PORTAL_DNS
else
	log "[NAT] Chain PORTAL_DNS already exists"
fi

# hook PREROUTING -> PORTAL_DNS
if ! iptables -t nat -C PREROUTING -i "$LAN_IF" -j PORTAL_DNS >/dev/null 2>&1; then
	log "[NAT] Hooking PORTAL_DNS into PREROUTING ($LAN_IF)"
	iptables -t nat -I PREROUTING 1 -i "$LAN_IF" -j PORTAL_DNS
else
	log "[NAT] PORTAL_DNS already hooked in PREROUTING"
fi

# rebuild PORTAL_DNS
log "[NAT] Rebuilding PORTAL_DNS rules"
iptables -t nat -F PORTAL_DNS

iptables -t nat -A PORTAL_DNS \
	-m set --match-set "$IPSET_NAME" src \
	-j RETURN

iptables -t nat -A PORTAL_DNS \
	-p udp --dport 53 \
	-j DNAT --to-destination "$PORTAL_IP:53"

iptables -t nat -A PORTAL_DNS \
	-p tcp --dport 53 \
	-j DNAT --to-destination "$PORTAL_IP:53"

iptables -t nat -A PORTAL_DNS -j RETURN

# ---------------------------------------------------------
# 3. FILTER table: forwarding control chain
# ---------------------------------------------------------
if ! iptables -L PORTAL_FWD >/dev/null 2>&1; then
	log "[FWD] Creating chain PORTAL_FWD"
	iptables -N PORTAL_FWD
else
	log "[FWD] Chain PORTAL_FWD already exists"
fi

# hook forwarding_lan_rule -> PORTAL_FWD
if ! iptables -C forwarding_lan_rule -j PORTAL_FWD >/dev/null 2>&1; then
	log "[FWD] Hooking PORTAL_FWD into forwarding_lan_rule"
	iptables -I forwarding_lan_rule 1 -j PORTAL_FWD
else
	log "[FWD] PORTAL_FWD already hooked"
fi

# rebuild PORTAL_FWD
log "[FWD] Rebuilding PORTAL_FWD rules"
iptables -F PORTAL_FWD

iptables -A PORTAL_FWD \
	-m set --match-set "$IPSET_NAME" src \
	-j ACCEPT

iptables -A PORTAL_FWD \
	-d "$PORTAL_IP" \
	-j ACCEPT

iptables -A PORTAL_FWD \
	-j REJECT --reject-with icmp-port-unreachable

log "[DONE] Captive Portal data plane initialized successfully"
