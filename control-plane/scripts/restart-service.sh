#!/usr/bin/env bash
set -e

SERVICE="$1"
NO_CACHE="$2"

usage() {
  echo "Usage:"
  echo "  $0 <service-name> [--no-cache]"
  echo ""
  echo "Examples:"
  echo "  $0 freeradius"
  echo "  $0 freeradius --no-cache"
  echo "  $0 all --no-cache"
  exit 1
}

if [[ -z "$SERVICE" ]]; then
  usage
fi

echo "========================================"
echo " Docker Compose restart helper"
echo " Service   : $SERVICE"
echo " No cache  : ${NO_CACHE:-false}"
echo "========================================"

if [[ "$SERVICE" == "all" ]]; then
  echo "[1/4] Stopping all services..."
  docker compose stop

  echo "[2/4] Removing all services..."
  docker compose rm -f

  echo "[3/4] Building all services..."
  if [[ "$NO_CACHE" == "--no-cache" ]]; then
    docker compose build --no-cache
  else
    docker compose build
  fi

  echo "[4/4] Starting all services..."
  docker compose up -d
else
  echo "[1/4] Stopping service: $SERVICE"
  docker compose stop "$SERVICE" || true

  echo "[2/4] Removing service: $SERVICE"
  docker compose rm -f "$SERVICE" || true

  echo "[3/4] Building service: $SERVICE"
  if [[ "$NO_CACHE" == "--no-cache" ]]; then
    docker compose build --no-cache "$SERVICE"
  else
    docker compose build "$SERVICE"
  fi

  echo "[4/4] Starting service: $SERVICE"
  docker compose up -d "$SERVICE"
fi

echo "========================================"
echo " Done."
echo "========================================"
