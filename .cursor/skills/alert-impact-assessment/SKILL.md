---
name: alert-impact-assessment
description: >-
  告警业务影响评估 Skill。通过 Apsara Stack 只读 POP OpenAPI 与 CMDB/工单系统，对单条或批量告警自动完成
  业务归属链路还原、客户感知指标分析、关联告警聚合与分级处置建议生成，并输出两份标准交付报告。
  所有操作为只读，不涉及环境变更。
  Use when the user asks to assess alert business impact, generate incident reports,
  evaluate Apsara Stack alarms, or produce P1/P2/P3 handling recommendations.
---

# 告警业务影响评估 Skill

## 何时使用

- 收到 Apsara Stack 云监控告警，需要评估业务影响
- 需要输出《告警业务影响评估报告》和《分级处置建议与客户沟通话术》
- 需要 P1/P2/P3 分级建议及中性、非追责的客户沟通草稿

## 安全约束（必须遵守）

1. **只读操作** — 禁止 Create/Modify/Delete/Restart 类 API
2. **RAM 最小权限** — 只读 AccessKey
3. **Endpoint 配置** — POP Endpoint 从 Apsara Uni-manager 服务注册变量获取
4. **限流** — 见 [rate-limiting.md](references/rate-limiting.md)
5. **脱敏** — 联系人仅展示末四位

## 工作流（Agent 必须按序执行）

```
解析告警 → 运行评估 CLI → 同步交付物 → 向用户展示结果
```

### Step 1：解析告警

| 输入方式 | 示例 |
|----------|------|
| JSON 文件 | `alerts/sample-alerts.json` |
| 用户粘贴 | 单条或多条告警 JSON |
| 自然语言 | 「ECS i-xxx CPU 超 90%」→ 转为 JSON |

必填：`resourceType`（ecs/slb/rds）、`resourceId`

若用户未提供 JSON，创建临时文件 `alerts/user-alerts.json`。

### Step 2：运行评估引擎

```bash
chmod +x scripts/assess.sh
./scripts/assess.sh --mock                          # 无凭证演示
./scripts/assess.sh --input alerts/user-alerts.json # 真实/自定义输入
```

真实 API 需先配置环境变量（见 [openapi-readonly.md](references/openapi-readonly.md)）。

### Step 3：确认交付物

运行后必须存在：

| 文件 | 内容 |
|------|------|
| `deliverables/01-告警业务影响评估报告.md` | 链路 Mermaid 图、30min 指标快照、关联聚合 |
| `deliverables/02-分级处置建议与客户沟通话术.md` | P1/P2/P3 建议 + 中性话术 |

向用户展示报告摘要，并告知完整文件路径。

### Step 4：能力说明（按需）

- 业务链路：实例 → 应用 → 业务系统 → 客户 → 责任人
- 客户感知：QPS、错误率、P99 近 30min 趋势
- 关联聚合：同 BusinessSystem + VPC 聚类，匹配 `config/runbook-scenarios.yaml`

## 参考文档

| 主题 | 文件 |
|------|------|
| 输入输出 | [references/input-output.md](references/input-output.md) |
| 只读 OpenAPI | [references/openapi-readonly.md](references/openapi-readonly.md) |
| 限流策略 | [references/rate-limiting.md](references/rate-limiting.md) |
| 完整设计 | [deliverables/03-Skill设计说明文档.md](../../deliverables/03-Skill设计说明文档.md) |

## Agent 检查清单

- [ ] 使用 `./scripts/assess.sh` 或等效 CLI，未手动编造指标
- [ ] 确认只读 API，未调用写操作
- [ ] 输出两份 deliverables 报告
- [ ] 客户话术为中性、非追责表述
- [ ] 标注 mock / 真实数据来源
