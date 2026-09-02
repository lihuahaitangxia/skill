---
name: alert-impact-assessment
description: >-
  告警业务影响评估 Skill（aliyun 专版）。对单条或批量告警自动完成业务链路还原、
  客户感知指标分析、关联告警聚合、P1/P2/P3 分级建议与客户沟通话术生成。
  只读操作，不变更环境。触发词：告警评估、业务影响、P1/P2/P3、aliyun 告警。
---

# 告警业务影响评估 Skill

## 何时触发

- 评估告警业务影响 / 生成影响评估报告
- 需要 P1/P2/P3 分级建议或客户沟通话术
- 分析 ECS/SLB/RDS 告警关联与业务链路

## 安全约束

1. **只读** — 禁止 Create/Modify/Delete/Restart
2. **RAM 最小权限** — 只读 AccessKey
3. **Endpoint** — 从阿里云控制台或 RAM 配置获取（默认 `*.aliyuncs.com`）
4. **脱敏** — 联系人末四位

## 工作流

```
进入 Skill 根目录 → 预检 → 解析告警 → 运行 CLI → 展示摘要 + 交付物路径
```

**AI Studio 独立包**：解压或导入后，所有命令在 `alert-impact-assessment/` 目录下执行（含 `SKILL.md` 的目录）。

```bash
cd alert-impact-assessment
```

### Step 0：环境预检

```bash
./scripts/assess.sh --validate
```

- `ready: true, mode: mock` → 使用 `--mock`
- `ready: true, mode: readonly` → 真实 API
- `ready: false` → 提示用户配置缺失的 Endpoint，仍可用 `--mock`

### Step 1：解析告警

| 输入 | 处理 |
|------|------|
| JSON 文件 | 直接使用 |
| 用户粘贴 JSON | 写入 `alerts/user-alerts.json` |
| 自然语言 | 按 [natural-language.md](references/natural-language.md) 转 JSON |

模板：[alert-template.json](references/alert-template.json)

### Step 2：运行评估

```bash
./scripts/assess.sh --input alerts/user-alerts.json --mock --summary
./scripts/assess.sh --input alerts/user-alerts.json --summary   # 真实 API
```

### Step 3：向用户汇报

必须包含：

1. **JSON 摘要**（`--summary` 输出）：P1/P2/P3 数量、建议
2. **报告路径**：
   - `deliverables/01-告警业务影响评估报告.md`
   - `deliverables/02-分级处置建议与客户沟通话术.md`
3. **关键发现**：最高优先级、关联聚合结论、客户感知
4. **数据来源**：mock / aliyun_readonly

**禁止**手动编造指标数据，必须来自 CLI 输出。

## 参考文档

| 主题 | 文件 |
|------|------|
| 输入输出 | [references/input-output.md](references/input-output.md) |
| 自然语言解析 | [references/natural-language.md](references/natural-language.md) |
| 只读 OpenAPI | [references/openapi-readonly.md](references/openapi-readonly.md) |
| 限流策略 | [references/rate-limiting.md](references/rate-limiting.md) |
| 其他平台集成 | [references/platform-integration.md](references/platform-integration.md) |
| 故障排查 | [references/troubleshooting.md](references/troubleshooting.md) |
| 完整设计 | [docs/SKILL-DESIGN.md](docs/SKILL-DESIGN.md) |

## Agent 检查清单

- [ ] 已运行 `validate` 或 `assess.sh`
- [ ] 未手动编造 QPS/错误率/延迟数据
- [ ] 已输出 deliverables 两份报告路径
- [ ] 已展示 P1/P2/P3 摘要
- [ ] 客户话术为中性、非追责表述
- [ ] 标注 mock / 真实数据来源
