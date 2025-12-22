#!/bin/sh
set -e

echo "[redis-entrypoint] starting redis..."

# 后台启动 redis
redis-server &
REDIS_PID=$!

# 等 redis 就绪
echo "[redis-entrypoint] waiting for redis..."
until redis-cli ping >/dev/null 2>&1; do
  sleep 0.2
done

echo "[redis-entrypoint] redis is ready"

# 执行初始化脚本（只要存在就执行）
if [ -f /docker-entrypoint-initdb.d/00-portal-redis-schema.sh ]; then
  echo "[redis-entrypoint] running portal redis bootstrap"
  sh /docker-entrypoint-initdb.d/00-portal-redis-schema.sh
fi

# 前台等待 redis
wait $REDIS_PID
