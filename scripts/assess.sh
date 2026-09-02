#!/usr/bin/env bash
# 告警业务影响评估 — 一键运行（只读）
# 支持：完整 Go 仓库（go build）或 AI Studio 独立包（预编译 bin/）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

INPUT="${INPUT:-alerts/sample-alerts.json}"
OUTPUT="${OUTPUT:-deliverables}"
MOCK=""
SUMMARY=""
VALIDATE_ONLY=false
BIN=""

usage() {
  echo "Usage: $0 [--input FILE] [--output DIR] [--mock] [--summary] [--validate]"
}

resolve_binary() {
  if [[ -n "$BIN" ]]; then
    return 0
  fi
  if [[ -x "$ROOT/bin/alert-assess" ]]; then
    BIN="$ROOT/bin/alert-assess"
    return 0
  fi
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"
  if [[ "$os" == "Darwin" ]]; then
    if [[ "$arch" == "arm64" && -f "$ROOT/bin/alert-assess-darwin-arm64" ]]; then
      BIN="$ROOT/bin/alert-assess-darwin-arm64"
    elif [[ -f "$ROOT/bin/alert-assess-darwin-amd64" ]]; then
      BIN="$ROOT/bin/alert-assess-darwin-amd64"
    fi
  elif [[ "$os" == "Linux" && -f "$ROOT/bin/alert-assess-linux" ]]; then
    BIN="$ROOT/bin/alert-assess-linux"
  fi
  if [[ -n "$BIN" ]]; then
    chmod +x "$BIN"
    return 0
  fi
  if [[ -f "$ROOT/go.mod" ]] && command -v go >/dev/null 2>&1; then
    mkdir -p "$ROOT/bin"
    go build -o "$ROOT/bin/alert-assess" ./cmd/alert-assess/
    BIN="$ROOT/bin/alert-assess"
    return 0
  fi
  echo "ERROR: 未找到 alert-assess 可执行文件，且无法 go build（需要 Go 1.22+ 或预编译 bin/）" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --input|-i) INPUT="$2"; shift 2 ;;
    --output|-o) OUTPUT="$2"; shift 2 ;;
    --mock) MOCK="--mock"; shift ;;
    --summary) SUMMARY="--summary"; shift ;;
    --validate) VALIDATE_ONLY=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1"; usage; exit 1 ;;
  esac
done

mkdir -p "$OUTPUT" reports
resolve_binary

if $VALIDATE_ONLY; then
  "$BIN" validate
  exit $?
fi

"$BIN" assess \
  --input "$INPUT" \
  --output-dir reports \
  --deliverables "$OUTPUT" \
  $MOCK $SUMMARY

echo ""
echo "Done. Deliverables:"
echo "  $OUTPUT/01-告警业务影响评估报告.md"
echo "  $OUTPUT/02-分级处置建议与客户沟通话术.md"
