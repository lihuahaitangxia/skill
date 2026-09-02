# 告警业务影响评估 Skill — 设计说明文档

> 版本：Aliyun 专版  
> 云厂商：Aliyun（POP 只读 OpenAPI）  
> 实现：Go CLI + Cursor Skill  
> 安全约束：**所有操作为只读，不涉及环境变更**

---

## 1. 概述

本 Skill 面向 Aliyun 运维/on-call 场景，自动完成：

1. 告警涉及资源的**业务归属链路还原**（实例 → 应用 → 业务系统 → 客户责任人）
2. **实时业务流量与客户感知判断**（QPS、错误率、响应时间近 30min 趋势）
3. **同源/关联告警聚合分析**，匹配已知预案
4. **分级处置建议**（P1/P2/P3）与**客户沟通话术草稿**（中性、非追责）

### 1.1 架构

```
告警 JSON → processor → aliyun POP API（ECS/SLB/RDS/CMS）
                ↓
         report generator → 两份 Markdown 报告
                ↓
              CMDB（只读 HTTP GET）
```

### 1.2 代码与 Skill 位置

| 组件 | 路径 |
|------|------|
| CLI 入口 | `cmd/alert-assess/main.go` |
| 告警处理 | `internal/processor/processor.go` |
| 报告生成 | `internal/report/generator.go` |
| API 客户端 | `internal/aliyun/` |
| Cursor Skill | `.cursor/skills/alert-impact-assessment/SKILL.md` |
| 预案库 | `config/runbook-scenarios.yaml` |

---

## 2. 输入输出定义

### 2.1 输入

**文件格式**：`alerts/*.json`

```json
{
  "alerts": [
    {
      "alarmId": "string, 必填",
      "alarmName": "string",
      "resourceType": "ecs|slb|rds, 必填",
      "resourceId": "string, 必填",
      "region": "cn-hangzhou-1",
      "zoneId": "cn-hangzhou-1-a, 可选",
      "az": "a, 可选",
      "severity": "critical|warning|info",
      "triggerTime": "ISO8601",
      "metricName": "string",
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

**CLI 用法**：

```bash
go build -o bin/alert-assess ./cmd/alert-assess/
./bin/alert-assess assess --input alerts/sample-alerts.json --mock --output-dir reports/
```

| 参数 | 说明 |
|------|------|
| `--input` / `-i` | 告警 JSON 文件（必填） |
| `--output-dir` / `-o` | 报告输出目录，默认 `reports/` |
| `--mock` | 无凭证演示模式 |

**环境变量**（真实 API 模式）：

| 变量 | 说明 |
|------|------|
| `ALIYUN_ACCESS_KEY_ID` | RAM 只读 AccessKey ID |
| `ALIYUN_ACCESS_KEY_SECRET` | RAM 只读 AccessKey Secret |
| `ALIYUN_REGION` | 地域 ID，默认 `cn-hangzhou-1` |
| `ALIYUN_AZ` | 可用区后缀，默认 `a`（完整 ZoneId: `cn-hangzhou-1-a`） |
| `ALIYUN_ECS_ENDPOINT` | ECS POP Endpoint |
| `ALIYUN_SLB_ENDPOINT` | SLB POP Endpoint |
| `ALIYUN_RDS_ENDPOINT` | RDS POP Endpoint |
| `ALIYUN_CMS_ENDPOINT` | 云监控 POP Endpoint |
| `CMDB_API_URL` | CMDB 基址（可选） |
| `CMDB_API_TOKEN` | CMDB Bearer Token（可选） |

### 2.2 输出

| 文件 | 内容 |
|------|------|
| `{ts}-impact-assessment-report.md` | 《告警业务影响评估报告》 |
| `{ts}-handling-recommendation.md` | 《分级处置建议与客户沟通话术》 |

固定交付目录：`deliverables/`（01/02/03 编号文件）

---

## 3. 依赖只读 OpenAPI 清单

> 通过 Aliyun POP 网关调用，Action 前缀限定：`Describe` / `Get` / `List` / `Query` / `Search`

### 3.1 ECS（弹性计算）

| Action | Version | 用途 |
|--------|---------|------|
| DescribeInstances | 2014-05-26 | 实例详情、VPC、标签 |
| DescribeInstanceStatus | 2014-05-26 | 实例运行状态 |

### 3.2 SLB（负载均衡）

| Action | Version | 用途 |
|--------|---------|------|
| DescribeLoadBalancerAttribute | 2014-05-15 | LB 详情、VPC、标签 |
| DescribeLoadBalancerListeners | 2014-05-15 | 监听器配置 |

### 3.3 RDS（云数据库）

| Action | Version | 用途 |
|--------|---------|------|
| DescribeDBInstanceAttribute | 2014-08-15 | 实例详情、引擎、VPC、标签 |
| DescribeReadDBInstanceDelay | 2014-08-15 | 只读实例延迟 |

### 3.4 CMS（云监控）

| Action | Version | 用途 |
|--------|---------|------|
| DescribeMetricList | 2019-01-01 | QPS/错误率/延迟时序（近 30min） |
| DescribeMetricLast | 2019-01-01 | 最新数据点 |
| DescribeAlertLogList | 2019-01-01 | 历史告警（关联分析扩展） |

**监控命名空间**：

| 资源 | Namespace | QPS | 错误率 | 延迟 |
|------|-----------|-----|--------|------|
| ECS | acs_ecs_dashboard | InternetInRate | CPUUtilization | IntranetOutRate |
| SLB | acs_slb_dashboard | Qps | HttpCode5xx | Latency |
| RDS | acs_rds_dashboard | MySQL_QPS | SlowQueries | MySQL_NetworkTraffic |

### 3.5 CMDB / 工单（HTTP 只读）

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/resources/{type}/{id}/lineage` | 业务归属链路 |
| GET | `/tickets?status=open&resourceId=` | 进行中工单 |

