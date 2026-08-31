"""Tencent Cloud TC3-HMAC-SHA256 signed API client (read-only operations only)."""

from __future__ import annotations

import hashlib
import hmac
import json
import time
from datetime import datetime, timezone
from typing import Any

import requests


class RateLimiter:
    """Simple token-bucket style limiter for API calls."""

    def __init__(self, max_calls: int = 20, period_seconds: float = 1.0):
        self.max_calls = max_calls
        self.period_seconds = period_seconds
        self._timestamps: list[float] = []

    def wait_if_needed(self) -> None:
        now = time.monotonic()
        self._timestamps = [t for t in self._timestamps if now - t < self.period_seconds]
        if len(self._timestamps) >= self.max_calls:
            sleep_time = self.period_seconds - (now - self._timestamps[0])
            if sleep_time > 0:
                time.sleep(sleep_time)
        self._timestamps.append(time.monotonic())


class TencentCloudClient:
    """Minimal TC3 client for read-only Describe/Get APIs."""

    READONLY_ACTION_PREFIXES = (
        "Describe",
        "Get",
        "List",
        "Search",
        "Query",
    )

    def __init__(
        self,
        secret_id: str,
        secret_key: str,
        region: str = "ap-guangzhou",
        endpoint: str | None = None,
        service: str = "cvm",
        rate_limit: RateLimiter | None = None,
    ):
        self.secret_id = secret_id
        self.secret_key = secret_key
        self.region = region
        self.service = service
        self.host = endpoint or f"{service}.tencentcloudapi.com"
        self.rate_limiter = rate_limit or RateLimiter(max_calls=20, period_seconds=1.0)

    def _sign(self, payload: str, timestamp: int, date: str) -> str:
        algorithm = "TC3-HMAC-SHA256"
        http_request_method = "POST"
        canonical_uri = "/"
        canonical_querystring = ""
        canonical_headers = f"content-type:application/json; charset=utf-8\nhost:{self.host}\n"
        signed_headers = "content-type;host"
        hashed_request_payload = hashlib.sha256(payload.encode("utf-8")).hexdigest()
        canonical_request = (
            f"{http_request_method}\n{canonical_uri}\n{canonical_querystring}\n"
            f"{canonical_headers}\n{signed_headers}\n{hashed_request_payload}"
        )
        credential_scope = f"{date}/{self.service}/tc3_request"
        hashed_canonical = hashlib.sha256(canonical_request.encode("utf-8")).hexdigest()
        string_to_sign = f"{algorithm}\n{timestamp}\n{credential_scope}\n{hashed_canonical}"
        secret_date = hmac.new(
            f"TC3{self.secret_key}".encode("utf-8"), date.encode("utf-8"), hashlib.sha256
        ).digest()
        secret_service = hmac.new(secret_date, self.service.encode("utf-8"), hashlib.sha256).digest()
        secret_signing = hmac.new(secret_service, b"tc3_request", hashlib.sha256).digest()
        signature = hmac.new(secret_signing, string_to_sign.encode("utf-8"), hashlib.sha256).hexdigest()
        return (
            f"{algorithm} Credential={self.secret_id}/{credential_scope}, "
            f"SignedHeaders={signed_headers}, Signature={signature}"
        )

    def call(self, action: str, params: dict[str, Any], version: str) -> dict[str, Any]:
        if not action.startswith(self.READONLY_ACTION_PREFIXES):
            raise ValueError(f"Action '{action}' is not in the read-only allowlist")

        self.rate_limiter.wait_if_needed()
        payload = json.dumps(params, separators=(",", ":"))
        timestamp = int(time.time())
        date = datetime.fromtimestamp(timestamp, tz=timezone.utc).strftime("%Y-%m-%d")
        authorization = self._sign(payload, timestamp, date)

        headers = {
            "Authorization": authorization,
            "Content-Type": "application/json; charset=utf-8",
            "Host": self.host,
            "X-TC-Action": action,
            "X-TC-Timestamp": str(timestamp),
            "X-TC-Version": version,
            "X-TC-Region": self.region,
        }

        response = requests.post(
            f"https://{self.host}",
            headers=headers,
            data=payload,
            timeout=30,
        )
        response.raise_for_status()
        body = response.json()
        if "Error" in body.get("Response", {}):
            err = body["Response"]["Error"]
            raise RuntimeError(f"Tencent Cloud API error: {err.get('Code')} - {err.get('Message')}")
        return body.get("Response", body)
