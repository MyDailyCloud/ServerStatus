#!/usr/bin/env bash
# 启动 docker compose（serverstatus），并按配置文件列表启动多个 agent
# 使用前请确保：
#   - docker / docker compose 可用
#   - data-server 与 monitor-agent 已构建或由 compose 提供
#   - agent 配置文件已准备（JSON），路径通过 AGENT_CONFIGS 传入
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-${ROOT}/deploy/docker-compose.yml}"
PROJECT_NAME="${PROJECT_NAME:-serverstatus}"
# 逗号分隔的 agent 配置文件列表（默认使用 configs/agent-*.json）
AGENT_CONFIGS="${AGENT_CONFIGS:-}"
# agent 可执行文件路径
AGENT_BIN="${AGENT_BIN:-${ROOT}/monitor-agent/monitor-agent}"

# docker-compose.yml 所需环境变量默认值，可被外部覆盖
export PROJECT_KEY="${PROJECT_KEY:-public}"
export SERVER_KEY="${SERVER_KEY:-serverstatus.ltd}"
export SERVER_PORT="${SERVER_PORT:-8080}"
export SERVER_URL="${SERVER_URL:-http://data-server:8080/api/data}"
export AGENT_HOSTNAME="${AGENT_HOSTNAME:-serverstatus-agent}"
export ENABLE_USER_RESOURCES="${ENABLE_USER_RESOURCES:-false}"
export DATA_DIR="${DATA_DIR:-${ROOT}/data}"
export LOG_DIR="${LOG_DIR:-${ROOT}/data-server/logs}"
export EXPORTS_DIR="${EXPORTS_DIR:-${ROOT}/data-server/exports}"
export FRONTEND_PORT="${FRONTEND_PORT:-8081}"

cleanup() {
  set +e
  if [ -f "${ROOT}/agent.pids" ]; then
    while read -r pid; do
      kill "${pid}" >/dev/null 2>&1 || true
    done < "${ROOT}/agent.pids"
    rm -f "${ROOT}/agent.pids"
  fi
}
trap cleanup EXIT

echo "==> 启动 docker compose（project=${PROJECT_NAME})"
# 强制清理可能残留的同名容器与网络，避免名称冲突
docker compose -f "${COMPOSE_FILE}" -p "${PROJECT_NAME}" down --remove-orphans --timeout 5 >/dev/null 2>&1 || true
# 防止容器名占用（特别是 serverstatus-frontend 可能残留）
docker rm -f serverstatus-frontend serverstatus-server serverstatus-agent >/dev/null 2>&1 || true
docker network rm "${PROJECT_NAME}_serverstatus" >/dev/null 2>&1 || true

mkdir -p "${DATA_DIR}" "${LOG_DIR}" "${EXPORTS_DIR}"
docker compose -f "${COMPOSE_FILE}" -p "${PROJECT_NAME}" up -d

HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:${SERVER_PORT}/api/servers}"
echo "==> 等待 data-server 健康检查 ${HEALTH_URL} ..."
for i in {1..30}; do
  if curl -sf "${HEALTH_URL}" >/dev/null 2>&1; then
    break
  fi
  sleep 1
  if [ "$i" -eq 30 ]; then
    echo "data-server 未就绪" >&2
    exit 1
  fi
done

# 确定 agent 配置列表
if [ -z "${AGENT_CONFIGS}" ]; then
  AGENT_CONFIGS=$(ls "${ROOT}"/configs/agent-*.json 2>/dev/null || true)
fi

if [ -z "${AGENT_CONFIGS}" ]; then
  echo "未找到 agent 配置，请通过 AGENT_CONFIGS 指定或放置 configs/agent-*.json" >&2
  exit 1
fi

echo "==> 启动 agents..."
> "${ROOT}/agent.pids"
for cfg in ${AGENT_CONFIGS//,/ }; do
  if [ ! -f "${cfg}" ]; then
    echo "配置不存在: ${cfg}" >&2
    exit 1
  fi
  echo "启动 agent，配置: ${cfg}"
  "${AGENT_BIN}" -config "${cfg}" >/dev/null 2>&1 &
  echo $! >> "${ROOT}/agent.pids"
done

echo "==> 冒烟检查 /api/servers"
curl -sf "http://127.0.0.1:8080/api/servers" >/dev/null

echo "✅ compose 与 agents 已启动（后台运行）。按需测试后可手动 docker compose down。"

