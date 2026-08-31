"""CMDB / ticket system read-only adapter."""

from __future__ import annotations

import os
from typing import Any

import requests


class CMDBClient:
    """Read-only CMDB lineage lookup."""

    def __init__(self, base_url: str | None = None, token: str | None = None):
        self.base_url = (base_url or os.environ.get("CMDB_API_URL", "")).rstrip("/")
        self.token = token or os.environ.get("CMDB_API_TOKEN", "")
        self.available = bool(self.base_url and self.token)

    def get_resource_lineage(self, resource_type: str, resource_id: str) -> dict[str, Any]:
        if not self.available:
            return {"available": False, "source": "none"}

        headers = {"Authorization": f"Bearer {self.token}"}
        url = f"{self.base_url}/resources/{resource_type}/{resource_id}/lineage"
        try:
            resp = requests.get(url, headers=headers, timeout=15)
            resp.raise_for_status()
            data = resp.json()
            return {
                "available": True,
                "source": "cmdb",
                "instance": data.get("instance", resource_id),
                "application": data.get("application"),
                "businessSystem": data.get("businessSystem"),
                "customer": data.get("customer"),
                "owner": data.get("owner"),
                "ownerContact": data.get("ownerContact"),
            }
        except requests.RequestException as exc:
            return {"available": False, "source": "cmdb_error", "error": str(exc)}


def lineage_from_tags(tags: dict[str, str], resource_id: str) -> dict[str, Any]:
    """Fallback lineage from cloud resource tags."""
    return {
        "available": True,
        "source": "tags",
        "instance": resource_id,
        "application": tags.get("Application") or tags.get("app"),
        "businessSystem": tags.get("BusinessSystem") or tags.get("business"),
        "customer": tags.get("Customer") or tags.get("customer"),
        "owner": tags.get("Owner") or tags.get("owner"),
        "ownerContact": tags.get("OwnerContact"),
    }
