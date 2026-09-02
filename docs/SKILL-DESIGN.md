# 告警业务影响评估 Skill — 设计说明文档

## 1. 概述

本 Skill 面向 **Aliyun** 运维/on-call 场景，在**只读**前提下自动完成告警的业务影响评估。智能体（Cursor Agent）加载 `.cursor/skills/alert-impact-assessment/SKILL.md` 后，按工作流调用本仓库 CLI 与 POP OpenAPI，输出两份标准化报告。

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
        │ ECS/SLB │   │   CMS    │   │  CMDB   │
        │   RDS   │   │ MetricList│  │ Lineage │
        └─────────┘   └──────────┘   └─────────┘
           Aliyun POP 只读 OpenAPI    只读 HTTP GET
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
      "resourceType": "ecs|slb|rds, 必填",
      "resourceId": "string, 必填",
      "region": "cn-hangzhou-1, 默认",
      "zoneId": "cn-hangzhou-1-a, 可选",
      "az": "a, 可用区后缀，可选",
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
| `ALIYUN_ACCESS_KEY_ID` | RAM 只读 AccessKey ID |
| `ALIYUN_ACCESS_KEY_SECRET` | RAM 只读 AccessKey Secret |
| `ALIYUN_REGION` | 地域 ID |
| `ALIYUN_AZ` | 可用区后缀（默认 `a`，完整 ZoneId 为 `cn-hangzhou-1-a`） |
| `ALIYUN_ECS_ENDPOINT` | ECS POP Endpoint |
| `ALIYUN_SLB_ENDPOINT` | SLB POP Endpoint |
| `ALIYUN_RDS_ENDPOINT` | RDS POP Endpoint |
| `ALIYUN_CMS_ENDPOINT` | 云监控 POP Endpoint |
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
| ECS | DescribeInstances | 实例详情、VPC、标签 |
| SLB | DescribeLoadBalancerAttribute | 负载均衡详情、标签 |
| RDS | DescribeDBInstanceAttribute | 数据库实例详情、标签 |
| CMS | DescribeMetricList | QPS/错误率/延迟时序 |
| Tag | ListTagResources | 补充业务标签（可选） |

### 3.2 CMDB / 工单（只读 HTTP）

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/resources/{type}/{id}/lineage` | 业务归属链路 |
| GET | `/tickets?alarmId=` | 关联工单（扩展） |

### 3.3 客户端只读防护

`Client.CallRPC()` 仅允许 Action 以 `Describe`、`Get`、`List`、`Search`、`Query` 开头，否则抛出异常。

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

### 4.3 Aliyun 配额参考

- Endpoint 与 QPS 限制因 Stack 版本与部署规模而异
- 建议在 RAM 策略中绑定 `AliyunECSReadOnlyAccess`、`AliyunSLBReadOnlyAccess`、`AliyunRDSReadOnlyAccess`、`AliyunCloudMonitorReadOnlyAccess`

### 4.4 降级策略

| 故障 | 降级行为 |
|------|----------|
| CMDB 超时 | 使用云标签链路，报告标注来源 |
| CMS 无数据 | 指标列显示 `-`，感知等级降为「待确认」 |
| 凭证/Endpoint 缺失 | 自动切换 Mock 模式 |
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

1. **新增产品**：在 `internal/aliyun/` 添加模块，更新 `monitor.go` 中的 `metricConfigs`
2. **对接工单**：扩展 `CMDBClient.GetRelatedTickets(alarmID)`
3. **自定义话术**：修改 `internal/report/generator.go` 中的 `customerScripts`
4. **Webhook 触发**：包装 `./bin/alert-assess assess` 为 CI/告警回调入口

## 8. 安全与合规

- 所有 API 调用经 Action 前缀白名单校验
- 报告中联系人信息脱敏（`138****5678`）
- 不包含任何 Create/Modify/Delete/Restart 类操作
- 建议 RAM 策略：`AliyunECSReadOnlyAccess`、`AliyunSLBReadOnlyAccess`、`AliyunRDSReadOnlyAccess`、`AliyunCloudMonitorReadOnlyAccess`
