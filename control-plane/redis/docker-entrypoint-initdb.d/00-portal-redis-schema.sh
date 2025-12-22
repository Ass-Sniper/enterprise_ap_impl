#!/bin/sh
set -e

echo "[portal-init] Initializing Redis keys..."

# ===============================
# Feature Flags
# ===============================
redis-cli SET feature:portal 1
redis-cli SET feature:portal:redirect 1
redis-cli SET feature:portal:walled_garden 1

redis-cli SET feature:auth:pap 1
redis-cli SET feature:auth:radius 1
redis-cli SET feature:auth:mac 0

# ===============================
# Auth Strategy
# ===============================
redis-cli SET auth:strategy:pap enabled
redis-cli SET auth:strategy:radius enabled
redis-cli SET auth:strategy:mac disabled

# ===============================
# Portal Security
# ===============================
redis-cli SET portal:hmac:key "CHANGE_ME_IN_PROD"

# ===============================
# Policy
# ===============================
redis-cli SET policy:global:rev 1

echo "[portal-init] Redis initialization done."
