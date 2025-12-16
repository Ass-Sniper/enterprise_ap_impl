#!/bin/sh
# =========================================================
# ImmortalWRT Captive Portal MAC sync (production-oriented)
#
# Features:
#   1) DHCP + ARP/NDP + FDB multi-source MAC discovery
#      - output: "mac source" with priority arp > dhcp > fdb
#   2) Minimum survival threshold:
#      - remove from ipset only after a MAC disappears for N consecutive runs
#   3) Source-based TTL strategy:
#      - ipset timeout = min(controller_session_ttl, source_cap)
#   4) Controller heartbeat integration:
#      - POST /portal/heartbeat first, fallback to GET /portal/status/{mac}
#
# Logs:
#   logread -e portal-sync
# =========================================================

# -----------------------------
# Config
# -----------------------------
TAG="portal-sync"

LAN_IF="br-lan"
IPSET_NAME="portal_allow"

CTRL_HOST="192.168.16.118"
CTRL_PORT="8443"
CTRL_URL="http://${CTRL_HOST}:${CTRL_PORT}"

CURL_TIMEOUT=2
CURL_RETRY=1

# Minimum survival threshold: consecutive missing runs to remove from ipset
MISS_THRESHOLD="${MISS_THRESHOLD:-3}"

# Source-based TTL caps (seconds)
TTL_CAP_DHCP="${TTL_CAP_DHCP:-3600}"
TTL_CAP_ARP="${TTL_CAP_ARP:-1800}"
TTL_CAP_FDB="${TTL_CAP_FDB:-600}"

# Optional: per-MAC verbose logging (0/1)
VERBOSE="${VERBOSE:-0}"

# State file (miss counters)
STATE_FILE="/tmp/portal-sync.state"

# Temp files
TMP_PREFIX="/tmp/portal-sync.$$"
TMP_RAW="${TMP_PREFIX}.raw"       # raw "mac source" lines (may contain duplicates)
TMP_CUR="${TMP_PREFIX}.cur"       # fused "mac best_source" lines
TMP_NEWSTATE="${TMP_PREFIX}.state"
TMP_RMLIST="${TMP_PREFIX}.rm"
TMP_RESP="${TMP_PREFIX}.resp"

# -----------------------------
# Counters
# -----------------------------
ACTIVE=0
SCANNED=0
AUTHORIZED=0
REFRESHED=0
REMOVED=0
MISS_REMOVED=0
ERRORS=0
HB_USED=0
STATUS_FALLBACK=0

# -----------------------------
# Helpers
# -----------------------------
log() {
    logger -t "$TAG" "$*"
}

vlog() {
    [ "$VERBOSE" = "1" ] && logger -t "$TAG" "$*"
}

have_cmd() {
    command -v "$1" >/dev/null 2>&1
}

now_epoch() {
    date +%s 2>/dev/null || echo 0
}

normalize_mac() {
    echo "$1" | tr 'A-F' 'a-f' \
      | sed -n 's/^\(..:..:..:..:..:..\).*$/\1/p'
}

is_mac() {
    echo "$1" | grep -Eq '^[0-9a-f]{2}(:[0-9a-f]{2}){5}$'
}

min2() {
    a="$1"; b="$2"
    [ -z "$a" ] && echo "$b" && return
    [ -z "$b" ] && echo "$a" && return
    [ "$a" -le "$b" ] 2>/dev/null && echo "$a" || echo "$b"
}

ttl_cap_for_source() {
    case "$1" in
        dhcp) echo "$TTL_CAP_DHCP" ;;
        arp)  echo "$TTL_CAP_ARP" ;;
        fdb)  echo "$TTL_CAP_FDB" ;;
        *)    echo "$TTL_CAP_ARP" ;;
    esac
}

