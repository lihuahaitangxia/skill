"""CLI entry point for alert impact assessment."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from .alert_processor import AlertInput, process_alerts
from .report_generator import write_reports


def parse_alerts_from_file(path: Path) -> list[AlertInput]:
    data = json.loads(path.read_text(encoding="utf-8"))
    raw_alerts = data.get("alerts", data if isinstance(data, list) else [])
    options = data.get("options", {}) if isinstance(data, dict) else {}
    window = options.get("metricWindowMinutes", 30)

    alerts = []
    for raw in raw_alerts:
        alerts.append(
            AlertInput(
                alarm_id=raw.get("alarmId", raw.get("alarm_id", "unknown")),
                alarm_name=raw.get("alarmName", raw.get("alarm_name", "")),
                resource_type=raw.get("resourceType", raw.get("resource_type", "cvm")),
                resource_id=raw.get("resourceId", raw.get("resource_id", "")),
                region=raw.get("region", "ap-guangzhou"),
                severity=raw.get("severity", "warning"),
                trigger_time=raw.get("triggerTime", raw.get("trigger_time", "")),
                metric_name=raw.get("metricName", raw.get("metric_name", "")),
                current_value=raw.get("currentValue", raw.get("current_value")),
                threshold=raw.get("threshold"),
            )
        )
    return alerts, window


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="告警业务影响评估（只读）")
    sub = parser.add_subparsers(dest="command")

    assess = sub.add_parser("assess", help="评估告警并生成报告")
    assess.add_argument("--input", "-i", required=True, help="告警 JSON 输入文件")
    assess.add_argument("--output-dir", "-o", default="reports", help="报告输出目录")
    assess.add_argument("--mock", action="store_true", help="使用 Mock 数据（无需凭证）")

    args = parser.parse_args(argv)
    if args.command != "assess":
        parser.print_help()
        return 1

    input_path = Path(args.input)
    if not input_path.exists():
        print(f"Error: input file not found: {input_path}", file=sys.stderr)
        return 1

    alerts, window_minutes = parse_alerts_from_file(input_path)
    if not alerts:
        print("Error: no alerts found in input", file=sys.stderr)
        return 1

    processed = process_alerts(alerts, mock=args.mock, window_minutes=window_minutes)
    data_source = "mock" if args.mock else "tencent_cloud_readonly"
    paths = write_reports(processed, Path(args.output_dir), data_source)

    print("Reports generated:")
    for name, path in paths.items():
        print(f"  {name}: {path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
