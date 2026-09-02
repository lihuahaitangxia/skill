# 告警业务影响评估 Skill

通过 Apsara Stack 只读 POP OpenAPI 与 CMDB，对单条或批量告警自动完成业务影响评估。**所有操作为只读。**

## 作为 AI Studio Skill 使用

Skill 已内置在仓库：

```
.cursor/skills/alert-impact-assessment/SKILL.md
```

在 AI Studio 中导入 **`dist/alert-impact-assessment-aistudio.zip`**（含 Skill + Go 二进制，无需本机安装 Go）。

触发示例：

- 「评估这批告警的业务影响」
- 「生成告警影响评估报告」
- 「分析 Apsara Stack 告警并输出 P1/P2/P3 建议」

## 快速运行（CLI）

```bash
# 构建
go build -o bin/alert-assess ./cmd/alert-assess/

# Mock 演示
./scripts/assess.sh --mock

# 真实 Apsara Stack（需配置环境变量）
export APSARASTACK_ACCESS_KEY_ID="..."
export APSARASTACK_ACCESS_KEY_SECRET="..."
export APSARASTACK_ECS_ENDPOINT="..."
./scripts/assess.sh
```

## 交付物

| 文件 | 说明 |
|------|------|
| `deliverables/01-告警业务影响评估报告.md` | 链路图 + 指标快照 + 关联聚合 |
| `deliverables/02-分级处置建议与客户沟通话术.md` | P1/P2/P3 + 中性话术 |
| `deliverables/03-Skill设计说明文档.md` | 输入输出 + OpenAPI + 限流 |

## 安装到其他项目

将 `.cursor/skills/alert-impact-assessment/` 复制到目标项目 Skill 目录，并 Clone 本仓库代码（Skill 依赖 `cmd/` 与 `internal/` 中的 Go CLI）。

## 打包分发

```bash
chmod +x scripts/package.sh
./scripts/package.sh
```

输出两个 zip：

| 文件 | 用途 |
|------|------|
| **`dist/alert-impact-assessment-aistudio.zip`** | **AI Studio 导入**（Skill + Go 二进制 + 配置，推荐） |
| `dist/alert-impact-assessment-skill.zip` | 开发者完整包（含 Go 源码），解压见 [INSTALL.md](INSTALL.md) |

## 文档

- [Skill 设计说明](docs/SKILL-DESIGN.md)
- [只读 OpenAPI 清单](docs/openapi-readonly-list.md)
- [Skill 参考文档](.cursor/skills/alert-impact-assessment/references/)
