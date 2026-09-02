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

FULL_ZIP="dist/alert-impact-assessment-skill.zip"
AISTUDIO_ZIP="dist/alert-impact-assessment-aistudio-import.zip"
rm -f "$FULL_ZIP" "$AISTUDIO_ZIP"

echo "Packing full distribution zip..."
zip -r "$FULL_ZIP" \
  INSTALL.md README.md go.mod go.sum \
  "$SKILL_SRC/" \
  cmd/ internal/ config/ alerts/ scripts/ docs/ \
  deliverables/03-Skill设计说明文档.md \
  dist/bin/alert-assess-linux \
  dist/bin/alert-assess-darwin-arm64

echo "Packing AI Studio import zip (top-level skill folder)..."
AISTUDIO_DIR="$(mktemp -d)/alert-impact-assessment"
mkdir -p "$AISTUDIO_DIR"
cp -R "$SKILL_SRC/." "$AISTUDIO_DIR/"
(
  cd "$(dirname "$AISTUDIO_DIR")"
  zip -r "$ROOT/$AISTUDIO_ZIP" alert-impact-assessment/
)

echo ""
echo "Packages created:"
ls -lh "$FULL_ZIP" "$AISTUDIO_ZIP"
echo ""
echo "  $FULL_ZIP"
echo "    完整项目包（CLI + Skill 目录结构），解压见 INSTALL.md"
echo ""
echo "  $AISTUDIO_ZIP"
echo "    AI Studio 导入 Skill 用（zip 内含 alert-impact-assessment/SKILL.md）"
