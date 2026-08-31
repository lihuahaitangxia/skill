"""Mock data for demo without Tencent Cloud credentials."""

from __future__ import annotations

from typing import Any


MOCK_RESOURCES: dict[str, dict[str, Any]] = {
    "ins-order-svc-001": {
        "resourceType": "cvm",
        "InstanceId": "ins-order-svc-001",
        "InstanceName": "order-service-prod-01",
        "VpcId": "vpc-core-prod",
        "Tags": {
            "Application": "order-service",
            "BusinessSystem": "ecommerce-core",
            "Customer": "华东零售集团",
            "Owner": "张三",
            "OwnerContact": "138****5678",
        },
        "found": True,
    },
    "lb-gateway-prod": {
        "resourceType": "clb",
        "LoadBalancerId": "lb-gateway-prod",
        "LoadBalancerName": "api-gateway-prod",
        "VpcId": "vpc-core-prod",
        "Tags": {
            "Application": "api-gateway",
            "BusinessSystem": "ecommerce-core",
            "Customer": "华东零售集团",
            "Owner": "李四",
        },
        "found": True,
    },
    "cdb-order-primary": {
        "resourceType": "cdb",
        "InstanceId": "cdb-order-primary",
        "InstanceName": "order-db-primary",
        "VpcId": "vpc-core-prod",
        "Tags": {
            "Application": "order-db",
            "BusinessSystem": "ecommerce-core",
            "Customer": "华东零售集团",
            "Owner": "王五",
        },
        "found": True,
    },
}

MOCK_METRICS: dict[str, dict[str, Any]] = {
    "ins-order-svc-001": {
        "qps": {"current": 1250, "avg": 1180, "trend": "rising", "label": "QPS"},
        "errorRate": {"current": 6.2, "avg": 3.1, "trend": "rising", "label": "ErrorRate"},
        "latencyP99": {"current": 820, "avg": 450, "trend": "rising", "label": "Latency"},
        "windowMinutes": 30,
    },
    "lb-gateway-prod": {
        "qps": {"current": 8500, "avg": 8200, "trend": "stable", "label": "QPS"},
        "errorRate": {"current": 1.2, "avg": 0.8, "trend": "stable", "label": "ErrorRate"},
        "latencyP99": {"current": 120, "avg": 115, "trend": "stable", "label": "Latency"},
        "windowMinutes": 30,
    },
    "cdb-order-primary": {
        "qps": {"current": 320, "avg": 300, "trend": "stable", "label": "QPS"},
        "errorRate": {"current": 0.5, "avg": 0.3, "trend": "stable", "label": "ErrorRate"},
        "latencyP99": {"current": 45, "avg": 40, "trend": "stable", "label": "Latency"},
        "windowMinutes": 30,
    },
}

MOCK_CMDB: dict[str, dict[str, Any]] = {
    "ins-order-svc-001": {
        "available": True,
        "source": "cmdb",
        "instance": "ins-order-svc-001",
        "application": "order-service",
        "businessSystem": "ecommerce-core",
        "customer": "华东零售集团",
        "owner": "张三",
        "ownerContact": "138****5678",
    },
    "lb-gateway-prod": {
        "available": True,
        "source": "cmdb",
        "instance": "lb-gateway-prod",
        "application": "api-gateway",
        "businessSystem": "ecommerce-core",
        "customer": "华东零售集团",
        "owner": "李四",
        "ownerContact": "139****1234",
    },
    "cdb-order-primary": {
        "available": True,
        "source": "cmdb",
        "instance": "cdb-order-primary",
        "application": "order-db",
        "businessSystem": "ecommerce-core",
        "customer": "华东零售集团",
        "owner": "王五",
        "ownerContact": "137****9876",
    },
}


def get_mock_resource(resource_type: str, resource_id: str) -> dict[str, Any]:
    data = MOCK_RESOURCES.get(resource_id, {})
    if data:
        return data
    return {
        "resourceType": resource_type,
        "found": False,
        "Tags": {},
        "VpcId": "vpc-unknown",
    }


def get_mock_metrics(resource_id: str) -> dict[str, Any]:
    return MOCK_METRICS.get(
        resource_id,
        {
            "qps": {"current": 100, "avg": 95, "trend": "stable", "label": "QPS"},
            "errorRate": {"current": 0.5, "avg": 0.4, "trend": "stable", "label": "ErrorRate"},
            "latencyP99": {"current": 80, "avg": 75, "trend": "stable", "label": "Latency"},
            "windowMinutes": 30,
        },
    )


def get_mock_lineage(resource_id: str) -> dict[str, Any]:
    return MOCK_CMDB.get(
        resource_id,
        {
            "available": True,
            "source": "tags",
            "instance": resource_id,
            "application": "unknown-app",
            "businessSystem": "unknown-system",
            "customer": "未知客户",
            "owner": "待确认",
        },
    )
