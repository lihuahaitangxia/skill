#!/usr/bin/env bash
# 打包 Skill 分发 zip
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SKILL_SRC=".cursor/skills/alert-impact-assessment"
if [[ ! -f "$SKILL_SRC/SKILL.md" ]]; then
  echo "ERROR: 未找到 $SKILL_SRC/SKILL.md"
  echo "请确认在项目根目录运行（应包含 cmd/、internal/、.cursor/skills/）"
  exit 1
fi

echo "Building binaries..."
mkdir -p dist/bin
GOOS=linux GOARCH=amd64 go build -o dist/bin/alert-assess-linux ./cmd/alert-assess/
GOOS=darwin GOARCH=arm64 go build -o dist/bin/alert-assess-darwin-arm64 ./cmd/alert-assess/
GOOS=darwin GOARCH=amd64 go build -o dist/bin/alert-assess-darwin-amd64 ./cmd/alert-assess/

FULL_ZIP="dist/alert-impact-assessment-skill.zip"
AISTUDIO_ZIP="dist/alert-impact-assessment-aistudio.zip"
rm -f "$FULL_ZIP" "$AISTUDIO_ZIP"

echo "Packing full distribution zip (dev / 含 Go 源码)..."
zip -r "$FULL_ZIP" \
  INSTALL.md README.md go.mod go.sum \
  "$SKILL_SRC/" \
  cmd/ internal/ config/ alerts/ scripts/ docs/ \
  deliverables/03-Skill设计说明文档.md \
  dist/bin/alert-assess-linux \
  dist/bin/alert-assess-darwin-arm64 \
  dist/bin/alert-assess-darwin-amd64

echo "Packing AI Studio zip (Skill + Go 二进制 + 配置，可直接导入)..."
AISTUDIO_DIR="$(mktemp -d)/alert-impact-assessment"
mkdir -p "$AISTUDIO_DIR/scripts" "$AISTUDIO_DIR/bin" "$AISTUDIO_DIR/deliverables" "$AISTUDIO_DIR/reports"
cp -R "$SKILL_SRC/." "$AISTUDIO_DIR/"
cp scripts/assess.sh "$AISTUDIO_DIR/scripts/"
chmod +x "$AISTUDIO_DIR/scripts/assess.sh"
cp -R config alerts docs "$AISTUDIO_DIR/"
cp dist/bin/alert-assess-linux \
  dist/bin/alert-assess-darwin-arm64 \
  dist/bin/alert-assess-darwin-amd64 \
  "$AISTUDIO_DIR/bin/"
chmod +x "$AISTUDIO_DIR/bin/"*
touch "$AISTUDIO_DIR/deliverables/.gitkeep"
(
  cd "$(dirname "$AISTUDIO_DIR")"
  zip -r "$ROOT/$AISTUDIO_ZIP" alert-impact-assessment/
)

echo ""
echo "Packages created:"
ls -lh "$FULL_ZIP" "$AISTUDIO_ZIP"
echo ""
echo "  $AISTUDIO_ZIP"
echo "    ★ AI Studio 导入此包（含 SKILL.md + Go 二进制 + scripts/config/alerts）"
echo ""
echo "  $FULL_ZIP"
echo "    开发者完整包（含 Go 源码），解压见 INSTALL.md"
