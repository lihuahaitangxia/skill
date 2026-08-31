# 腾讯云只读 OpenAPI 清单

> 本 Skill 仅调用以下只读接口。Action 名称以 `Describe` / `Get` / `List` / `Query` / `Search` 开头。

## CVM（云服务器）

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| DescribeInstances | 2017-03-12 | 实例详情、IP、VPC、状态、标签 | InstanceIds |
| DescribeInstanceStatus | 2017-03-12 | 实例运行状态 | InstanceIds |

**Endpoint**: `cvm.tencentcloudapi.com`

## CLB（负载均衡）

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| DescribeLoadBalancers | 2018-03-17 | LB 详情、VPC、状态、标签 | LoadBalancerIds |
| DescribeListeners | 2018-03-17 | 监听器配置 | LoadBalancerId |
| DescribeTargets | 2018-03-17 | 后端主机健康状态 | LoadBalancerId |

**Endpoint**: `clb.tencentcloudapi.com`

## CDB（云数据库 MySQL）

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| DescribeDBInstances | 2017-03-20 | 实例详情、引擎、VPC、标签 | InstanceIds |
| DescribeDBInstanceInfo | 2017-03-20 | 实例运行信息 | InstanceId |

**Endpoint**: `cdb.tencentcloudapi.com`

## 云监控 Monitor

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| GetMonitorData | 2018-07-24 | 时序指标（QPS/错误/延迟） | Namespace, MetricName, Instances, StartTime, EndTime |
| DescribeAlarmHistories | 2018-07-24 | 历史告警（关联分析扩展） | Module, StartTime, EndTime |
| DescribeAlarmPolicies | 2018-07-24 | 告警策略详情 | Module, PolicyIds |

**Endpoint**: `monitor.tencentcloudapi.com`

### 常用监控命名空间与指标

| 资源 | Namespace | QPS 指标 | 错误指标 | 延迟指标 |
|------|-----------|----------|----------|----------|
| CVM | QCE/CVM | AccOuttraffic | CpuUsage* | MemUsage* |
| CLB | QCE/LB_PUBLIC | ClientConnum | Http5xx | RspAvg |
| CDB | QCE/CDB | Qps | SlowQueries | QueryLatency |

\* CVM 默认指标为资源利用率代理；生产环境建议对接应用层自定义上报。

## 标签 Tag

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| DescribeTags | 2018-08-13 | 资源标签列表 | ServiceType, ResourcePrefix, ResourceIds |

**Endpoint**: `tag.tencentcloudapi.com`

## CMDB / 工单系统（HTTP 只读）

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/resources/{type}/{id}/lineage` | 业务归属链路 |
| GET | `/resources/{type}/{id}/tags` | 扩展标签 |
| GET | `/tickets?status=open&resourceId=` | 进行中工单 |
| GET | `/runbooks/{id}` | 预案详情 |

## CAM 只读策略建议

```json
{
  "version": "2.0",
  "statement": [
    {
      "effect": "allow",
      "action": [
        "cvm:DescribeInstances",
        "cvm:DescribeInstanceStatus",
        "clb:DescribeLoadBalancers",
        "clb:DescribeListeners",
        "clb:DescribeTargets",
        "cdb:DescribeDBInstances",
        "monitor:GetMonitorData",
        "monitor:DescribeAlarmHistories",
        "monitor:DescribeAlarmPolicies",
        "tag:DescribeTags"
      ],
      "resource": "*"
    }
  ]
}
```

或使用腾讯云预设策略：
- `QcloudCVMReadOnlyAccess`
- `QcloudCLBReadOnlyAccess`
- `QcloudCDBReadOnlyAccess`
- `QcloudMonitorReadOnlyAccess`

## 调用频次建议

| 场景 | 频率 | 说明 |
|------|------|------|
| 单条告警实时评估 | 即时 1 次 | ≈5 个 API 调用 |
| 批量告警（≤50） | 事件驱动 | 批间 2s 间隔 |
| 定时巡检补充 | 每 5min | 仅 P2 观察中的告警 |
| CMDB 链路缓存 | TTL 10min | 减少重复查询 |

## 限流与退避

1. 客户端 `RateLimiter`：20 次/秒/token bucket
2. 收到 HTTP 429 或 `RequestLimitExceeded`：指数退避 1/2/4/8 秒
3. 单告警评估超时 60s，整体批量超时 300s
