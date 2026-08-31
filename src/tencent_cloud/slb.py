"""CLB/SLB read-only queries."""

from __future__ import annotations

from typing import Any

from .client import TencentCloudClient


def describe_load_balancer(client: TencentCloudClient, lb_id: str) -> dict[str, Any]:
    client.service = "clb"
    client.host = "clb.tencentcloudapi.com"
    resp = client.call(
        "DescribeLoadBalancers",
        {"LoadBalancerIds": [lb_id]},
        version="2018-03-17",
    )
    lbs = resp.get("LoadBalancerSet", [])
    if not lbs:
        return {"LoadBalancerId": lb_id, "found": False}
    lb = lbs[0]
    tags = {t["TagKey"]: t["TagValue"] for t in lb.get("Tags", [])}
    return {
        "LoadBalancerId": lb.get("LoadBalancerId"),
        "LoadBalancerName": lb.get("LoadBalancerName"),
        "VpcId": lb.get("VpcId"),
        "LoadBalancerType": lb.get("LoadBalancerType"),
        "Status": lb.get("Status"),
        "Tags": tags,
        "found": True,
    }