json_get() {
    # Prefer jsonfilter (OpenWrt), fallback to jq if available
    body="$1"
    expr="$2"
    if have_cmd jsonfilter; then
        echo "$body" | jsonfilter -e "$expr" 2>/dev/null
    elif have_cmd jq; then
        key="$(echo "$expr" | sed 's/^@/./')"
        echo "$body" | jq -r "$key" 2>/dev/null
    else
        echo ""
    fi
}

ensure_ipset() {
    if ! ipset list "$IPSET_NAME" >/dev/null 2>&1; then
        ipset create "$IPSET_NAME" hash:mac
        log "event=ipset_create name=$IPSET_NAME"
    fi
}

cleanup() {
    rm -f "$TMP_RAW" "$TMP_CUR" "$TMP_NEWSTATE" "$TMP_RMLIST" "$TMP_RESP" 2>/dev/null
}

trap cleanup EXIT

# -----------------------------
# MAC discovery backends
# -----------------------------
discover_macs_dhcp() {
    [ -f /tmp/dhcp.leases ] || return 0
    awk '{print tolower($2)}' /tmp/dhcp.leases 2>/dev/null \
      | while read -r mac; do
            mac="$(normalize_mac "$mac")"
            is_mac "$mac" || continue
            echo "$mac dhcp"
        done
}

discover_macs_arp() {
    # ip neigh output varies; only accept lines with lladdr and valid mac
    ip neigh show dev "$LAN_IF" 2>/dev/null \
      | awk '
          $2=="lladdr" {
            mac=tolower($3)
            if (mac ~ /^[0-9a-f]{2}(:[0-9a-f]{2}){5}$/)
              print mac
          }
        ' \
      | while read -r mac; do
            mac="$(normalize_mac "$mac")"
            is_mac "$mac" || continue
            echo "$mac arp"
        done
}

discover_macs_fdb() {
    # Prefer "bridge fdb", fallback to brctl
    if have_cmd bridge; then
        bridge fdb show br "$LAN_IF" 2>/dev/null \
          | awk '
              !/ self / && !/ permanent / {
                mac=tolower($1)
                if (mac ~ /^[0-9a-f]{2}(:[0-9a-f]{2}){5}$/)
                  print mac
              }
            ' \
          | while read -r mac; do
                mac="$(normalize_mac "$mac")"
                is_mac "$mac" || continue
                echo "$mac fdb"
            done
    elif have_cmd brctl; then
        brctl showmacs "$LAN_IF" 2>/dev/null \
          | awk '
              NR>1 {
                mac=tolower($2)
                if (mac ~ /^[0-9a-f]{2}(:[0-9a-f]{2}){5}$/)
                  print mac
              }
            ' \
          | while read -r mac; do
                mac="$(normalize_mac "$mac")"
                is_mac "$mac" || continue
                echo "$mac fdb"
            done
    fi
}

# -----------------------------
# Multi-source fusion (arp > dhcp > fdb)
# Output: TMP_CUR lines: "mac best_source"
# -----------------------------
collect_active_macs() {
    : > "$TMP_RAW"
    : > "$TMP_CUR"

    SELF_MAC="$(cat /sys/class/net/$LAN_IF/address 2>/dev/null | tr 'A-F' 'a-f')"

    # Collect raw
    discover_macs_dhcp >> "$TMP_RAW"
    discover_macs_arp  >> "$TMP_RAW"
    discover_macs_fdb  >> "$TMP_RAW"

    # Filter self MAC and fuse by priority
    # priority: arp=3, dhcp=2, fdb=1
    awk -v self="$SELF_MAC" '
      function prio(s) {
        if (s=="arp")  return 3
        if (s=="dhcp") return 2
        if (s=="fdb")  return 1
        return 0
      }
      {
        mac=$1; src=$2
        if (mac=="" || src=="") next
        if (mac==self) next
        if (!(mac in best) || prio(src) > prio(best[mac])) best[mac]=src
      }
      END {
        for (m in best) print m, best[m]
      }
    ' "$TMP_RAW" | sort -u > "$TMP_CUR"

    ACTIVE="$(wc -l < "$TMP_CUR" 2>/dev/null | tr -d ' ')"
}

