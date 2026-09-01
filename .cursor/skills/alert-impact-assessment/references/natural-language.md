# 自然语言 → 告警 JSON 转换指南

Agent 收到自然语言告警描述时，按以下规则转为 `alerts/user-alerts.json`：

## 资源类型识别

| 关键词 | resourceType |
|--------|--------------|
| ECS、云服务器、i- 开头实例 | ecs |
| SLB、负载均衡、lb- 开头 | slb |
| RDS、数据库、rm- 开头 | rds |

## 字段映射

| 自然语言 | JSON 字段 |
|----------|-----------|
| 「告警 ID xxx」 | alarmId |
| 「CPU 92%」 | metricName: CPUUtilization, currentValue: 92 |
| 「延迟 45s」 | metricName: ReplicationDelay, currentValue: 45 |
| 「5xx 错误」 | metricName: HttpCode5xx |
| 未提及 region | region: cn-hangzhou-1 |
| 未提及 az | az: a, zoneId: cn-hangzhou-1-a |

## 示例

**输入**：「ECS i-order-svc-001 CPU 使用率 92%，阈值 90，严重告警」

**输出**：
```json
{
  "alerts": [{
    "alarmId": "alm-user-001",
    "alarmName": "ECS CPU 使用率过高",
    "resourceType": "ecs",
    "resourceId": "i-order-svc-001",
    "region": "cn-hangzhou-1",
    "zoneId": "cn-hangzhou-1-a",
    "severity": "critical",
    "metricName": "CPUUtilization",
    "currentValue": 92,
    "threshold": 90
  }]
}
```

写入 `alerts/user-alerts.json` 后执行 `./scripts/assess.sh --input alerts/user-alerts.json`。
