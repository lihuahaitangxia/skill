"""Markdown report generation."""

from __future__ import annotations

from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from .alert_processor import ProcessedAlert, correlate_alerts


def _mermaid_lineage(items: list[ProcessedAlert]) -> str:
    lines = ["```mermaid", "flowchart LR"]
    seen: set[str] = set()
    for item in items:
        inst = item.lineage.get("instance") or item.alert.resource_id
        app = item.lineage.get("application") or "未知应用"
        bs = item.lineage.get("businessSystem") or "未知业务系统"
        cust = item.lineage.get("customer") or "未知客户"
        owner = item.lineage.get("owner") or "待确认"

        for node_id, label in [
            (f"inst_{inst}", f"实例\\n{inst}"),
            (f"app_{app}", f"应用\\n{app}"),
            (f"bs_{bs}", f"业务系统\\n{bs}"),
            (f"cust_{cust}", f"客户\\n{cust}"),
            (f"owner_{owner}", f"责任人\\n{owner}"),
        ]:
            if node_id not in seen:
                lines.append(f'    {node_id}["{label}"]')
                seen.add(node_id)

        lines.append(f"    inst_{inst} --> app_{app}")
        lines.append(f"    app_{app} --> bs_{bs}")
        lines.append(f"    bs_{bs} --> cust_{cust}")
        lines.append(f"    cust_{cust} --> owner_{owner}")

    lines.append("```")
    return "\n".join(lines)


def _metrics_table(items: list[ProcessedAlert]) -> str:
    rows = [
        "| 告警 ID | 资源 | QPS | 错误率 | P99 延迟 | 趋势 | 感知等级 |",
        "|---------|------|-----|--------|----------|------|----------|",
    ]
    for item in items:
        m = item.metrics
        qps = m.get("qps", {})
        err = m.get("errorRate", {})
        lat = m.get("latencyP99", {})
        rows.append(
            f"| {item.alert.alarm_id} | {item.alert.resource_id} "
            f"| {qps.get('current', '-')} ({qps.get('trend', '-')}) "
            f"| {err.get('current', '-')}% ({err.get('trend', '-')}) "
            f"| {lat.get('current', '-')}ms ({lat.get('trend', '-')}) "
            f"| 错误率{err.get('trend', '-')} "
            f"| {item.perception_level} |"
        )
    return "\n".join(rows)


def _correlation_section(items: list[ProcessedAlert]) -> str:
    groups = correlate_alerts(items)
    lines = ["## 关联告警聚合结果", ""]
    if len(items) == 1:
        lines.append("- **聚合结论**：单条告警，无同源关联")
        if items[0].matched_scenarios:
            s = items[0].matched_scenarios[0]
            lines.append(f"- **预案匹配**：已知问题 — {s.get('name')}（{s.get('id')}）")
        else:
            lines.append("- **预案匹配**：未匹配已知预案，按独立事件处理")
        return "\n".join(lines)

    lines.append(f"- **告警总数**：{len(items)} 条")
    for key, alarm_ids in groups.items():
        bs, vpc = key.split("|")
        lines.append(f"- **聚合组** `{bs}` / `{vpc}`：{len(alarm_ids)} 条 — {', '.join(alarm_ids)}")

    multi = [g for g in groups.values() if len(g) >= 2]
    if multi:
        lines.append("- **聚合结论**：检测到同源关联告警，建议合并分析")
    else:
        lines.append("- **聚合结论**：告警分散于不同业务域，按独立事件处理")

    known = [i for i in items if i.matched_scenarios]
    if known:
        lines.append(f"- **预案匹配**：{len(known)} 条匹配已知预案")
    return "\n".join(lines)


def generate_impact_report(items: list[ProcessedAlert], data_source: str) -> str:
    now = datetime.now(timezone.utc).astimezone().strftime("%Y-%m-%d %H:%M:%S %z")
    sections = [
        "# 告警业务影响评估报告",
        "",
        f"> 生成时间：{now}  ",
        f"> 数据来源：{data_source}  ",
        f"> 评估告警数：{len(items)}",
        "",
        "## 1. 执行摘要",
        "",
    ]

    p1 = sum(1 for i in items if i.priority == "P1")
    p2 = sum(1 for i in items if i.priority == "P2")
    p3 = sum(1 for i in items if i.priority == "P3")
    sections.append(f"- P1（立即介入）：{p1} 条")
    sections.append(f"- P2（观察）：{p2} 条")
    sections.append(f"- P3（可延后）：{p3} 条")
    sections.append("")
    sections.extend(["## 2. 资源-业务链路图", "", _mermaid_lineage(items), ""])
    sections.extend(["## 3. 客户感知指标快照（近 30 分钟）", "", _metrics_table(items), ""])
    sections.append(_correlation_section(items))
    sections.extend(["", "## 4. 告警明细", ""])

    for item in items:
        sections.append(f"### {item.alert.alarm_id} — {item.alert.alarm_name}")
        sections.append(f"- 资源：{item.alert.resource_type}/{item.alert.resource_id}")
        sections.append(f"- 触发：{item.alert.metric_name} = {item.alert.current_value}（阈值 {item.alert.threshold}）")
        sections.append(f"- 业务系统：{item.lineage.get('businessSystem', '未知')}")
        sections.append(f"- 客户：{item.lineage.get('customer', '未知')}")
        sections.append(f"- 责任人：{item.lineage.get('owner', '待确认')}")
        sections.append(f"- 链路来源：{item.lineage.get('source', 'unknown')}")
        sections.append("")

    return "\n".join(sections)


