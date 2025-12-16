#!/usr/bin/env bash
set -e

ROOT="control-plane"

echo "Creating control plane structure..."

mkdir -p $ROOT/ap-controller/app
mkdir -p $ROOT/captive-portal/{portal-server,portal-agent}

# ap-controller
touch $ROOT/ap-controller/{Dockerfile,docker-compose.yml,README.md}

# captive-portal
touch $ROOT/captive-portal/{Dockerfile,README.md}

# control-plane compose
touch $ROOT/docker-compose.yml

echo "Done."
echo
tree -L 3 $ROOT || true

