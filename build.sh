#!/usr/bin/env bash
#
# Release build for the resume service.
#
# Invoked by the deployment control plane (packageFromGit) with:
#   cwd = repo root, APP_VERSION = <short hash>
# Must produce outputs/ containing the runtime fileset that gets rsynced into
# RUNTIME_DIR:
#   bin/resume-server, scripts/*.sh, web/*  (frontend static build)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

VERSION="${APP_VERSION:-local}"
OUT="${ROOT}/outputs"
rm -rf "${OUT}"
mkdir -p "${OUT}/bin" "${OUT}/scripts"

# Isolate Go caches inside the build tree so they never land in the runtime dir.
export GOMODCACHE="${ROOT}/.gomodcache"
export GOCACHE="${ROOT}/.gocache"
export GOPATH="${ROOT}/.gopath"
# Default to a reachable module proxy if the caller didn't set one.
if [ -z "${GOPROXY:-}" ]; then
  export GOPROXY="https://goproxy.cn,direct"
fi

echo "[build] version=${VERSION}"

# Backend (Go). Respects GOOS/GOARCH injected by the platform for remote targets.
(cd backend && go build -trimpath -o "${OUT}/bin/resume-server" .)

# PDF text extractor (macOS PDFKit) — optional; backend falls back to its own parser if absent.
if command -v swiftc >/dev/null 2>&1; then
  swiftc -O tools/pdftext.swift -o "${OUT}/bin/pdftext" || echo "[build][warn] pdftext 编译失败（不影响主流程）" >&2
fi

# Frontend (React). npm ci uses the committed package-lock.json for reproducible builds.
(cd frontend && npm ci --no-audit --no-fund --prefer-offline && npm run build)
cp -r frontend/dist "${OUT}/web"

# Runtime scripts.
cp scripts/*.sh "${OUT}/scripts/"
chmod +x "${OUT}/bin/"* "${OUT}/scripts/"*.sh

echo "[build] outputs ready:"
ls -1 "${OUT}"
