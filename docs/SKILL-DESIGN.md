# 告警业务影响评估 Skill — 设计说明文档

## 1. 概述

本 Skill 面向腾讯云运维/on-call 场景，在**只读**前提下自动完成告警的业务影响评估。智能体（Cursor Agent）加载 `.cursor/skills/alert-impact-assessment/SKILL.md` 后，按工作流调用本仓库 CLI 与 OpenAPI，输出两份标准化报告。

### 1.1 设计目标

| 能力 | 说明 |
|------|------|
| 业务链路还原 | 实例 → 应用 → 业务系统 → 客户责任人 |
| 客户感知判断 | QPS、错误率、P99 延迟近 30min 趋势 |
| 关联聚合 | 同源/同业务系统告警合并，匹配已知预案 |
| 分级建议 | P1/P2/P3 + 中性客户沟通话术 |

### 1.2 架构

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ 告警输入     │────▶│ alert_processor  │────▶│ report_generator│
│ JSON/CLI    │     │ 链路/指标/聚合    │     │ Markdown 报告    │
└─────────────┘     └────────┬─────────┘     └─────────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        ┌─────────┐   ┌──────────┐   ┌─────────┐
        │ ECS/CVM │   │ Monitor  │   │  CMDB   │
        │ SLB/RDS │   │ GetData  │   │ Lineage │
        └─────────┘   └──────────┘   └─────────┘
              腾讯云只读 OpenAPI          只读 HTTP GET
```

## 2. 输入输出定义

### 2.1 输入

**文件路径**：`alerts/*.json`

```json
{
  "alerts": [
    {
      "alarmId": "string, 必填",
      "alarmName": "string",
      "resourceType": "cvm|clb|cdb, 必填",
      "resourceId": "string, 必填",
      "region": "ap-guangzhou, 默认",
      "severity": "critical|warning|info",
      "triggerTime": "ISO8601",
      "metricName": "string",
      "currentValue": "number",
      "threshold": "number"
    }
  ],
  "options": {
    "metricWindowMinutes": 30,
    "correlationWindowMinutes": 15,
    "includeCustomerScript": true
  }
}
```

**CLI 参数**：

| 参数 | 说明 |
|------|------|
| `--input` | 告警 JSON 文件 |
| `--output-dir` | 报告目录，默认 `reports/` |
| `--mock` | 无凭证演示模式 |

**环境变量**（真实 API 模式）：

| 变量 | 说明 |
|------|------|
| `TENCENTCLOUD_SECRET_ID` | 只读子账号 SecretId |
| `TENCENTCLOUD_SECRET_KEY` | 只读子账号 SecretKey |
| `TENCENTCLOUD_REGION` | 默认地域 |
| `CMDB_API_URL` | CMDB 基址（可选） |
| `CMDB_API_TOKEN` | CMDB Bearer Token（可选） |

### 2.2 输出

| 文件 | 格式 | 内容 |
|------|------|------|
| `{ts}-impact-assessment-report.md` | Markdown | 链路 Mermaid 图、指标快照表、关联聚合、告警明细 |
| `{ts}-handling-recommendation.md` | Markdown | P1/P2/P3 处置建议、判定依据、客户话术草稿 |

## 3. 依赖只读 OpenAPI 清单

详见 [openapi-readonly-list.md](./openapi-readonly-list.md)。

### 3.1 核心接口

| 产品 | Action | 用途 |
|------|--------|------|
| CVM | DescribeInstances | 实例详情、VPC、标签 |
| CLB | DescribeLoadBalancers | 负载均衡详情、标签 |
| CDB | DescribeDBInstances | 数据库实例详情、标签 |
| Monitor | GetMonitorData | QPS/错误率/延迟时序 |
| Tag | DescribeTags | 补充业务标签（可选） |

### 3.2 CMDB / 工单（只读 HTTP）

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/resources/{type}/{id}/lineage` | 业务归属链路 |
| GET | `/tickets?alarmId=` | 关联工单（扩展） |

### 3.3 客户端只读防护

`TencentCloudClient.call()` 仅允许 Action 以 `Describe`、`Get`、`List`、`Search`、`Query` 开头，否则抛出异常。

## 4. 调用频次与限流策略

### 4.1 默认限流

| 层级 | 策略 | 参数 |
|------|------|------|
| 客户端 Token Bucket | 每服务域 | 20 次/秒 |
| 批量告警 | 分批处理 | 每批 ≤50 条，批间间隔 2s |
| API 429/限流响应 | 指数退避 | 1s → 2s → 4s → 8s，最多 4 次 |
| 并发 | 单地域串行 | 避免突发打满配额 |

### 4.2 单次评估调用估算

以 1 条告警为例（真实 API 模式）：

| 步骤 | API 调用次数 |
|------|-------------|
| 资源详情 | 1 |
| 监控指标（QPS/错误/延迟） | 3 |
| CMDB 链路 | 1 |
| **合计** | **≈5 次/告警** |

批量 10 条 ≈ 50 次，在默认 20 QPS 下约 3 秒完成（不含网络延迟）。

### 4.3 腾讯云官方配额参考

- CVM DescribeInstances：默认 40 次/秒（地域级）
- Monitor GetMonitorData：默认 20 次/秒
- 建议在 CAM 策略中绑定 `QcloudMonitorReadOnlyAccess` 等只读策略

### 4.4 降级策略

| 故障 | 降级行为 |
|------|----------|
| CMDB 超时 | 使用云标签链路，报告标注来源 |
| Monitor 无数据 | 指标列显示 `-`，感知等级降为「待确认」 |
| 凭证缺失 | 自动切换 Mock 模式 |
| 批量>50 | 自动分批，合并报告 |

## 5. 分级规则

| 级别 | 触发条件 | 处置 |
|------|----------|------|
| P1 | 错误率>5% 且上升；或 ≥3 条同源关联告警 | 立即介入，15min 同步 |
| P2 | 明显/轻微影响；匹配 P2 预案 | 观察 15min |
| P3 | 正常感知；匹配 P3 预案；非核心 | 可延后，常规巡检跟进 |

## 6. 预案匹配

配置文件：`config/runbook-scenarios.yaml`

匹配维度：`resourceType`、`metricName`、阈值区间、`BusinessSystem` 标签、关联告警数量。

## 7. 扩展指南

1. **新增产品**：在 `src/tencent_cloud/` 添加 `describe_*` 模块，更新 `METRIC_CONFIG`
2. **对接工单**：扩展 `CMDBClient.get_related_tickets(alarm_id)`
3. **自定义话术**：修改 `report_generator.CUSTOMER_SCRIPTS` 或按客户模板外置
4. **Webhook 触发**：包装 `python -m src.cli assess` 为 CI/告警回调入口

## 8. 安全与合规

- 所有 API 调用经 Action 前缀白名单校验
- 报告中联系人信息脱敏（`138****5678`）
- 不包含任何 Create/Modify/Delete/Restart 类操作
- 建议 CAM 策略：`QcloudCVMReadOnlyAccess`、`QcloudCLBReadOnlyAccess`、`QcloudCDBReadOnlyAccess`、`QcloudMonitorReadOnlyAccess`
