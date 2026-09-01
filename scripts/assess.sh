#!/usr/bin/env bash
# 告警业务影响评估 — 一键运行脚本（只读）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

INPUT="${INPUT:-alerts/sample-alerts.json}"
OUTPUT="${OUTPUT:-deliverables}"
MOCK=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --input|-i) INPUT="$2"; shift 2 ;;
    --output|-o) OUTPUT="$2"; shift 2 ;;
    --mock) MOCK="--mock"; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

mkdir -p bin "$OUTPUT"
go build -o bin/alert-assess ./cmd/alert-assess/
./bin/alert-assess assess --input "$INPUT" $MOCK --output-dir reports/

# 同步最新报告到 deliverables 固定文件名
LATEST_IMPACT=$(ls -t reports/*-impact-assessment-report.md 2>/dev/null | head -1)
LATEST_HANDLING=$(ls -t reports/*-handling-recommendation.md 2>/dev/null | head -1)

if [[ -n "$LATEST_IMPACT" ]]; then
  cp "$LATEST_IMPACT" "$OUTPUT/01-告警业务影响评估报告.md"
fi
if [[ -n "$LATEST_HANDLING" ]]; then
  cp "$LATEST_HANDLING" "$OUTPUT/02-分级处置建议与客户沟通话术.md"
fi

echo ""
echo "Deliverables updated:"
echo "  $OUTPUT/01-告警业务影响评估报告.md"
echo "  $OUTPUT/02-分级处置建议与客户沟通话术.md"
