#!/usr/bin/env bash
# 打包 Skill 分发 zip
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "Building binaries..."
mkdir -p dist/bin
GOOS=linux GOARCH=amd64 go build -o dist/bin/alert-assess-linux ./cmd/alert-assess/
GOOS=darwin GOARCH=arm64 go build -o dist/bin/alert-assess-darwin-arm64 ./cmd/alert-assess/

ZIP="dist/alert-impact-assessment-skill.zip"
rm -f "$ZIP"

zip -r "$ZIP" \
  INSTALL.md README.md go.mod go.sum \
  .cursor/skills/alert-impact-assessment/ \
  cmd/ internal/ config/ alerts/ scripts/ docs/ \
  deliverables/03-Skill设计说明文档.md \
  dist/bin/alert-assess-linux \
  dist/bin/alert-assess-darwin-arm64

echo ""
echo "Package created: $ZIP"
ls -lh "$ZIP"
