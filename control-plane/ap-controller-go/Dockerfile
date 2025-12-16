# =========================
# Build stage
# =========================
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.22-alpine AS builder

WORKDIR /build

# -------------------------
# Build arguments
# -------------------------
# Whether to generate OpenAPI spec and enable swagger build tag
ARG GEN_OPENAPI=true

# Go module proxy
ENV GOPROXY=https://goproxy.cn,direct

# -------------------------
# Speed up apk (use China mirror)
# -------------------------
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.tuna.tsinghua.edu.cn|g' /etc/apk/repositories

# swag requires git (only needed when generating OpenAPI)
RUN if [ "$GEN_OPENAPI" = "true" ]; then \
      apk add --no-cache git ; \
    fi

# -------------------------
# Go modules
# -------------------------
COPY go.mod go.sum ./
RUN go mod download

# -------------------------
# Source code
# -------------------------
COPY . .

# -------------------------
# Install swag & generate OpenAPI (optional)
# -------------------------
RUN if [ "$GEN_OPENAPI" = "true" ]; then \
      go install github.com/swaggo/swag/cmd/swag@v1.8.1 && \
      swag init \
        --generalInfo cmd/ap-controller/main.go \
        --output docs/openapi ; \
    fi

# -------------------------
# Build controller binary
# IMPORTANT:
# - GEN_OPENAPI=true  -> build with "swagger" tag
# - GEN_OPENAPI=false -> build without swagger
# -------------------------
RUN if [ "$GEN_OPENAPI" = "true" ]; then \
      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      go build -tags swagger -o ap-controller ./cmd/ap-controller ; \
    else \
      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      go build -o ap-controller ./cmd/ap-controller ; \
    fi


# =========================
# Runtime stage
# =========================
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/alpine:3.19

WORKDIR /app

# -------------------------
# Runtime arguments / env
# -------------------------
ARG ENABLE_SWAGGER_UI=false
ENV ENABLE_SWAGGER_UI=${ENABLE_SWAGGER_UI}

# CA certs for HTTPS / Redis TLS
RUN apk add --no-cache ca-certificates

# -------------------------
# Copy artifacts
# -------------------------
COPY --from=builder /build/ap-controller /app/ap-controller

# OpenAPI spec (exists only when GEN_OPENAPI=true, but directory is safe)
COPY --from=builder /build/docs/openapi /app/docs/openapi

# Controller config mount point
RUN mkdir -p /app/config

EXPOSE 8443

ENTRYPOINT ["/app/ap-controller"]
