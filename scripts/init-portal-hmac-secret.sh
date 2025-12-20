#!/usr/bin/env bash
#
# init-portal-hmac-secret.sh
#
# Initialize Portal HMAC secret on host
# - For Docker (non-swarm) secret file mount
# - Compatible with future JWT / mTLS secret layout
#
# Author: you
# Usage:
#   sudo ./init-portal-hmac-secret.sh
#   sudo ./init-portal-hmac-secret.sh rotate
#

set -euo pipefail

# =========================
# Configurable parameters
# =========================
SECRET_ROOT="/opt/portal/secrets"
SECRET_NAME="portal_hmac"
CURRENT_KID="v1"
KEY_BYTES=32

# =========================
# Helper functions
# =========================
log() {
    echo "[INFO] $*"
}

fatal() {
    echo "[ERROR] $*" >&2
    exit 1
}

require_root() {
    if [ "$(id -u)" -ne 0 ]; then
        fatal "This script must be run as root"
    fi
}

gen_secret() {
    openssl rand -base64 "${KEY_BYTES}"
}

secret_path() {
    echo "${SECRET_ROOT}/${SECRET_NAME}_${CURRENT_KID}"
}

# =========================
# Main
# =========================
require_root

ACTION="${1:-init}"

log "Action: ${ACTION}"
log "Secret root: ${SECRET_ROOT}"
log "Current KID: ${CURRENT_KID}"

mkdir -p "${SECRET_ROOT}"
chmod 700 "${SECRET_ROOT}"

case "${ACTION}" in
    init)
        if [ -f "$(secret_path)" ]; then
            fatal "Secret already exists: $(secret_path)"
        fi

        log "Generating new HMAC secret (${KEY_BYTES} bytes)..."
        gen_secret | tr -d '\n' > "$(secret_path)"

        chmod 600 "$(secret_path)"

        log "Secret created:"
        log "  $(secret_path)"
        ;;
    rotate)
        NEXT_KID="v$(($(echo "${CURRENT_KID}" | tr -d 'v') + 1))"
        NEXT_PATH="${SECRET_ROOT}/${SECRET_NAME}_${NEXT_KID}"

        log "Rotating key: ${CURRENT_KID} -> ${NEXT_KID}"

        if [ -f "${NEXT_PATH}" ]; then
            fatal "Next secret already exists: ${NEXT_PATH}"
        fi

        gen_secret | tr -d '\n' > "${NEXT_PATH}"
        chmod 600 "${NEXT_PATH}"

        log "New secret created:"
        log "  ${NEXT_PATH}"
        log "⚠️  Remember to update PORTAL_HMAC_CURRENT_KID=${NEXT_KID} in container env"
        ;;
    *)
        fatal "Unknown action: ${ACTION} (use: init | rotate)"
        ;;
esac

log "Done."