# -----------------------------
# Controller calls
# -----------------------------
ctrl_post_json() {
    path="$1"
    json="$2"

    # write body to TMP_RESP, return http_code in stdout
    code="$(curl -sS \
        --max-time "$CURL_TIMEOUT" \
        --retry "$CURL_RETRY" \
        -H 'Content-Type: application/json' \
        -d "$json" \
        -o "$TMP_RESP" \
        -w '%{http_code}' \
        "$CTRL_URL$path" 2>/dev/null)"

    echo "$code"
}

ctrl_get() {
    path="$1"
    code="$(curl -sS \
        --max-time "$CURL_TIMEOUT" \
        --retry "$CURL_RETRY" \
        -o "$TMP_RESP" \
        -w '%{http_code}' \
        "$CTRL_URL$path" 2>/dev/null)"

    echo "$code"
}

query_session() {
    mac="$1"
    src="$2"

    # Try heartbeat first
    hb_code="$(ctrl_post_json "/portal/heartbeat" "{\"mac\":\"$mac\",\"source\":\"$src\"}")"
    if [ "$hb_code" = "200" ]; then
        HB_USED=$((HB_USED + 1))
        cat "$TMP_RESP"
        return 0
    fi

    # If heartbeat not found or failed, fallback to status
    st_code="$(ctrl_get "/portal/status/$mac")"
    if [ "$st_code" = "200" ]; then
        STATUS_FALLBACK=$((STATUS_FALLBACK + 1))
        cat "$TMP_RESP"
        return 0
    fi

    # Failure
    ERRORS=$((ERRORS + 1))
    return 1
}

# -----------------------------
# State update: miss counters & removal list
# -----------------------------
update_state_and_build_removals() {
    now="$(now_epoch)"
    : > "$TMP_RMLIST"

    # Old state format per line:
    #   mac miss last_src last_seen
    # Example:
    #   70:4d:... 0 arp 1700000000
    [ -f "$STATE_FILE" ] || : > "$STATE_FILE"

    awk -v now="$now" -v thr="$MISS_THRESHOLD" '
      # Read current list into cur[mac]=src
      FNR==NR {
        if ($1!="" && $2!="") cur[$1]=$2
        next
      }
      # Read old state
      {
        mac=$1; miss=$2; last_src=$3; last_seen=$4
        if (mac=="") next
        old_miss[mac]=miss
        old_src[mac]=last_src
        old_seen[mac]=last_seen
      }
      END {
        # For MACs in current list: miss=0, update src, seen
        for (m in cur) {
          print m, 0, cur[m], now
          touched[m]=1
        }

        # For MACs not seen now: miss++
        for (m in old_miss) {
          if (touched[m]) continue
          miss=old_miss[m]+0
          miss=miss+1
          src=old_src[m]
          seen=old_seen[m]
          print m, miss, src, seen
          if (miss >= thr) {
            # mark for removal
            print m > "'"$TMP_RMLIST"'"
          }
        }
      }
    ' "$TMP_CUR" "$STATE_FILE" > "$TMP_NEWSTATE"

    mv "$TMP_NEWSTATE" "$STATE_FILE"
}

