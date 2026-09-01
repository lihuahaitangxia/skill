---
name: alert-impact-assessment
description: >-
  告警业务影响评估 Skill。通过 Apsara Stack 只读 POP OpenAPI 与 CMDB/工单系统，对单条或批量告警自动完成
  业务归属链路还原、客户感知指标分析、关联告警聚合与分级处置建议生成。
  所有操作为只读，不涉及环境变更。Use when the user asks to assess alert business impact,
  generate incident reports, or evaluate Apsara Stack / Alibaba Cloud alarms.
---

# 告警业务影响评估 Skill

## 适用场景

- 收到 Apsara Stack 云监控告警、CMDB 事件或工单系统推送，需要快速评估业务影响
- 单条或批量告警需要统一输出《告警业务影响评估报告》与《分级处置建议与客户沟通话术》
- 值班/on-call 需要中性、非追责的客户沟通草稿

## 安全约束（必须遵守）

1. **只读操作**：禁止调用任何创建、修改、删除、重启类 API
2. **凭证最小化**：使用具备只读权限的 RAM 子账号 AccessKey
3. **Endpoint 配置**：POP Endpoint 必须从 Apsara Uni-manager 服务注册变量获取
4. **限流保护**：遵守 `docs/SKILL-DESIGN.md` 中的调用频次与退避策略
5. **数据脱敏**：报告中客户姓名、手机号等敏感信息仅展示末四位或代号

## 工作流

```
输入告警 → 资源识别 → 业务链路还原 → 指标采集(30min) → 关联聚合 → 预案匹配 → 分级建议 → 输出报告
```

### Step 1：解析告警输入

| 格式 | 示例 |
|------|------|
| JSON 文件 | `alerts/sample-alerts.json` |
| 单条 JSON | `{"alarmId":"xxx","resourceType":"ecs","resourceId":"i-xxx"}` |
| 自然语言描述 | "ECS i-abc123 CPU 超过 90%" |

必填字段：`resourceType`（ecs/slb/rds，兼容 cvm/clb/cdb 别名）、`resourceId`。

### Step 2：调用评估引擎

```bash
go build -o bin/alert-assess ./cmd/alert-assess/

# Mock 模式
./bin/alert-assess assess --input alerts/sample-alerts.json --mock --output-dir reports/

# 真实只读 API
export APSARASTACK_ACCESS_KEY_ID="xxx"
export APSARASTACK_ACCESS_KEY_SECRET="xxx"
export APSARASTACK_REGION="cn-hangzhou-1"
export APSARASTACK_AZ="a"
export APSARASTACK_ECS_ENDPOINT="ecs.cn-hangzhou-1.xxx.stack.local"
export APSARASTACK_SLB_ENDPOINT="slb.cn-hangzhou-1.xxx.stack.local"
export APSARASTACK_RDS_ENDPOINT="rds.cn-hangzhou-1.xxx.stack.local"
export APSARASTACK_CMS_ENDPOINT="metrics.cn-hangzhou-1.xxx.stack.local"
./bin/alert-assess assess --input alerts/sample-alerts.json --output-dir reports/
```

### Step 3：业务链路还原

1. **资源层**：`DescribeInstances` / `DescribeLoadBalancerAttribute` / `DescribeDBInstanceAttribute`
2. **标签层**：资源 Tags → 提取 `BusinessSystem`、`Application`、`Owner`
3. **CMDB 层**：`GET /resources/{id}/lineage`

### Step 4：客户感知指标（近 30 分钟）

| 指标 | 命名空间 | 含义 |
|------|----------|------|
| QPS | acs_slb_dashboard 等 | 请求量 |
| 错误率 | HttpCode5xx / SlowQueries 等 | 客户可感知失败 |
| 响应时间 | Latency 等 | 体验延迟 |

### Step 5～7

关联聚合 → 预案匹配 → 生成两份 Markdown 报告（详见 `docs/SKILL-DESIGN.md`）。

## Agent 检查清单

- [ ] 确认使用只读 API，未调用写操作
- [ ] 确认 POP Endpoint 已配置（非 Mock 模式）
- [ ] 确认输出两份报告
- [ ] 客户话术为中性、非追责表述
