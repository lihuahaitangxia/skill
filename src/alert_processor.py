"""Alert processing: lineage, metrics, correlation, priority."""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml

from .mock_data import get_mock_lineage, get_mock_metrics, get_mock_resource
from .tencent_cloud.client import TencentCloudClient
from .tencent_cloud.cmdb import CMDBClient, lineage_from_tags
from .tencent_cloud.ecs import describe_instance
from .tencent_cloud.monitor import fetch_perception_metrics
from .tencent_cloud.rds import describe_db_instance
from .tencent_cloud.slb import describe_load_balancer


@dataclass
class AlertInput:
    alarm_id: str
    alarm_name: str
    resource_type: str
    resource_id: str
    region: str = "ap-guangzhou"
    severity: str = "warning"
    trigger_time: str = ""
    metric_name: str = ""
    current_value: float | None = None
    threshold: float | None = None


@dataclass
class ProcessedAlert:
    alert: AlertInput
    resource: dict[str, Any]
    lineage: dict[str, Any]
    metrics: dict[str, Any]
    perception_level: str
    matched_scenarios: list[dict[str, Any]] = field(default_factory=list)
    priority: str = "P2"
    priority_reason: str = ""


def load_runbook_scenarios(config_path: Path | None = None) -> list[dict[str, Any]]:
    path = config_path or Path(__file__).resolve().parent.parent / "config" / "runbook-scenarios.yaml"
    with open(path, encoding="utf-8") as f:
        data = yaml.safe_load(f)
    return data.get("scenarios", [])


def normalize_resource_type(raw: str) -> str:
    mapping = {
        "cvm": "cvm",
        "ecs": "cvm",
        "clb": "clb",
        "slb": "clb",
        "lb": "clb",
        "cdb": "cdb",
        "rds": "cdb",
        "mysql": "cdb",
    }
    return mapping.get(raw.lower(), raw.lower())


def fetch_resource(
    client: TencentCloudClient | None,
    resource_type: str,
    resource_id: str,
    mock: bool,
) -> dict[str, Any]:
    if mock:
        return get_mock_resource(resource_type, resource_id)

    assert client is not None
    rt = normalize_resource_type(resource_type)
    if rt == "cvm":
        return describe_instance(client, resource_id)
    if rt == "clb":
        return describe_load_balancer(client, resource_id)
    if rt == "cdb":
        return describe_db_instance(client, resource_id)
    return {"found": False, "resourceId": resource_id}


def fetch_lineage(
    cmdb: CMDBClient,
    resource: dict[str, Any],
    resource_type: str,
    resource_id: str,
    mock: bool,
) -> dict[str, Any]:
    if mock:
        return get_mock_lineage(resource_id)

    lineage = cmdb.get_resource_lineage(resource_type, resource_id)
    if lineage.get("available"):
        return lineage
    tags = resource.get("Tags", {})
    return lineage_from_tags(tags, resource_id)


def fetch_metrics(
    client: TencentCloudClient | None,
    resource_type: str,
    resource_id: str,
    window_minutes: int,
    mock: bool,
) -> dict[str, Any]:
    if mock:
        return get_mock_metrics(resource_id)
    assert client is not None
    return fetch_perception_metrics(client, normalize_resource_type(resource_type), resource_id, window_minutes)


def assess_perception(metrics: dict[str, Any]) -> str:
    err = metrics.get("errorRate", {})
    lat = metrics.get("latencyP99", {})
    err_val = err.get("current") or 0
    lat_val = lat.get("current") or 0
    err_trend = err.get("trend", "stable")

    if err_val > 5 and err_trend == "rising":
        return "严重影响"
    if err_val > 2 or lat_val > 500:
        return "明显影响"
    if err_val > 0.5 or lat_val > 200:
        return "轻微影响"
    return "正常"