# -----------------------------
# Apply ipset changes for currently active MACs
# -----------------------------
sync_active_macs() {
    [ -s "$TMP_CUR" ] || return 0

    while read -r mac src; do
        SCANNED=$((SCANNED + 1))

        # Query controller (heartbeat preferred)
        body="$(query_session "$mac" "$src")" || continue

        auth="$(json_get "$body" '@.authorized')"
        ttl="$(json_get "$body" '@.ttl')"
        role="$(json_get "$body" '@.role')"

        # Unauthorized or invalid TTL => remove immediately (not waiting for miss threshold)
        if [ "$auth" != "true" ] || [ -z "$ttl" ] || [ "$ttl" -le 0 ] 2>/dev/null; then
            if ipset test "$IPSET_NAME" "$mac" >/dev/null 2>&1; then
                ipset del "$IPSET_NAME" "$mac" >/dev/null 2>&1
                REMOVED=$((REMOVED + 1))
                vlog "mac=$mac source=$src action=del reason=unauthorized role=${role:-na}"
            else
                vlog "mac=$mac source=$src action=skip reason=unauthorized role=${role:-na}"
            fi
            continue
        fi

        AUTHORIZED=$((AUTHORIZED + 1))

        cap="$(ttl_cap_for_source "$src")"
        eff_ttl="$(min2 "$ttl" "$cap")"

        # Final guard
        if [ -z "$eff_ttl" ] || [ "$eff_ttl" -le 0 ] 2>/dev/null; then
            ERRORS=$((ERRORS + 1))
            vlog "mac=$mac source=$src action=skip reason=bad_ttl ttl=${ttl:-na} cap=${cap:-na}"
            continue
        fi

        if ipset add "$IPSET_NAME" "$mac" timeout "$eff_ttl" -exist >/dev/null 2>&1; then
            REFRESHED=$((REFRESHED + 1))
            vlog "mac=$mac source=$src action=refresh role=${role:-na} ttl=$ttl cap=$cap eff_ttl=$eff_ttl"
        else
            ERRORS=$((ERRORS + 1))
            vlog "mac=$mac source=$src action=error reason=ipset_add_failed"
        fi
    done < "$TMP_CUR"
}

# -----------------------------
# Remove MACs that disappeared for N consecutive runs
# -----------------------------
apply_miss_based_removals() {
    [ -s "$TMP_RMLIST" ] || return 0
    while read -r mac; do
        [ -z "$mac" ] && continue
        if ipset test "$IPSET_NAME" "$mac" >/dev/null 2>&1; then
            ipset del "$IPSET_NAME" "$mac" >/dev/null 2>&1
            MISS_REMOVED=$((MISS_REMOVED + 1))
            vlog "mac=$mac source=state action=del reason=miss_threshold"
        fi
    done < "$TMP_RMLIST"

    # Optional: prune state entries that already exceeded threshold to keep file small
    # (kept simple & safe)
    awk -v thr="$MISS_THRESHOLD" '
      { if ($2+0 < thr) print }
    ' "$STATE_FILE" > "${STATE_FILE}.tmp" && mv "${STATE_FILE}.tmp" "$STATE_FILE"
}

# -----------------------------
# Main
# -----------------------------
main() {
    ensure_ipset
    collect_active_macs
    update_state_and_build_removals

    if [ "$ACTIVE" -le 0 ]; then
        apply_miss_based_removals
        log "event=sync_done active=0 scanned=0 authorized=0 refreshed=0 removed=$REMOVED miss_removed=$MISS_REMOVED errors=$ERRORS hb_used=$HB_USED status_fallback=$STATUS_FALLBACK thr=$MISS_THRESHOLD ipset=$IPSET_NAME ctrl=$CTRL_HOST:$CTRL_PORT"
        exit 0
    fi

    sync_active_macs
    apply_miss_based_removals

    log "event=sync_done active=$ACTIVE scanned=$SCANNED authorized=$AUTHORIZED refreshed=$REFRESHED removed=$REMOVED miss_removed=$MISS_REMOVED errors=$ERRORS hb_used=$HB_USED status_fallback=$STATUS_FALLBACK thr=$MISS_THRESHOLD ttlcap_dhcp=$TTL_CAP_DHCP ttlcap_arp=$TTL_CAP_ARP ttlcap_fdb=$TTL_CAP_FDB ipset=$IPSET_NAME ctrl=$CTRL_HOST:$CTRL_PORT"
}

main "$@"