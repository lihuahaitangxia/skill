"""Cloud Monitor read-only metric queries."""

from __future__ import annotations

from datetime import datetime, timedelta, timezone
from typing import Any

from .client import TencentCloudClient


METRIC_CONFIG: dict[str, dict[str, str]] = {
    "cvm": {
        "namespace": "QCE/CVM",
        "qps_metric": "AccOuttraffic",
        "error_metric": "CpuUsage",
        "latency_metric": "MemUsage",
    },
    "clb": {
        "namespace": "QCE/LB_PUBLIC",
        "qps_metric": "ClientConnum",
        "error_metric": "Http5xx",
        "latency_metric": "RspAvg",
    },
    "cdb": {
        "namespace": "QCE/CDB",
        "qps_metric": "Qps",
        "error_metric": "SlowQueries",
        "latency_metric": "QueryLatency",
    },
}


def get_monitor_data(
    client: TencentCloudClient,
    namespace: str,
    metric_name: str,
    dimensions: list[dict[str, str]],
    window_minutes: int = 30,
    period: int = 60,
) -> dict[str, Any]:
    client.service = "monitor"
    client.host = "monitor.tencentcloudapi.com"
    end = datetime.now(timezone.utc)
    start = end - timedelta(minutes=window_minutes)
    resp = client.call(
        "GetMonitorData",
        {
            "Namespace": namespace,
            "MetricName": metric_name,
            "Period": period,
            "StartTime": start.strftime("%Y-%m-%dT%H:%M:%S+00:00"),
            "EndTime": end.strftime("%Y-%m-%dT%H:%M:%S+00:00"),
            "Instances": [{"Dimensions": dimensions}],
        },
        version="2018-07-24",
    )
    data_points = []
    for point_set in resp.get("DataPoints", []):
        data_points.extend(point_set.get("Values", []))
    return {
        "metricName": metric_name,
        "namespace": namespace,
        "values": data_points,
        "timestamps": resp.get("DataPoints", [{}])[0].get("Timestamps", []) if resp.get("DataPoints") else [],
    }


def fetch_perception_metrics(
    client: TencentCloudClient,
    resource_type: str,
    resource_id: str,
    window_minutes: int = 30,
) -> dict[str, Any]:
    cfg = METRIC_CONFIG.get(resource_type, METRIC_CONFIG["cvm"])
    dim_key = {"cvm": "InstanceId", "clb": "loadBalancerId", "cdb": "InstanceId"}.get(
        resource_type, "InstanceId"
    )
    dimensions = [{"Name": dim_key, "Value": resource_id}]

    qps = get_monitor_data(client, cfg["namespace"], cfg["qps_metric"], dimensions, window_minutes)
    errors = get_monitor_data(client, cfg["namespace"], cfg["error_metric"], dimensions, window_minutes)
    latency = get_monitor_data(client, cfg["namespace"], cfg["latency_metric"], dimensions, window_minutes)

    return {
        "qps": _summarize_series(qps["values"], "QPS"),
        "errorRate": _summarize_series(errors["values"], "ErrorRate"),
        "latencyP99": _summarize_series(latency["values"], "Latency"),
        "windowMinutes": window_minutes,
    }


def _summarize_series(values: list[float], label: str) -> dict[str, Any]:
    if not values:
        return {"current": None, "avg": None, "trend": "unknown", "label": label}
    current = values[-1]
    avg = sum(values) / len(values)
    first_half = values[: len(values) // 2] or values
    second_half = values[len(values) // 2 :] or values
    h1_avg = sum(first_half) / len(first_half)
    h2_avg = sum(second_half) / len(second_half)
    if h1_avg == 0:
        trend = "stable"
    elif (h2_avg - h1_avg) / h1_avg > 0.2:
        trend = "rising"
    elif (h1_avg - h2_avg) / h1_avg > 0.2:
        trend = "falling"
    else:
        trend = "stable"
    return {"current": round(current, 2), "avg": round(avg, 2), "trend": trend, "label": label}
