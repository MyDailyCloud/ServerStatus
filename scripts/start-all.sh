#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${PROJECT_ROOT}/.env"
BUILD_FLAG="--build"

usage() {
  echo "Usage: $0 [--no-build]"
  echo "  --no-build   跳过构建，直接使用已存在的镜像"
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-build|--skip-build)
      BUILD_FLAG="--no-build"
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Unknown option: $1"
      usage
      ;;
  esac
done

generate_secret() {
  LC_CTYPE=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c24
}

ensure_env() {
  local updated=0
  umask 077
  : > /dev/null

  if [[ ! -f "$ENV_FILE" ]]; then
    touch "$ENV_FILE"
  fi

  if ! grep -q '^PROJECT_KEY=' "$ENV_FILE"; then
    echo "PROJECT_KEY=$(generate_secret)" >> "$ENV_FILE"
    updated=1
  fi

  if ! grep -q '^SERVER_KEY=' "$ENV_FILE"; then
    echo "SERVER_KEY=$(generate_secret)" >> "$ENV_FILE"
    updated=1
  fi

  if ! grep -q '^SERVER_PORT=' "$ENV_FILE"; then
    echo "SERVER_PORT=8080" >> "$ENV_FILE"
    updated=1
  fi

  # Agent defaults (single agent)
  if ! grep -q '^AGENT_HOSTNAME=' "$ENV_FILE"; then
    echo "AGENT_HOSTNAME=serverstatus-agent" >> "$ENV_FILE"
    updated=1
  fi
  if ! grep -q '^ENABLE_USER_RESOURCES=' "$ENV_FILE"; then
    echo "ENABLE_USER_RESOURCES=false" >> "$ENV_FILE"
    updated=1
  fi

  if ! grep -q '^SERVER_URL=' "$ENV_FILE"; then
    echo "SERVER_URL=http://data-server:8080/api/data" >> "$ENV_FILE"
    updated=1
  fi

  if [[ $updated -eq 1 ]]; then
    echo "Secrets/env created/updated at .env (values not shown)."
  fi
}

cd "$PROJECT_ROOT"

ensure_env
set -a
. "$ENV_FILE"
set +a

echo "Stopping and cleaning previous stack..."
docker compose --env-file "$ENV_FILE" -f deploy/docker-compose.yml down -v

if [[ "$BUILD_FLAG" == "--no-build" ]]; then
  echo "Starting stack without rebuild (server + frontend + 1 agent)..."
else
  echo "Building and starting stack (server + frontend + 1 agent)..."
fi
docker compose --env-file "$ENV_FILE" -f deploy/docker-compose.yml up "$BUILD_FLAG" -d

echo "Done. Endpoints:"
echo "  API/UI:   http://localhost:8080"
echo "  Frontend: http://localhost:8081"