def match_scenarios(alert: AlertInput, resource: dict[str, Any], scenarios: list[dict[str, Any]]) -> list[dict[str, Any]]:
    matched = []
    rt = normalize_resource_type(alert.resource_type)
    tags = resource.get("Tags", {})

    for scenario in scenarios:
        match_cfg = scenario.get("match", {})
        types = match_cfg.get("resourceTypes", [])
        metrics = match_cfg.get("metricNames", [])

        if types and rt not in types:
            continue
        if metrics and alert.metric_name not in metrics:
            continue

        tag_bs = match_cfg.get("tagBusinessSystem")
        if tag_bs and tags.get("BusinessSystem") != tag_bs:
            continue

        threshold_min = match_cfg.get("thresholdMin")
        if threshold_min is not None and (alert.current_value or 0) < threshold_min:
            continue

        threshold_max = match_cfg.get("thresholdMax")
        if threshold_max is not None and (alert.current_value or 0) > threshold_max:
            continue

        matched.append(scenario)
    return matched


def determine_priority(
    alert: AlertInput,
    metrics: dict[str, Any],
    perception: str,
    matched_scenarios: list[dict[str, Any]],
    correlated_count: int,
) -> tuple[str, str]:
    err = metrics.get("errorRate", {})
    if perception == "严重影响" or (err.get("current", 0) > 5 and err.get("trend") == "rising"):
        return "P1", "错误率超过 5% 且呈上升趋势，客户感知为严重影响"
    if correlated_count >= 3:
        return "P1", f"检测到 {correlated_count} 条同源关联告警，疑似系统性问题"
    if matched_scenarios:
        default_p = matched_scenarios[0].get("defaultPriority", "P2")
        return default_p, f"匹配已知预案：{matched_scenarios[0].get('name')}"
    if perception in ("明显影响", "轻微影响"):
        return "P2", f"客户感知为{perception}，建议持续观察"
    return "P3", "指标波动在可控范围，可延后处理"


def correlate_alerts(processed: list[ProcessedAlert]) -> dict[str, list[str]]:
    """Group alerts by business system and VPC."""
    groups: dict[str, list[str]] = {}
    for item in processed:
        bs = item.lineage.get("businessSystem") or "unknown"
        vpc = item.resource.get("VpcId") or "unknown"
        key = f"{bs}|{vpc}"
        groups.setdefault(key, []).append(item.alert.alarm_id)
    return groups


def process_alerts(
    alerts: list[AlertInput],
    mock: bool = False,
    window_minutes: int = 30,
) -> list[ProcessedAlert]:
    scenarios = load_runbook_scenarios()
    cmdb = CMDBClient()

    client: TencentCloudClient | None = None
    if not mock:
        secret_id = os.environ.get("TENCENTCLOUD_SECRET_ID", "")
        secret_key = os.environ.get("TENCENTCLOUD_SECRET_KEY", "")
        region = os.environ.get("TENCENTCLOUD_REGION", "ap-guangzhou")
        if secret_id and secret_key:
            client = TencentCloudClient(secret_id, secret_key, region)

    results: list[ProcessedAlert] = []
    for alert in alerts:
        resource = fetch_resource(client, alert.resource_type, alert.resource_id, mock or client is None)
        lineage = fetch_lineage(cmdb, resource, alert.resource_type, alert.resource_id, mock or client is None)
        metrics = fetch_metrics(client, alert.resource_type, alert.resource_id, window_minutes, mock or client is None)
        perception = assess_perception(metrics)
        matched = match_scenarios(alert, resource, scenarios)
        results.append(
            ProcessedAlert(
                alert=alert,
                resource=resource,
                lineage=lineage,
                metrics=metrics,
                perception_level=perception,
                matched_scenarios=matched,
            )
        )

    groups = correlate_alerts(results)
    for item in results:
        bs = item.lineage.get("businessSystem") or "unknown"
        vpc = item.resource.get("VpcId") or "unknown"
        key = f"{bs}|{vpc}"
        count = len(groups.get(key, []))
        priority, reason = determine_priority(
            item.alert, item.metrics, item.perception_level, item.matched_scenarios, count
        )
        item.priority = priority
        item.priority_reason = reason

    return results
