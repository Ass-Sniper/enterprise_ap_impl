# =========================
# Build stage
# =========================
# 原始: FROM golang:1.22-alpine AS builder
# 修改为华为云 SWR 路径
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.22-alpine AS builder

WORKDIR /build

# Enable Go proxy (optional)
# 这一行是用于 Go 模块下载的代理，与基础镜像拉取无关，可以保留
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o ap-controller ./cmd/ap-controller

# =========================
# Runtime stage
# =========================
# 原始: FROM alpine:3.19
# 修改为华为云 SWR 路径
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/alpine:3.19

WORKDIR /app

# CA certs for HTTPS / Redis TLS
RUN apk add --no-cache ca-certificates

COPY --from=builder /build/ap-controller /app/ap-controller

# Controller config will be mounted here
RUN mkdir -p /app/config

EXPOSE 8443

ENTRYPOINT ["/app/ap-controller"]
