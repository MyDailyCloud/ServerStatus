#!/usr/bin/env sh
# 示例：在 Docker 主机上自动拉取最新镜像并用 Watchtower 监控滚动更新

# 使用前请先登录镜像仓库（如 GHCR）：docker login ghcr.io

set -e

COMPOSE_FILE=${COMPOSE_FILE:-docker-compose.yml}
STACK_DIR=${STACK_DIR:-/opt/serverstatus}

echo "==> 切换到部署目录: ${STACK_DIR}"
cd "${STACK_DIR}"

echo "==> 拉取最新镜像"
docker compose -f "${COMPOSE_FILE}" pull

echo "==> 以滚动方式重启服务"
docker compose -f "${COMPOSE_FILE}" up -d

echo "==> 启动/更新 Watchtower（每30分钟检测一次）"
docker run -d --name watchtower \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower \
  --interval 1800 \
  serverstatus-server serverstatus-agent-1 serverstatus-agent-2

echo "完成。Watchtower 将自动检测镜像更新并滚动重启上述容器。"

