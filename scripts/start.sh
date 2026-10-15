#!/usr/bin/env bash
#
# 启动 resume 服务。由控制面以 restartCmd/startCmd 调用：cwd = RUNTIME_DIR，
# 注入 SERVICE_PORT / PORT / APP_VERSION / RUNTIME_DIR。
#
# runtime 布局（backend/ 内容由平台保留，不会被 --delete 清掉）：
#   bin/resume-server     可执行文件（来自发版包）
#   scripts/*.sh          本目录（来自发版包）
#   web/                  前端静态资源（来自发版包）
#   backend/runtime.pid   进程号
#   backend/server.log    标准输出/错误
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUNTIME_DIR="${RUNTIME_DIR:-$(cd "${DIR}/.." && pwd)}"
PORT="${SERVICE_PORT:-${PORT:-4000}}"
APP_VERSION="${APP_VERSION:-dev}"

BIN="${RUNTIME_DIR}/bin/resume-server"
WEB_DIR="${RUNTIME_DIR}/web"
BACKEND="${RUNTIME_DIR}/backend"
PID_FILE="${BACKEND}/runtime.pid"
LOG_FILE="${BACKEND}/server.log"

log() { echo "[start] $*"; }
die() { echo "[start][错误] $*" >&2; exit 1; }

case "${PORT}" in
  *[!0-9]*|"") die "端口必须是 1-65535 的整数，当前：${PORT}（来自 SERVICE_PORT/PORT）" ;;
esac
if [ "${PORT}" -lt 1 ] || [ "${PORT}" -gt 65535 ]; then
  die "端口超出范围：${PORT}（SERVICE_PORT/PORT 应为 1-65535）"
fi

[ -x "${BIN}" ] || die "缺少可执行文件 ${BIN}（发版包内容不完整？）"
mkdir -p "${BACKEND}"

# 已经跑着就不再重复启动（幂等）。
if [ -f "${PID_FILE}" ]; then
  old="$(tr -d '[:space:]' < "${PID_FILE}" || true)"
  if [ -n "${old}" ] && kill -0 "${old}" 2>/dev/null; then
    log "already running pid=${old}"
    exit 0
  fi
  rm -f "${PID_FILE}"
fi

export PORT="${PORT}"
export RUNTIME_DIR="${RUNTIME_DIR}"
export APP_VERSION="${APP_VERSION}"
export WEB_DIR="${WEB_DIR}"
# PostgreSQL：默认本机库；可用 runtimeEnv 注入 DATABASE_URL 覆盖。
export DATABASE_URL="${DATABASE_URL:-postgres://gaolei@localhost:5432/resume?sslmode=disable}"

cd "${RUNTIME_DIR}"
nohup "${BIN}" >> "${LOG_FILE}" 2>&1 &
echo $! > "${PID_FILE}"
log "pid=$(cat "${PID_FILE}") port=${PORT}"
