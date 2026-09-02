# 其他平台集成指南

## 通用：CLI 是唯一执行入口

```bash
./scripts/assess.sh --input alerts/your.json [--mock] [--summary]
```

## 平台对照

| 平台 | 集成方式 |
|------|----------|
| **AI Studio** | 导入 Skill zip，或加载 Skill 后调用 `assess.sh` |
| **Shell / Cron** | 定时 `assess.sh`，报告在 `deliverables/` |
| **Jenkins / GitLab CI** | pipeline 中 `go build && ./scripts/assess.sh` |
| **K8s Job** | 容器内运行 CLI，Secret 注入 ALIYUN_* |
| **Dify / Coze / FastGPT** | 自定义工具执行 shell 命令 |
| **Webhook 告警** | 接收 POST → 写 JSON → 调 CLI → 推送 Markdown |

## Webhook 伪代码

```python
@app.post("/alert/webhook")
def handle_alert(payload: dict):
    write_json("alerts/incoming.json", normalize(payload))
    run("./scripts/assess.sh --input alerts/incoming.json --summary")
    return read("deliverables/01-告警业务影响评估报告.md")
```

## JSON 摘要输出（供 API / Webhook 使用）

```bash
./bin/alert-assess assess --input alerts/x.json --deliverables deliverables --summary
```

输出示例：
```json
{
  "dataSource": "mock",
  "alertCount": 3,
  "priorities": {"P1": 1, "P2": 2, "P3": 0},
  "recommendation": "**建议**：1 条 P1 告警需立即介入..."
}
```

## 环境预检

```bash
./scripts/assess.sh --validate
# 或
./bin/alert-assess validate
```

## 跨平台部署清单

1. Clone 仓库
2. 安装 Go 1.22+（或交叉编译二进制）
3. 配置 `ALIYUN_*` 环境变量
4. 准备告警 JSON（见 alert-template.json）
5. 运行 `./scripts/assess.sh`
