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
CURSOR_ZIP="dist/alert-impact-assessment-cursor-import.zip"
rm -f "$FULL_ZIP" "$CURSOR_ZIP"

echo "Packing full distribution zip..."
zip -r "$FULL_ZIP" \
  INSTALL.md README.md go.mod go.sum \
  "$SKILL_SRC/" \
  cmd/ internal/ config/ alerts/ scripts/ docs/ \
  deliverables/03-Skill设计说明文档.md \
  dist/bin/alert-assess-linux \
  dist/bin/alert-assess-darwin-arm64

echo "Packing Cursor Skill import zip (SKILL.md at root)..."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
cp -R "$SKILL_SRC/." "$TMP_DIR/"
(
  cd "$TMP_DIR"
  zip -r "$ROOT/$CURSOR_ZIP" SKILL.md references/
)

echo ""
echo "Packages created:"
ls -lh "$FULL_ZIP" "$CURSOR_ZIP"
echo ""
echo "  $FULL_ZIP"
echo "    完整项目包（CLI + Skill 目录结构），解压见 INSTALL.md"
echo ""
echo "  $CURSOR_ZIP"
echo "    Cursor 导入 Skill 用（zip 根目录含 SKILL.md）"
