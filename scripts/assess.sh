#!/usr/bin/env bash
# 告警业务影响评估 — 一键运行（只读）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

INPUT="${INPUT:-alerts/sample-alerts.json}"
OUTPUT="${OUTPUT:-deliverables}"
MOCK=""
SUMMARY=""
VALIDATE_ONLY=false

usage() {
  echo "Usage: $0 [--input FILE] [--output DIR] [--mock] [--summary] [--validate]"
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

mkdir -p bin "$OUTPUT"

if [[ ! -f bin/alert-assess ]] || [[ cmd/alert-assess/main.go -nt bin/alert-assess ]]; then
  go build -o bin/alert-assess ./cmd/alert-assess/
fi

if $VALIDATE_ONLY; then
  ./bin/alert-assess validate
  exit $?
fi

./bin/alert-assess assess \
  --input "$INPUT" \
  --output-dir reports \
  --deliverables "$OUTPUT" \
  $MOCK $SUMMARY

echo ""
echo "Done. Deliverables:"
echo "  $OUTPUT/01-告警业务影响评估报告.md"
echo "  $OUTPUT/02-分级处置建议与客户沟通话术.md"
