# 安装说明

## 解压

**完整项目包**（含 CLI、Go 源码）：

```bash
unzip alert-impact-assessment-skill.zip -d my-alert-skill
cd my-alert-skill
```

**Cursor 导入 Skill**（Settings → Skills → Import，需 zip 根目录有 `SKILL.md`）：

```bash
# 使用 alert-impact-assessment-cursor-import.zip，不是完整项目包
```

## 方式 A：作为 Cursor Skill 使用

1. 将 `.cursor/skills/alert-impact-assessment/` 复制到你的项目 `.cursor/skills/` 下
2. 将 `cmd/`、`internal/`、`config/`、`scripts/`、`go.mod` 等复制到项目根目录（或保持本目录结构）
3. 在 Cursor 中说「评估告警业务影响」即可触发

## 方式 B：仅 CLI 使用（无需 Cursor）

```bash
# 需要 Go 1.22+
./scripts/assess.sh --validate
./scripts/assess.sh --mock --summary
```

## 方式 C：预编译二进制（无需 Go）

Linux amd64 用户可直接使用 `bin/alert-assess-linux`（若包内提供）：

```bash
chmod +x bin/alert-assess-linux
./bin/alert-assess-linux assess --input alerts/sample-alerts.json --mock --deliverables deliverables
```

## 环境变量（真实 Apsara Stack）

```bash
export APSARASTACK_ACCESS_KEY_ID="..."
export APSARASTACK_ACCESS_KEY_SECRET="..."
export APSARASTACK_REGION="cn-hangzhou-1"
export APSARASTACK_AZ="a"
export APSARASTACK_ECS_ENDPOINT="..."
export APSARASTACK_SLB_ENDPOINT="..."
export APSARASTACK_RDS_ENDPOINT="..."
export APSARASTACK_CMS_ENDPOINT="..."
```

## 包内文件说明

| 路径 | 说明 |
|------|------|
| `.cursor/skills/alert-impact-assessment/` | Cursor Skill |
| `cmd/` + `internal/` | Go 评估引擎 |
| `scripts/assess.sh` | 一键运行脚本 |
| `config/runbook-scenarios.yaml` | 预案库 |
| `alerts/sample-alerts.json` | 样例输入 |
| `docs/` | 设计文档 |
| `deliverables/03-Skill设计说明文档.md` | 完整设计说明 |

## 输出

运行后在 `deliverables/` 生成：

- `01-告警业务影响评估报告.md`
- `02-分级处置建议与客户沟通话术.md`
