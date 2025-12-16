#!/bin/sh
# ==========================================================
# refactor_ap_controller_to_go.sh
#
# Purpose:
#   Non-destructively introduce a Go-based AP Controller
#   alongside the existing Python implementation.
#
# Design goals:
#   - Keep Python controller intact (rollback-friendly)
#   - Reuse controller.yaml (single source of truth)
#   - Prepare for production-grade Go refactor
#   - Zero impact on data-plane (iptables/ipset)
#
# ==========================================================

set -e

ROOT="$(pwd)"
CP_DIR="$ROOT/control-plane"
GO_DIR="$CP_DIR/ap-controller-go"

echo "[INFO] Starting AP Controller Go refactor (non-destructive)"

# ----------------------------------------------------------
# 0. Sanity checks
# ----------------------------------------------------------
if [ ! -d "$CP_DIR" ]; then
  echo "[ERROR] control-plane directory not found"
  exit 1
fi

if [ -d "$GO_DIR" ]; then
  echo "[ERROR] ap-controller-go already exists, aborting"
  exit 1
fi

# ----------------------------------------------------------
# 1. Create Go controller directory layout
# ----------------------------------------------------------
echo "[INFO] Creating ap-controller-go directory structure"

mkdir -p "$GO_DIR"/cmd/ap-controller
mkdir -p "$GO_DIR"/internal/audit
mkdir -p "$GO_DIR"/internal/config
mkdir -p "$GO_DIR"/internal/roles
mkdir -p "$GO_DIR"/internal/store
mkdir -p "$GO_DIR"/internal/http
mkdir -p "$GO_DIR"/config

# ----------------------------------------------------------
# 2. Copy shared controller.yaml
# ----------------------------------------------------------
echo "[INFO] Copying controller.yaml"

if [ -f "$CP_DIR/config/controller.yaml" ]; then
  cp "$CP_DIR/config/controller.yaml" "$GO_DIR/config/controller.yaml"
else
  echo "[WARN] controller.yaml not found, creating placeholder"
  cat > "$GO_DIR/config/controller.yaml" <<'EOF'
# controller.yaml placeholder
# Copy your real controller.yaml here
EOF
fi

# ----------------------------------------------------------
# 3. Initialize Go module
# ----------------------------------------------------------
echo "[INFO] Initializing go.mod"

cat > "$GO_DIR/go.mod" <<'EOF'
module ap-controller-go

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/redis/go-redis/v9 v9.2.1
    gopkg.in/yaml.v3 v3.0.1
)
EOF

# ----------------------------------------------------------
# 4. Create production-grade Dockerfile (multi-stage)
# ----------------------------------------------------------
echo "[INFO] Creating Dockerfile"

cat > "$GO_DIR/Dockerfile" <<'EOF'
# ---------- build stage ----------
FROM golang:1.22 AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/ap-controller ./cmd/ap-controller

# ---------- runtime stage ----------
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /out/ap-controller /app/ap-controller
COPY config/controller.yaml /app/config/controller.yaml

ENV CONTROLLER_CONFIG=/app/config/controller.yaml

EXPOSE 8443

ENTRYPOINT ["/app/ap-controller"]
EOF

# ----------------------------------------------------------
# 5. Create README
# ----------------------------------------------------------
echo "[INFO] Creating README.md"

cat > "$GO_DIR/README.md" <<'EOF'
# ap-controller-go

Go-based implementation of the AP Controller.

## Design principles

- controller.yaml driven (no hard-coded logic)
- Redis-backed Session Schema v2
- role_rules with priority + wildcard support
- policy_version aware (AP-safe updates)
- HMAC-signed audit logs
- batch_status API for portal-sync.sh

## Build

```bash
docker build -t ap-controller-go .

## Run

```bash
docker run \
  -p 8443:8443 \
  -e AUDIT_SECRET=change_me \
  ap-controller-go

Notes

Python implementation remains untouched

Data-plane scripts (portal-fw.sh / portal-sync.sh) unchanged

docker-compose.yml can switch implementations safely
EOF

# ----------------------------------------------------------
# 6. Stub main.go (placeholder)
# ----------------------------------------------------------

echo "[INFO] Creating stub main.go"

cat > "$GO_DIR/cmd/ap-controller/main.go" <<'EOF'
package main

import "fmt"

func main() {
fmt.Println("ap-controller-go stub")
fmt.Println("Replace this with real controller implementation")
}
EOF

# ----------------------------------------------------------
# 7. Final summary
# ----------------------------------------------------------

echo
echo "[SUCCESS] Go controller skeleton created:"
echo " $GO_DIR"
echo
echo "Next steps:"
echo " 1. Implement config loader (controller.yaml)"
echo " 2. Port session store (Redis schema v2)"
echo " 3. Implement decide_role() with priority + wildcard"
echo " 4. Add audit logging (HMAC)"
echo " 5. Switch docker-compose.yml when ready"
echo
echo "No existing Python code or data-plane scripts were modified."


# ---

## 使用方式

# ```bash
# chmod +x refactor_ap_controller_to_go.sh
# ./refactor_ap_controller_to_go.sh
# ```