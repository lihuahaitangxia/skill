package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhe-xing/alert-impact-assessment/internal/apsarastack"
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
	"github.com/zhe-xing/alert-impact-assessment/internal/processor"
)

type customerScript struct {
	opening, detail, action, closing string
}

var customerScripts = map[string]customerScript{
	"P1": {
		opening: "您好，我们监测到您业务系统出现波动，技术团队已在第一时间介入排查。",
		detail:  "当前部分请求可能出现延迟或偶发失败，我们正在定位根因并推进恢复。",
		action:  "如您侧有具体受影响的功能模块或时间点，欢迎随时反馈，便于我们加速定位。",
		closing: "后续进展我们将每 15 分钟同步一次，感谢您的理解与配合。",
	},
	"P2": {
		opening: "您好，我们注意到您业务系统的部分监控指标出现波动。",
		detail:  "经初步评估，当前整体服务能力正常，个别指标略高于日常基线。",
		action:  "技术团队正在持续观察，如波动扩大我们将及时升级处理并通知您。",
		closing: "目前无需您侧配合操作，如有异常感知请随时联系我们。",
	},
	"P3": {
		opening: "您好，向您同步一条监控提示信息。",
		detail:  "系统检测到部分非核心指标出现轻微波动，当前业务运行正常。",
		action:  "该情况已在我们的关注范围内，团队将在常规巡检窗口内跟进。",
		closing: "如后续有任何疑问，欢迎随时联系您的专属技术对接人。",
	},
}

var actionMap = map[string]string{
	"P1": "立即介入",
	"P2": "观察 15 分钟",
	"P3": "可延后处理",
}

