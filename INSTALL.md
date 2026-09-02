# 安装说明

## AI Studio 导入（推荐，含 Go 二进制）

1. 打包：

```bash
chmod +x scripts/package.sh
./scripts/package.sh
```

2. 在 AI Studio → Skills → **+ 安装** → **上传 ZIP**，选择：

```
dist/alert-impact-assessment-aistudio.zip
```

3. 包内结构：

```
alert-impact-assessment-aistudio.zip
└── alert-impact-assessment/
    ├── SKILL.md
    ├── references/
    ├── scripts/assess.sh      # 无需本机安装 Go
    ├── bin/                   # 预编译 alert-assess（Mac/Linux）
    ├── config/
    ├── alerts/
    └── docs/
```

4. 在 AI Studio 对话中触发 Skill 后，终端命令应在 Skill 根目录执行：

```bash
cd alert-impact-assessment
./scripts/assess.sh --validate
./scripts/assess.sh --mock --summary
```

## 解压完整开发包（含 Go 源码）

```bash
unzip alert-impact-assessment-skill.zip -d my-alert-skill
cd my-alert-skill
```

## 方式 A：作为 AI Studio Skill 使用

导入 `alert-impact-assessment-aistudio.zip` 即可，包内已含 Go 二进制，**不需要**单独安装 Go。

## 方式 B：开发者 CLI（含源码）

```bash
# 需要 Go 1.22+
./scripts/assess.sh --validate
./scripts/assess.sh --mock --summary
```

## 方式 C：预编译二进制（手动调用）

```bash
chmod +x bin/alert-assess-darwin-arm64   # Apple Silicon Mac
./bin/alert-assess-darwin-arm64 assess --input alerts/sample-alerts.json --mock --deliverables deliverables
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
| `SKILL.md` + `references/` | AI Studio Skill |
| `bin/` | 预编译 Go CLI（AI Studio 包） |
| `cmd/` + `internal/` | Go 源码（仅完整开发包） |
| `scripts/assess.sh` | 一键运行脚本 |
| `config/runbook-scenarios.yaml` | 预案库 |
| `alerts/sample-alerts.json` | 样例输入 |
| `docs/` | 设计文档 |

## 输出

运行后在 `deliverables/` 生成：

- `01-告警业务影响评估报告.md`
- `02-分级处置建议与客户沟通话术.md`
