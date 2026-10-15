#!/usr/bin/env bash
#
# 停止 resume 服务（先 pidfile，再 SERVICE_PORT 上的 LISTEN）。
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUNTIME_DIR="${RUNTIME_DIR:-$(cd "${DIR}/.." && pwd)}"
PORT="${SERVICE_PORT:-${PORT:-4000}}"
PID_FILE="${RUNTIME_DIR}/backend/runtime.pid"

log() { echo "[stop] $*"; }

if [ -f "${PID_FILE}" ]; then
  pid="$(tr -d '[:space:]' < "${PID_FILE}" || true)"
  if [ -n "${pid}" ] && kill -0 "${pid}" 2>/dev/null; then
    log "TERM → pid=${pid}"
    kill "${pid}" 2>/dev/null || true
    for _ in $(seq 1 30); do
      kill -0 "${pid}" 2>/dev/null || break
      sleep 1
    done
    if kill -0 "${pid}" 2>/dev/null; then
      log "KILL → pid=${pid}"
      kill -9 "${pid}" 2>/dev/null || true
    fi
  fi
  rm -f "${PID_FILE}"
else
  log "没有 pid 文件，视为未运行"
fi

# 兜底：把还占着端口的进程停掉。
case "${PORT}" in
  *[!0-9]*|"") ;;
  *)
    pids="$(lsof -ti tcp:"${PORT}" 2>/dev/null || true)"
    if [ -n "${pids}" ]; then
      log "按端口 ${PORT} 清理残留进程: ${pids}"
      kill ${pids} 2>/dev/null || true
    fi
    ;;
esac

log "done"
