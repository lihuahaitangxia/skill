# 故障排查

| 现象 | 原因 | 处理 |
|------|------|------|
| `command not found: go` | 未安装 Go | 安装 Go 1.22+ 或使用预编译二进制 |
| `no alerts found` | JSON 格式错误 | 对照 alert-template.json 检查 |
| `config incomplete` | 有 AK 但缺 Endpoint | 配置 APSARASTACK_*_ENDPOINT |
| 自动切 Mock | 无凭证或 Endpoint 占位 | `./scripts/assess.sh --validate` 检查 |
| CMDB 未响应 | CMDB_API_URL/TOKEN 未配 | 降级为云标签，报告标注 source: tags |
| 指标全为 `-` | CMS 无数据或 Mock 外资源 | 确认 resourceId 在 Mock 数据中 |
| 全部 P1 | 多条同源关联告警 | 优化后：感知正常的告警按预案降级为 P2/P3 |
| Mermaid 不渲染 | 查看工具不支持 | 复制到支持 Mermaid 的编辑器 |

## 验证命令

```bash
./scripts/assess.sh --validate          # 检查环境
./scripts/assess.sh --mock --summary    # Mock 全链路测试
```
