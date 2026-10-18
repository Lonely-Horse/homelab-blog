#!/usr/bin/env bash
set -Eeuo pipefail

APP_DIR="/home/lonelyhorse/server/homelab-blog"

cd "$APP_DIR"

echo "正在拉取 latest 镜像..."
docker compose pull

echo "正在更新服务..."
docker compose up -d

echo "✅ T630 部署完成"
