#!/usr/bin/env bash
set -euo pipefail

API_PORT="${API_PORT:-8080}"
WEB_PORT="${WEB_PORT:-5173}"

kill_port() {
  local port="$1"
  local pids

  pids="$(lsof -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
  if [[ -z "${pids}" ]]; then
    echo "[minidock] porta ${port}: livre"
    return
  fi

  echo "[minidock] encerrando processo(s) na porta ${port}: ${pids}"
  # shellcheck disable=SC2086
  kill ${pids} 2>/dev/null || true
  sleep 1

  pids="$(lsof -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
  if [[ -n "${pids}" ]]; then
    echo "[minidock] forçando término na porta ${port}: ${pids}"
    # shellcheck disable=SC2086
    kill -9 ${pids} 2>/dev/null || true
  fi
}

kill_port "${API_PORT}"
kill_port "${WEB_PORT}"
