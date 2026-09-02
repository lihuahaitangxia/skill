# Aliyun 只读 OpenAPI 清单

> 本 Skill 通过 Aliyun POP 网关调用以下只读接口。Action 名称以 `Describe` / `Get` / `List` / `Query` / `Search` 开头。

## 认证与 Endpoint

Aliyun 使用 **AccessKey + HMAC-SHA1** 签名（POP RPC 风格），Endpoint 需从 **阿里云控制台 运维控制台 → 服务注册变量** 获取，或通过现场运维工程师提供。

| 环境变量 | 说明 |
|----------|------|
| `ALIYUN_ACCESS_KEY_ID` | RAM 只读子账号 AccessKey ID |
| `ALIYUN_ACCESS_KEY_SECRET` | RAM 只读子账号 AccessKey Secret |
| `ALIYUN_REGION` | 地域 ID，如 `cn-hangzhou-1` |
| `ALIYUN_AZ` | 可用区后缀，如 `a`（完整 ZoneId 为 `cn-hangzhou-1-a`） |
| `ALIYUN_ECS_ENDPOINT` | ECS POP Endpoint |
| `ALIYUN_SLB_ENDPOINT` | SLB POP Endpoint |
| `ALIYUN_RDS_ENDPOINT` | RDS POP Endpoint |
| `ALIYUN_CMS_ENDPOINT` | 云监控 POP Endpoint |
| `ALIYUN_POP_ENDPOINT` | 统一 POP 网关（可选，作为各服务 fallback） |

也支持标准别名：`ALIBABA_CLOUD_ACCESS_KEY_ID`、`ALIBABA_CLOUD_ACCESS_KEY_SECRET`、`ALIBABA_CLOUD_REGION`。

## ECS（弹性计算）

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| DescribeInstances | 2014-05-26 | 实例详情、VPC、标签 | InstanceIds |
| DescribeInstanceStatus | 2014-05-26 | 实例运行状态 | InstanceIds |

**Endpoint 变量**: `ALIYUN_ECS_ENDPOINT`

## SLB（负载均衡）

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| DescribeLoadBalancerAttribute | 2014-05-15 | LB 详情、VPC、标签 | LoadBalancerId |
| DescribeLoadBalancerAttribute | 2014-05-15 | 后端健康状态 | LoadBalancerId |
| DescribeLoadBalancerListeners | 2014-05-15 | 监听器配置 | LoadBalancerId |

**Endpoint 变量**: `ALIYUN_SLB_ENDPOINT`

## RDS（云数据库）

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| DescribeDBInstanceAttribute | 2014-08-15 | 实例详情、引擎、VPC、标签 | DBInstanceId |
| DescribeDBInstancePerformance | 2014-08-15 | 实例性能（扩展） | DBInstanceId |
| DescribeReadDBInstanceDelay | 2014-08-15 | 只读实例延迟 | DBInstanceId |

**Endpoint 变量**: `ALIYUN_RDS_ENDPOINT`

## 云监控 CMS

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| DescribeMetricList | 2019-01-01 | 时序指标（QPS/错误/延迟） | Namespace, MetricName, Dimensions, StartTime, EndTime |
| DescribeMetricLast | 2019-01-01 | 最新数据点 | Namespace, MetricName, Dimensions |
| DescribeAlertLogList | 2019-01-01 | 历史告警（关联分析扩展） | — |

**Endpoint 变量**: `ALIYUN_CMS_ENDPOINT`

### 常用监控命名空间与指标

| 资源 | Namespace | QPS 指标 | 错误指标 | 延迟指标 |
|------|-----------|----------|----------|----------|
| ECS | acs_ecs_dashboard | InternetInRate | CPUUtilization* | IntranetOutRate* |
| SLB | acs_slb_dashboard | Qps | HttpCode5xx | Latency |
| RDS | acs_rds_dashboard | MySQL_QPS | SlowQueries | MySQL_NetworkTraffic |

\* ECS 默认指标为资源利用率/流量代理；生产环境建议对接应用层自定义上报。

## 标签 Tag

| Action | Version | 用途 | 关键参数 |
|--------|---------|------|----------|
| ListTagResources | 2018-08-28 | 资源标签列表 | ResourceType, ResourceId |

## CMDB / 工单系统（HTTP 只读）

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/resources/{type}/{id}/lineage` | 业务归属链路 |
| GET | `/resources/{type}/{id}/tags` | 扩展标签 |
| GET | `/tickets?status=open&resourceId=` | 进行中工单 |
| GET | `/runbooks/{id}` | 预案详情 |

## RAM 只读策略建议

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:DescribeInstances",
        "ecs:DescribeInstanceStatus",
        "slb:DescribeLoadBalancerAttribute",
        "slb:DescribeLoadBalancerListeners",
        "rds:DescribeDBInstanceAttribute",
        "rds:DescribeDBInstancePerformance",
        "cms:DescribeMetricList",
        "cms:DescribeMetricLast",
        "cms:DescribeAlertLogList",
        "tag:ListTagResources"
      ],
      "Resource": "*"
    }
  ]
}
```

或使用系统策略：
- `AliyunECSReadOnlyAccess`
- `AliyunSLBReadOnlyAccess`
- `AliyunRDSReadOnlyAccess`
- `AliyunCloudMonitorReadOnlyAccess`

## 调用频次建议

| 场景 | 频率 | 说明 |
|------|------|------|
| 单条告警实时评估 | 即时 1 次 | ≈5 个 API 调用 |
| 批量告警（≤50） | 事件驱动 | 批间 2s 间隔 |
| 定时巡检补充 | 每 5min | 仅 P2 观察中的告警 |
| CMDB 链路缓存 | TTL 10min | 减少重复查询 |

## 限流与退避

1. 客户端 `RateLimiter`：20 次/秒/token bucket
2. 收到 HTTP 429 或 `Throttling`：指数退避 1/2/4/8 秒
3. 单告警评估超时 60s，整体批量超时 300s
