"""ECS/CVM read-only queries."""

from __future__ import annotations

from typing import Any

from .client import TencentCloudClient


def describe_instance(client: TencentCloudClient, instance_id: str) -> dict[str, Any]:
    client.service = "cvm"
    client.host = "cvm.tencentcloudapi.com"
    resp = client.call(
        "DescribeInstances",
        {"InstanceIds": [instance_id], "Limit": 1},
        version="2017-03-12",
    )
    instances = resp.get("InstanceSet", [])
    if not instances:
        return {"InstanceId": instance_id, "found": False}
    inst = instances[0]
    tags = {t["Key"]: t["Value"] for t in inst.get("Tags", [])}
    return {
        "InstanceId": inst.get("InstanceId"),
        "InstanceName": inst.get("InstanceName"),
        "PrivateIpAddresses": inst.get("PrivateIpAddresses", []),
        "PublicIpAddresses": inst.get("PublicIpAddresses", []),
        "VpcId": inst.get("VirtualPrivateCloud", {}).get("VpcId"),
        "Zone": inst.get("Placement", {}).get("Zone"),
        "InstanceState": inst.get("InstanceState"),
        "Tags": tags,
        "found": True,
    }
