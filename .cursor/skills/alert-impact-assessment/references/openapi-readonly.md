# 只读 OpenAPI 清单

详见 [docs/openapi-readonly-list.md](../../../docs/openapi-readonly-list.md)。

## 核心接口

| 产品 | Action | Version |
|------|--------|---------|
| ECS | DescribeInstances | 2014-05-26 |
| SLB | DescribeLoadBalancerAttribute | 2014-05-15 |
| RDS | DescribeDBInstanceAttribute | 2014-08-15 |
| CMS | DescribeMetricList | 2019-01-01 |

## 环境变量

```
APSARASTACK_ACCESS_KEY_ID
APSARASTACK_ACCESS_KEY_SECRET
APSARASTACK_REGION=cn-hangzhou-1
APSARASTACK_AZ=a
APSARASTACK_ECS_ENDPOINT
APSARASTACK_SLB_ENDPOINT
APSARASTACK_RDS_ENDPOINT
APSARASTACK_CMS_ENDPOINT
```

## 只读防护

`CallRPC()` 仅允许 Action 以 Describe/Get/List/Query/Search 开头。
