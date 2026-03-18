#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

API_PORT="${API_PORT:-8080}"
WEB_PORT="${WEB_PORT:-5173}"
VITE_API_BASE_URL="${VITE_API_BASE_URL:-http://localhost:${API_PORT}}"
MINIDOCK_RUNTIME_MODE="${MINIDOCK_RUNTIME_MODE:-processo-local}"
MINIDOCK_SEED_DEMO="${MINIDOCK_SEED_DEMO:-true}"
MINIDOCK_CONTAINER_ROOTFS="${MINIDOCK_CONTAINER_ROOTFS:-${ROOT_DIR}/examples/rootfs/demo}"

port_in_use() {
  local port="$1"
  lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
}

print_port_conflict() {
  local port="$1"
  echo "[minidock] porta ${port} já está em uso:"
  lsof -nP -iTCP:"${port}" -sTCP:LISTEN || true
}

cleanup() {
  if [[ -n "${API_PID:-}" ]]; then
    kill "${API_PID}" 2>/dev/null || true
  fi
  if [[ -n "${WEB_PID:-}" ]]; then
    kill "${WEB_PID}" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

if port_in_use "${API_PORT}"; then
  print_port_conflict "${API_PORT}"
  echo "[minidock] execute: make stop"
  exit 1
fi

if port_in_use "${WEB_PORT}"; then
  print_port_conflict "${WEB_PORT}"
  echo "[minidock] execute: make stop"
  exit 1
fi

echo "[minidock] iniciando API em http://localhost:${API_PORT}"
(
  cd "${ROOT_DIR}/apps/api"
  API_PORT="${API_PORT}" \
  MINIDOCK_RUNTIME_MODE="${MINIDOCK_RUNTIME_MODE}" \
  MINIDOCK_SEED_DEMO="${MINIDOCK_SEED_DEMO}" \
  MINIDOCK_CONTAINER_ROOTFS="${MINIDOCK_CONTAINER_ROOTFS}" \
  go run ./cmd/server
) &
API_PID=$!

echo "[minidock] iniciando Web em http://localhost:${WEB_PORT}"
(
  cd "${ROOT_DIR}/apps/web"
  VITE_API_BASE_URL="${VITE_API_BASE_URL}" npm run dev -- --host 0.0.0.0 --port "${WEB_PORT}" --strictPort
) &
WEB_PID=$!

EXIT_CODE=0
while true; do
  if ! kill -0 "${API_PID}" 2>/dev/null; then
    wait "${API_PID}" || EXIT_CODE=$?
    break
  fi
  if ! kill -0 "${WEB_PID}" 2>/dev/null; then
    wait "${WEB_PID}" || EXIT_CODE=$?
    break
  fi
  sleep 1
done

exit "${EXIT_CODE}"
