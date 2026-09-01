# 输入输出定义

详见项目根目录 [docs/SKILL-DESIGN.md](../../../docs/SKILL-DESIGN.md) 第 2 节。

## 输入

- 文件：`alerts/*.json`
- 必填：`alarmId`, `resourceType`（ecs/slb/rds）, `resourceId`
- 可选：`region`（默认 cn-hangzhou-1）, `zoneId`, `az`

## 输出

- `deliverables/01-告警业务影响评估报告.md`
- `deliverables/02-分级处置建议与客户沟通话术.md`

## CLI

```bash
./scripts/assess.sh --mock
./scripts/assess.sh --input alerts/your-alerts.json
```
