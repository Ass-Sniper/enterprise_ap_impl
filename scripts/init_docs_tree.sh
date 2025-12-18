#!/usr/bin/env bash
set -e

BASE_DIR="docs"

echo "📁 Initializing docs directory structure (idempotent)..."

# 创建目录（mkdir -p 本身就是幂等的）
mkdir -p "${BASE_DIR}"

# 需要的 markdown 文件列表
MD_FILES=(
  "README.md"
  "overview.md"
  "captive-portal.md"
  "os-portal-detection.md"
  "data-plane.md"
)

# 只在文件不存在时创建
for file in "${MD_FILES[@]}"; do
  target="${BASE_DIR}/${file}"
  if [ ! -f "$target" ]; then
    echo "  ➕ creating ${target}"
    touch "$target"
  else
    echo "  ✔ exists ${target}, skip"
  fi
done

# 图片资源目录
mkdir -p "${BASE_DIR}/assets/images"/{ios,android,windows,architecture}

# 抓包目录
mkdir -p "${BASE_DIR}/artifacts/pcap"/{ios,android,windows}

# Mermaid / diagram 目录
mkdir -p "${BASE_DIR}/diagrams"

# 给“真正空目录”补 .gitkeep
find "${BASE_DIR}" -type d -empty -exec touch {}/.gitkeep \;

echo "✅ Docs directory structure initialized safely."
echo
echo "Tree:"
tree "${BASE_DIR}" || echo "(tree not installed)"