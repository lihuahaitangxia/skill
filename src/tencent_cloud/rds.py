"""RDS/CDB read-only queries."""

from __future__ import annotations

from typing import Any

from .client import TencentCloudClient


def describe_db_instance(client: TencentCloudClient, instance_id: str) -> dict[str, Any]:
    client.service = "cdb"
    client.host = "cdb.tencentcloudapi.com"
    resp = client.call(
        "DescribeDBInstances",
        {"InstanceIds": [instance_id], "Limit": 1},
        version="2017-03-20",
    )
    instances = resp.get("Items", [])
    if not instances:
        return {"InstanceId": instance_id, "found": False}
    db = instances[0]
    tags = {t["TagKey"]: t["TagValue"] for t in db.get("TagList", [])}
    return {
        "InstanceId": db.get("InstanceId"),
        "InstanceName": db.get("InstanceName"),
        "VpcId": db.get("UniqVpcId"),
        "Status": db.get("Status"),
        "Engine": db.get("Engine"),
        "EngineVersion": db.get("EngineVersion"),
        "Tags": tags,
        "found": True,
    }