CUSTOMER_SCRIPTS: dict[str, dict[str, str]] = {
    "P1": {
        "opening": "您好，我们监测到您业务系统出现波动，技术团队已在第一时间介入排查。",
        "detail": "当前部分请求可能出现延迟或偶发失败，我们正在定位根因并推进恢复。",
        "action": "如您侧有具体受影响的功能模块或时间点，欢迎随时反馈，便于我们加速定位。",
        "closing": "后续进展我们将每 15 分钟同步一次，感谢您的理解与配合。",
    },
    "P2": {
        "opening": "您好，我们注意到您业务系统的部分监控指标出现波动。",
        "detail": "经初步评估，当前整体服务能力正常，个别指标略高于日常基线。",
        "action": "技术团队正在持续观察，如波动扩大我们将及时升级处理并通知您。",
        "closing": "目前无需您侧配合操作，如有异常感知请随时联系我们。",
    },
    "P3": {
        "opening": "您好，向您同步一条监控提示信息。",
        "detail": "系统检测到部分非核心指标出现轻微波动，当前业务运行正常。",
        "action": "该情况已在我们的关注范围内，团队将在常规巡检窗口内跟进。",
        "closing": "如后续有任何疑问，欢迎随时联系您的专属技术对接人。",
    },
}


def generate_handling_report(items: list[ProcessedAlert]) -> str:
    now = datetime.now(timezone.utc).astimezone().strftime("%Y-%m-%d %H:%M:%S %z")
    sections = [
        "# 分级处置建议与客户沟通话术",
        "",
        f"> 生成时间：{now}",
        "",
        "> 以下话术均为中性、非追责类表述，可直接发送或按客户沟通规范微调。",
        "",
    ]

    for priority in ("P1", "P2", "P3"):
        group = [i for i in items if i.priority == priority]
        if not group:
            continue

        action_map = {
            "P1": "立即介入",
            "P2": "观察 15 分钟",
            "P3": "可延后处理",
        }
        sections.extend([
            f"## {priority} — {action_map[priority]}",
            "",
            f"**涉及告警**：{', '.join(i.alert.alarm_id for i in group)}",
            "",
        ])

        for item in group:
            sections.append(f"### 告警 {item.alert.alarm_id}")
            sections.append(f"- **处置建议**：{action_map[priority]}")
            sections.append(f"- **判定依据**：{item.priority_reason}")
            sections.append(f"- **客户感知**：{item.perception_level}")
            if item.matched_scenarios:
                s = item.matched_scenarios[0]
                sections.append(f"- **预案参考**：{s.get('name')} — {s.get('description')}")
            sections.append("")

        script = CUSTOMER_SCRIPTS[priority]
        sections.extend([
            f"### {priority} 客户沟通话术草稿",
            "",
            script["opening"],
            "",
            script["detail"],
            "",
        ])

        for item in group:
            if item.matched_scenarios:
                note = item.matched_scenarios[0].get("customerNote", "")
                if note:
                    sections.append(f"补充说明（{item.alert.alarm_id}）：{note}")
                    sections.append("")

        sections.extend([script["action"], "", script["closing"], ""])

    return "\n".join(sections)


def write_reports(
    items: list[ProcessedAlert],
    output_dir: Path,
    data_source: str = "mock",
) -> dict[str, Path]:
    output_dir.mkdir(parents=True, exist_ok=True)
    ts = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")

    impact_path = output_dir / f"{ts}-impact-assessment-report.md"
    handling_path = output_dir / f"{ts}-handling-recommendation.md"

    impact_path.write_text(generate_impact_report(items, data_source), encoding="utf-8")
    handling_path.write_text(generate_handling_report(items), encoding="utf-8")

    return {"impact": impact_path, "handling": handling_path}