func mermaidLineage(items []models.ProcessedAlert) string {
	lines := []string{"```mermaid", "flowchart LR"}
	seen := make(map[string]bool)

	for _, item := range items {
		inst := item.Lineage.Instance
		if inst == "" {
			inst = item.Alert.ResourceID
		}
		app := orDefault(item.Lineage.Application, "未知应用")
		bs := orDefault(item.Lineage.BusinessSystem, "未知业务系统")
		cust := orDefault(item.Lineage.Customer, "未知客户")
		owner := orDefault(item.Lineage.Owner, "待确认")

		nodes := []struct {
			id, label string
		}{
			{"inst_" + inst, "实例\\n" + inst},
			{"app_" + app, "应用\\n" + app},
			{"bs_" + bs, "业务系统\\n" + bs},
			{"cust_" + cust, "客户\\n" + cust},
			{"owner_" + owner, "责任人\\n" + owner},
		}
		for _, n := range nodes {
			if !seen[n.id] {
				lines = append(lines, fmt.Sprintf("    %s[\"%s\"]", n.id, n.label))
				seen[n.id] = true
			}
		}
		lines = append(lines,
			fmt.Sprintf("    inst_%s --> app_%s", inst, app),
			fmt.Sprintf("    app_%s --> bs_%s", app, bs),
			fmt.Sprintf("    bs_%s --> cust_%s", bs, cust),
			fmt.Sprintf("    cust_%s --> owner_%s", cust, owner),
		)
	}
	lines = append(lines, "```")
	return strings.Join(lines, "\n")
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func alertRegion(item models.ProcessedAlert) string {
	if item.Alert.Region != "" {
		return item.Alert.Region
	}
	if item.Resource.Region != "" {
		return item.Resource.Region
	}
	return apsarastack.DefaultRegion
}

func alertZone(item models.ProcessedAlert) string {
	if item.Alert.ZoneID != "" {
		return item.Alert.ZoneID
	}
	if item.Resource.ZoneID != "" {
		return item.Resource.ZoneID
	}
	if item.Alert.Az != "" {
		return apsarastack.ZoneID(alertRegion(item), item.Alert.Az)
	}
	return apsarastack.DefaultZoneID
}

func metricsTable(items []models.ProcessedAlert) string {
	rows := []string{
		"| 告警 ID | 资源 | QPS | 错误率 | P99 延迟 | 趋势 | 感知等级 |",
		"|---------|------|-----|--------|----------|------|----------|",
	}
	for _, item := range items {
		qps := formatMetric(item.Metrics.QPS)
		err := formatMetric(item.Metrics.ErrorRate)
		lat := formatMetric(item.Metrics.LatencyP99)
		rows = append(rows, fmt.Sprintf(
			"| %s | %s | %s | %s%% (%s) | %sms (%s) | 错误率%s | %s |",
			item.Alert.AlarmID,
			item.Alert.ResourceID,
			qps.val,
			err.val,
			err.trend,
			lat.val,
			lat.trend,
			err.trend,
			item.PerceptionLevel,
		))
	}
	return strings.Join(rows, "\n")
}

type metricFmt struct {
	val, trend string
}

func formatMetric(m models.MetricSeries) metricFmt {
	val := "-"
	if m.Current != nil {
		val = fmt.Sprintf("%v", *m.Current)
	}
	trend := m.Trend
	if trend == "" {
		trend = "-"
	}
	return metricFmt{val: val, trend: trend}
}

func correlationSection(items []models.ProcessedAlert) string {
	groups := processor.CorrelateAlerts(items)
	lines := []string{"## 关联告警聚合结果", ""}

	if len(items) == 1 {
		lines = append(lines, "- **聚合结论**：单条告警，无同源关联")
		if len(items[0].MatchedScenarios) > 0 {
			s := items[0].MatchedScenarios[0]
			lines = append(lines, fmt.Sprintf("- **预案匹配**：已知问题 — %s（%s）", s.Name, s.ID))
		} else {
			lines = append(lines, "- **预案匹配**：未匹配已知预案，按独立事件处理")
		}
		return strings.Join(lines, "\n")
	}

	lines = append(lines, fmt.Sprintf("- **告警总数**：%d 条", len(items)))
	for key, ids := range groups {
		parts := strings.SplitN(key, "|", 2)
		bs, vpc := parts[0], parts[1]
		lines = append(lines, fmt.Sprintf("- **聚合组** `%s` / `%s`：%d 条 — %s", bs, vpc, len(ids), strings.Join(ids, ", ")))
	}

	multi := false
	for _, ids := range groups {
		if len(ids) >= 2 {
			multi = true
			break
		}
	}
	if multi {
		lines = append(lines, "- **聚合结论**：检测到同源关联告警，建议合并分析")
	} else {
		lines = append(lines, "- **聚合结论**：告警分散于不同业务域，按独立事件处理")
	}

	known := 0
	for _, item := range items {
		if len(item.MatchedScenarios) > 0 {
			known++
		}
	}
	if known > 0 {
		lines = append(lines, fmt.Sprintf("- **预案匹配**：%d 条匹配已知预案", known))
	}
	return strings.Join(lines, "\n")
}

func GenerateImpactReport(items []models.ProcessedAlert, dataSource string) string {
	now := time.Now().Format("2006-01-02 15:04:05 MST")
	window := 30
	if len(items) > 0 && items[0].Metrics.WindowMinutes > 0 {
		window = items[0].Metrics.WindowMinutes
	}
	var p1, p2, p3 int
	for _, item := range items {
		switch item.Priority {
		case "P1":
			p1++
		case "P2":
			p2++
		case "P3":
			p3++
		}
	}

	sections := []string{
		"# 告警业务影响评估报告",
		"",
		fmt.Sprintf("> 生成时间：%s  ", now),
		fmt.Sprintf("> 数据来源：%s  ", dataSource),
		fmt.Sprintf("> 评估告警数：%d", len(items)),
		"",
		"## 1. 执行摘要",
		"",
		fmt.Sprintf("- P1（立即介入）：%d 条", p1),
		fmt.Sprintf("- P2（观察）：%d 条", p2),
		fmt.Sprintf("- P3（可延后）：%d 条", p3),
		"",
		executiveRecommendation(p1, p2, p3),
		"",
		"## 2. 资源-业务链路图",
		"",
		mermaidLineage(items),
		"",
		fmt.Sprintf("## 3. 客户感知指标快照（近 %d 分钟）", window),
		"",
		metricsTable(items),
		"",
		correlationSection(items),
		"",
		"## 4. 告警明细",
		"",
	}

	for _, item := range items {
		cur := "-"
		if item.Alert.CurrentValue != nil {
			cur = fmt.Sprintf("%v", *item.Alert.CurrentValue)
		}
		th := "-"
		if item.Alert.Threshold != nil {
			th = fmt.Sprintf("%v", *item.Alert.Threshold)
		}
		sections = append(sections,
			fmt.Sprintf("### %s — %s", item.Alert.AlarmID, item.Alert.AlarmName),
			fmt.Sprintf("- 资源：%s/%s", item.Alert.ResourceType, item.Alert.ResourceID),
			fmt.Sprintf("- 地域/可用区：%s / %s", alertRegion(item), alertZone(item)),
			fmt.Sprintf("- 触发：%s = %s（阈值 %s）", item.Alert.MetricName, cur, th),
			fmt.Sprintf("- 业务系统：%s", orDefault(item.Lineage.BusinessSystem, "未知")),
			fmt.Sprintf("- 客户：%s", orDefault(item.Lineage.Customer, "未知")),
			fmt.Sprintf("- 责任人：%s", orDefault(item.Lineage.Owner, "待确认")),
			fmt.Sprintf("- 链路来源：%s", orDefault(item.Lineage.Source, "unknown")),
			"",
		)
	}
	return strings.Join(sections, "\n")
}

func GenerateHandlingReport(items []models.ProcessedAlert) string {
	now := time.Now().Format("2006-01-02 15:04:05 MST")
	sections := []string{
		"# 分级处置建议与客户沟通话术",
		"",
		fmt.Sprintf("> 生成时间：%s", now),
		"",
		"> 以下话术均为中性、非追责类表述，可直接发送或按客户沟通规范微调。",
		"",
	}

	for _, priority := range []string{"P1", "P2", "P3"} {
		group := filterByPriority(items, priority)
		if len(group) == 0 {
			continue
		}

		ids := make([]string, len(group))
		for i, item := range group {
			ids[i] = item.Alert.AlarmID
		}

		sections = append(sections,
			fmt.Sprintf("## %s — %s", priority, actionMap[priority]),
			"",
			fmt.Sprintf("**涉及告警**：%s", strings.Join(ids, ", ")),
			"",
		)

		for _, item := range group {
			sections = append(sections,
				fmt.Sprintf("### 告警 %s", item.Alert.AlarmID),
				fmt.Sprintf("- **处置建议**：%s", actionMap[priority]),
				fmt.Sprintf("- **判定依据**：%s", item.PriorityReason),
				fmt.Sprintf("- **客户感知**：%s", item.PerceptionLevel),
			)
			if len(item.MatchedScenarios) > 0 {
				s := item.MatchedScenarios[0]
				sections = append(sections, fmt.Sprintf("- **预案参考**：%s — %s", s.Name, s.Description))
			}
			sections = append(sections, "")
		}

		script := customerScripts[priority]
		sections = append(sections,
			fmt.Sprintf("### %s 客户沟通话术草稿", priority),
			"",
			script.opening,
			"",
			script.detail,
			"",
		)
		for _, item := range group {
			if len(item.MatchedScenarios) > 0 && item.MatchedScenarios[0].CustomerNote != "" {
				sections = append(sections,
					fmt.Sprintf("补充说明（%s）：%s", item.Alert.AlarmID, item.MatchedScenarios[0].CustomerNote),
					"",
				)
			}
		}
		sections = append(sections, script.action, "", script.closing, "")
	}

	// 附录：未触发的级别仍提供标准话术模板
	triggered := map[string]bool{}
	for _, item := range items {
		triggered[item.Priority] = true
	}
	for _, priority := range []string{"P1", "P2", "P3"} {
		if triggered[priority] {
			continue
		}
		script := customerScripts[priority]
		sections = append(sections,
			fmt.Sprintf("## %s — %s（本次未触发，标准模板）", priority, actionMap[priority]),
			"",
			script.opening,
			"",
			script.detail,
			"",
			script.action,
			"",
			script.closing,
			"",
		)
	}

	return strings.Join(sections, "\n")
}

func executiveRecommendation(p1, p2, p3 int) string {
	switch {
	case p1 > 0:
		return fmt.Sprintf("**建议**：%d 条 P1 告警需立即介入，建议合并关联告警统一排查，每 15 分钟同步客户。", p1)
	case p2 > 0:
		return fmt.Sprintf("**建议**：%d 条 P2 告警建议持续观察 15 分钟，如指标扩大则升级 P1。", p2)
	default:
		return "**建议**：当前无紧急告警，可纳入常规巡检窗口处理。"
	}
}

func filterByPriority(items []models.ProcessedAlert, priority string) []models.ProcessedAlert {
	out := make([]models.ProcessedAlert, 0)
	for _, item := range items {
		if item.Priority == priority {
			out = append(out, item)
		}
	}
	return out
}

func WriteReports(items []models.ProcessedAlert, outputDir, dataSource string) (map[string]string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	impactPath := filepath.Join(outputDir, ts+"-impact-assessment-report.md")
	handlingPath := filepath.Join(outputDir, ts+"-handling-recommendation.md")

	if err := os.WriteFile(impactPath, []byte(GenerateImpactReport(items, dataSource)), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(handlingPath, []byte(GenerateHandlingReport(items)), 0o644); err != nil {
		return nil, err
	}

	return map[string]string{
		"impact":   impactPath,
		"handling": handlingPath,
	}, nil
}

// WriteDeliverables copies reports to fixed deliverable filenames.
func WriteDeliverables(items []models.ProcessedAlert, dir, dataSource string) (map[string]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	paths := map[string]string{
		"impact":   filepath.Join(dir, "01-告警业务影响评估报告.md"),
		"handling": filepath.Join(dir, "02-分级处置建议与客户沟通话术.md"),
	}
	if err := os.WriteFile(paths["impact"], []byte(GenerateImpactReport(items, dataSource)), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(paths["handling"], []byte(GenerateHandlingReport(items)), 0o644); err != nil {
		return nil, err
	}
	return paths, nil
}

// Summary returns a brief JSON-friendly overview for Agent / webhook use.
func Summary(items []models.ProcessedAlert, dataSource string) map[string]interface{} {
	counts := map[string]int{"P1": 0, "P2": 0, "P3": 0}
	for _, item := range items {
		counts[item.Priority]++
	}
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.Alert.AlarmID
	}
	return map[string]interface{}{
		"dataSource":  dataSource,
		"alertCount":  len(items),
		"priorities":  counts,
		"alarmIds":    ids,
		"recommendation": executiveRecommendation(counts["P1"], counts["P2"], counts["P3"]),
	}
}
