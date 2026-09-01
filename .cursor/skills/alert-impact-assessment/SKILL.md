---
name: alert-impact-assessment
description: >-
  告警业务影响评估 Skill。通过腾讯云只读 OpenAPI 与 CMDB/工单系统，对单条或批量告警自动完成
  业务归属链路还原、客户感知指标分析、关联告警聚合与分级处置建议生成。
  所有操作为只读，不涉及环境变更。Use when the user asks to assess alert business impact,
  generate incident reports, or evaluate Tencent Cloud alarms.
---

# 告警业务影响评估 Skill

## 适用场景

- 收到腾讯云监控告警、CMDB 事件或工单系统推送，需要快速评估业务影响
- 单条或批量告警需要统一输出《告警业务影响评估报告》与《分级处置建议与客户沟通话术》
- 值班/on-call 需要中性、非追责的客户沟通草稿

## 安全约束（必须遵守）

1. **只读操作**：禁止调用任何创建、修改、删除、重启类 API
2. **凭证最小化**：使用具备只读权限的子账号 SecretId/SecretKey 或 CAM 角色
3. **限流保护**：遵守 `docs/SKILL-DESIGN.md` 中的调用频次与退避策略
4. **数据脱敏**：报告中客户姓名、手机号等敏感信息仅展示末四位或代号

## 工作流

```
输入告警 → 资源识别 → 业务链路还原 → 指标采集(30min) → 关联聚合 → 预案匹配 → 分级建议 → 输出报告
```

### Step 1：解析告警输入

支持以下输入格式（可混合批量）：

| 格式 | 示例 |
|------|------|
| JSON 文件 | `alerts/sample-alerts.json` |
| 单条 JSON | `{"alarmId":"xxx","resourceType":"cvm","resourceId":"ins-xxx"}` |
| 告警 ID 列表 | `--alarm-ids alm-001,alm-002` |
| 自然语言描述 | "CVM ins-abc123 CPU 超过 90%" |

必填字段：`resourceType`（cvm/clb/cdb/cvm 等）、`resourceId` 或可从告警内容解析的资源标识。

### Step 2：调用评估引擎

在项目根目录执行（优先使用 mock 模式验证流程）：

```bash
# 构建（首次）
go build -o bin/alert-assess ./cmd/alert-assess/

# Mock 模式（无凭证，使用内置样例数据）
./bin/alert-assess assess --input alerts/sample-alerts.json --mock --output-dir reports/

# 真实只读 API（需配置环境变量）
export TENCENTCLOUD_SECRET_ID="xxx"
export TENCENTCLOUD_SECRET_KEY="xxx"
export TENCENTCLOUD_REGION="ap-guangzhou"
export CMDB_API_URL="https://cmdb.example.com/api"   # 可选
export CMDB_API_TOKEN="xxx"                           # 可选
./bin/alert-assess assess --input alerts/sample-alerts.json --output-dir reports/
```

### Step 3：业务链路还原

按以下顺序查询（详见 `docs/openapi-readonly-list.md`）：

1. **资源层**：`DescribeInstances` / `DescribeLoadBalancers` / `DescribeDBInstances`
2. **标签层**：`DescribeTags` → 提取 `BusinessSystem`、`Application`、`Owner`
3. **CMDB 层**：`GET /resources/{id}/lineage` → 补全应用→业务系统→客户责任人

输出 Mermaid 链路图写入评估报告。

### Step 4：客户感知指标（近 30 分钟）

从云监控拉取并计算趋势：

| 指标 | 命名空间 | 含义 |
|------|----------|------|
| QPS | QCE/LB_PUBLIC 或应用自定义 | 请求量 |
| 错误率 | 5xx/(2xx+5xx) | 客户可感知失败 |
| 响应时间 P99 | 毫秒 | 体验延迟 |

趋势判定：`rising`（上升>20%）、`stable`、`falling`。结合阈值给出感知等级：正常 / 轻微影响 / 明显影响 / 严重影响。

### Step 5：关联告警聚合

- 按 `resourceId`、`vpcId`、`BusinessSystem` 标签、时间窗口（±15min）聚类
- 与 `config/runbook-scenarios.yaml` 预案库匹配
- 输出：独立事件 / 已知问题（附预案 ID）/ 需进一步关联分析

### Step 6：分级处置建议

| 级别 | 条件（满足任一） | 建议动作 |
|------|------------------|----------|
| P1 | 错误率>5% 且上升；或多资源同业务系统；核心客户 | 立即介入 |
| P2 | 指标异常但未达 P1；单资源；有预案 | 观察 15min |
| P3 | 低峰期；非核心；已知维护窗口 | 可延后 |

### Step 7：生成输出物

引擎自动生成两份 Markdown 报告至 `--output-dir`：

1. `{timestamp}-impact-assessment-report.md` — 《告警业务影响评估报告》
2. `{timestamp}-handling-recommendation.md` — 《分级处置建议与客户沟通话术》

Agent 应阅读报告，按需补充人工判断说明，**不得**在报告中建议执行变更操作。

## 输入输出定义

### 输入

```json
{
  "alerts": [
    {
      "alarmId": "string",
      "alarmName": "string",
      "resourceType": "cvm|clb|cdb|rds",
      "resourceId": "string",
      "region": "ap-guangzhou",
      "severity": "critical|warning|info",
      "triggerTime": "2026-08-31T09:00:00+08:00",
      "metricName": "CpuUsage",
      "currentValue": 92.5,
      "threshold": 90
    }
  ],
  "options": {
    "metricWindowMinutes": 30,
    "correlationWindowMinutes": 15,
    "includeCustomerScript": true
  }
}
```

### 输出

| 文件 | 内容 |
|------|------|
| impact-assessment-report.md | 链路图、指标快照、关联聚合、感知判定 |
| handling-recommendation.md | P1/P2/P3 建议、观察要点、客户话术（中性表述） |

## 依赖与参考

- 设计说明：`docs/SKILL-DESIGN.md`
- 只读 API 清单：`docs/openapi-readonly-list.md`
- 预案场景库：`config/runbook-scenarios.yaml`
- 报告模板：`templates/`

## 故障处理

| 现象 | 处理 |
|------|------|
| API 限流 | 等待退避后重试，见 RateLimiter |
| CMDB 不可用 | 降级为标签链路，报告中标注「CMDB 未响应」 |
| 凭证缺失 | 使用 `--mock` 演示流程，提示用户配置只读子账号 |
| 批量>50 条 | 自动分批，每批间隔 2s |

## Agent 检查清单

- [ ] 确认使用只读 API，未调用写操作
- [ ] 确认输出两份报告
- [ ] 客户话术为中性、非追责表述
- [ ] 链路图与指标数据可追溯至 API 响应
- [ ] 标注 mock/真实数据来源
