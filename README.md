# 告警业务影响评估 Skill

通过腾讯云只读 OpenAPI 与 CMDB，对单条或批量告警自动完成业务影响评估、关联分析与分级处置建议生成。**所有操作为只读，不涉及环境变更。**

## 功能

1. **业务归属链路还原**：实例 → 应用 → 业务系统 → 客户责任人
2. **客户感知指标**：QPS、错误率、P99 延迟近 30 分钟趋势
3. **关联告警聚合**：同源/同业务系统聚类，匹配已知预案
4. **分级处置建议**：P1/P2/P3 + 中性客户沟通话术

## 快速开始

```bash
pip install -r requirements.txt

# Mock 模式（无需腾讯云凭证）
python -m src.cli assess --input alerts/sample-alerts.json --mock --output-dir reports/
```

输出：
- `reports/{timestamp}-impact-assessment-report.md` — 告警业务影响评估报告
- `reports/{timestamp}-handling-recommendation.md` — 分级处置建议与客户沟通话术

## 真实 API 模式

```bash
export TENCENTCLOUD_SECRET_ID="your-readonly-secret-id"
export TENCENTCLOUD_SECRET_KEY="your-readonly-secret-key"
export TENCENTCLOUD_REGION="ap-guangzhou"
export CMDB_API_URL="https://cmdb.example.com/api"   # 可选
export CMDB_API_TOKEN="your-token"                    # 可选

python -m src.cli assess --input alerts/sample-alerts.json --output-dir reports/
```

## Cursor Skill 使用

Skill 位于 `.cursor/skills/alert-impact-assessment/SKILL.md`。在 Cursor 中提及「告警评估」「业务影响分析」等场景时，Agent 将自动加载并按工作流执行。

## 文档

- [Skill 设计说明](docs/SKILL-DESIGN.md) — 输入输出、限流策略、分级规则
- [只读 OpenAPI 清单](docs/openapi-readonly-list.md) — 依赖接口与 CAM 策略
- [预案场景库](config/runbook-scenarios.yaml) — 已知问题匹配规则

## 项目结构

```
.cursor/skills/alert-impact-assessment/SKILL.md  # Cursor Skill
src/
  cli.py                    # CLI 入口
  alert_processor.py        # 告警处理核心
  report_generator.py       # 报告生成
  mock_data.py              # Mock 演示数据
  tencent_cloud/            # 腾讯云只读 API 客户端
config/runbook-scenarios.yaml
alerts/sample-alerts.json
docs/
```

## 安全说明

- 客户端硬编码 Action 前缀白名单（Describe/Get/List/Query/Search）
- 建议使用具备只读 CAM 策略的子账号
- 报告中联系人信息自动脱敏
