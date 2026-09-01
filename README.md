# 告警业务影响评估 Skill

通过 **Apsara Stack** 只读 OpenAPI 与 CMDB，对单条或批量告警自动完成业务影响评估、关联分析与分级处置建议生成。**所有操作为只读，不涉及环境变更。**

实现语言：**Go 1.22+**

## 功能

1. **业务归属链路还原**：实例 → 应用 → 业务系统 → 客户责任人
2. **客户感知指标**：QPS、错误率、P99 延迟近 30 分钟趋势
3. **关联告警聚合**：同源/同业务系统聚类，匹配已知预案
4. **分级处置建议**：P1/P2/P3 + 中性客户沟通话术

## 快速开始

```bash
# 构建
go build -o bin/alert-assess ./cmd/alert-assess/

# Mock 模式（无需 Apsara Stack 凭证）
./bin/alert-assess assess --input alerts/sample-alerts.json --mock --output-dir reports/
```

## 真实 API 模式（Apsara Stack）

Endpoint 需从 Apsara Uni-manager **服务注册变量** 获取：

```bash
export APSARASTACK_ACCESS_KEY_ID="your-readonly-access-key-id"
export APSARASTACK_ACCESS_KEY_SECRET="your-readonly-access-key-secret"
export APSARASTACK_REGION="cn-hangzhou-1"
export APSARASTACK_AZ="a"

# 各产品 POP Endpoint（从运维控制台获取）
export APSARASTACK_ECS_ENDPOINT="ecs.cn-hangzhou-1.xxx.stack.local"
export APSARASTACK_SLB_ENDPOINT="slb.cn-hangzhou-1.xxx.stack.local"
export APSARASTACK_RDS_ENDPOINT="rds.cn-hangzhou-1.xxx.stack.local"
export APSARASTACK_CMS_ENDPOINT="metrics.cn-hangzhou-1.xxx.stack.local"

export CMDB_API_URL="https://cmdb.example.com/api"   # 可选
export CMDB_API_TOKEN="your-token"                    # 可选

./bin/alert-assess assess --input alerts/sample-alerts.json --output-dir reports/
```

## Cursor Skill 使用

Skill 位于 `.cursor/skills/alert-impact-assessment/SKILL.md`。

## 文档

- [Skill 设计说明](docs/SKILL-DESIGN.md)
- [只读 OpenAPI 清单](docs/openapi-readonly-list.md)
- [预案场景库](config/runbook-scenarios.yaml)

## 项目结构

```
.cursor/skills/alert-impact-assessment/SKILL.md
cmd/alert-assess/main.go
internal/
  apsarastack/          # Apsara Stack POP 只读 API 客户端
  processor/            # 告警处理核心
  report/               # 报告生成
  mockdata/             # Mock 演示数据
  models/               # 数据结构
config/runbook-scenarios.yaml
alerts/sample-alerts.json
```

## 安全说明

- 客户端硬编码 Action 前缀白名单（Describe/Get/List/Query/Search）
- 建议使用 RAM 只读子账号
- POP Endpoint 因 Stack 部署而异，必须通过环境变量配置