### 3.6 只读防护

`Client.CallRPC()` 硬编码 Action 前缀白名单，非只读 Action 直接拒绝。

### 3.7 RAM 只读策略

- `AliyunECSReadOnlyAccess`
- `AliyunSLBReadOnlyAccess`
- `AliyunRDSReadOnlyAccess`
- `AliyunCloudMonitorReadOnlyAccess`

---

## 4. 调用频次与限流策略

### 4.1 默认限流

| 层级 | 策略 | 参数 |
|------|------|------|
| 客户端 Token Bucket | 全局 | **20 次/秒** |
| 批量告警 | 分批 | 每批 **≤50 条**，批间 **2s** |
| HTTP 429 / Throttling | 指数退避 | **1s → 2s → 4s → 8s**，最多 4 次 |
| 并发 | 单地域 | 串行，避免突发打满配额 |
| 超时 | 单告警 / 批量 | **60s / 300s** |

### 4.2 单次评估 API 调用估算

| 步骤 | 次数 |
|------|------|
| 资源详情（ECS/SLB/RDS） | 1 |
| 监控指标（QPS + 错误率 + 延迟） | 3 |
| CMDB 链路 | 1 |
| **合计** | **≈5 次/告警** |

批量 10 条 ≈ 50 次 API 调用，默认 20 QPS 下约 3 秒完成（不含网络延迟）。

### 4.3 调用频次建议

| 场景 | 频率 |
|------|------|
| 单条告警实时评估 | 事件驱动，即时 1 次 |
| 批量告警（≤50） | 事件驱动，批间 2s |
| P2 观察中告警补充 | 每 5min |
| CMDB 链路缓存 | TTL 10min |

### 4.4 降级策略

| 故障 | 降级行为 |
|------|----------|
| CMDB 不可用 | 降级为云标签链路，报告标注来源 |
| CMS 无数据 | 指标显示 `-`，感知等级「待确认」 |
| 凭证/Endpoint 缺失 | 自动切换 Mock 模式 |
| 批量 > 50 | 自动分批，合并报告 |

---

## 5. 分级规则与预案

| 级别 | 条件 | 处置 |
|------|------|------|
| P1 | 错误率>5% 且上升；或 ≥3 条同源关联 | 立即介入 |
| P2 | 明显/轻微影响；匹配 P2 预案 | 观察 15min |
| P3 | 正常；匹配 P3 预案 | 可延后 |

预案配置：`config/runbook-scenarios.yaml`（RB-001 ~ RB-005）

---

## 6. 安全与合规

- 禁止 Create / Modify / Delete / Restart 类 API
- 报告联系人脱敏（`138****5678`）
- POP Endpoint 从 阿里云控制台 服务注册变量获取
- 建议使用 RAM 只读子账号，最小权限原则
