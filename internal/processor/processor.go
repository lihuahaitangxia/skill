package processor

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zhe-xing/alert-impact-assessment/internal/mockdata"
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
	"github.com/zhe-xing/alert-impact-assessment/internal/tencentcloud"
	"gopkg.in/yaml.v3"
)

var resourceTypeMap = map[string]string{
	"cvm":   "cvm",
	"ecs":   "cvm",
	"clb":   "clb",
	"slb":   "clb",
	"lb":    "clb",
	"cdb":   "cdb",
	"rds":   "cdb",
	"mysql": "cdb",
}

func NormalizeResourceType(raw string) string {
	if v, ok := resourceTypeMap[strings.ToLower(raw)]; ok {
		return v
	}
	return strings.ToLower(raw)
}

func LoadRunbookScenarios(configPath string) ([]models.Scenario, error) {
	if configPath == "" {
		configPath = "config/runbook-scenarios.yaml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg models.RunbookConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg.Scenarios, nil
}

func AssessPerception(metrics models.PerceptionMetrics) string {
	errVal := metricVal(metrics.ErrorRate.Current)
	latVal := metricVal(metrics.LatencyP99.Current)
	errTrend := metrics.ErrorRate.Trend

	if errVal > 5 && errTrend == "rising" {
		return "严重影响"
	}
	if errVal > 2 || latVal > 500 {
		return "明显影响"
	}
	if errVal > 0.5 || latVal > 200 {
		return "轻微影响"
	}
	return "正常"
}

func metricVal(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func MatchScenarios(alert models.AlertInput, resource models.Resource, scenarios []models.Scenario) []models.Scenario {
	rt := NormalizeResourceType(alert.ResourceType)
	matched := make([]models.Scenario, 0)

	for _, scenario := range scenarios {
		m := scenario.Match
		if len(m.ResourceTypes) > 0 && !contains(m.ResourceTypes, rt) {
			continue
		}
		if len(m.MetricNames) > 0 && !contains(m.MetricNames, alert.MetricName) {
			continue
		}
		if m.TagBusinessSystem != "" && resource.Tags["BusinessSystem"] != m.TagBusinessSystem {
			continue
		}
		if m.ThresholdMin != nil && metricVal(alert.CurrentValue) < *m.ThresholdMin {
			continue
		}
		if m.ThresholdMax != nil && metricVal(alert.CurrentValue) > *m.ThresholdMax {
			continue
		}
		matched = append(matched, scenario)
	}
	return matched
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func DeterminePriority(
	alert models.AlertInput,
	metrics models.PerceptionMetrics,
	perception string,
	matched []models.Scenario,
	correlatedCount int,
) (string, string) {
	errVal := metricVal(metrics.ErrorRate.Current)
	errTrend := metrics.ErrorRate.Trend

	if perception == "严重影响" || (errVal > 5 && errTrend == "rising") {
		return "P1", "错误率超过 5% 且呈上升趋势，客户感知为严重影响"
	}
	if correlatedCount >= 3 {
		return "P1", "检测到 " + itoa(correlatedCount) + " 条同源关联告警，疑似系统性问题"
	}
	if len(matched) > 0 {
		p := matched[0].DefaultPriority
		if p == "" {
			p = "P2"
		}
		return p, "匹配已知预案：" + matched[0].Name
	}
	if perception == "明显影响" || perception == "轻微影响" {
		return "P2", "客户感知为" + perception + "，建议持续观察"
	}
	return "P3", "指标波动在可控范围，可延后处理"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}

func CorrelateAlerts(processed []models.ProcessedAlert) map[string][]string {
	groups := make(map[string][]string)
	for _, item := range processed {
		bs := item.Lineage.BusinessSystem
		if bs == "" {
			bs = "unknown"
		}
		vpc := item.Resource.VpcID
		if vpc == "" {
			vpc = "unknown"
		}
		key := bs + "|" + vpc
		groups[key] = append(groups[key], item.Alert.AlarmID)
	}
	return groups
}

type ProcessOptions struct {
	Mock            bool
	WindowMinutes   int
	RunbookPath     string
	ProjectRoot     string
}

func ProcessAlerts(alerts []models.AlertInput, opts ProcessOptions) ([]models.ProcessedAlert, error) {
	runbookPath := opts.RunbookPath
	if runbookPath == "" && opts.ProjectRoot != "" {
		runbookPath = filepath.Join(opts.ProjectRoot, "config", "runbook-scenarios.yaml")
	}
	if runbookPath == "" {
		runbookPath = "config/runbook-scenarios.yaml"
	}

	scenarios, err := LoadRunbookScenarios(runbookPath)
	if err != nil {
		return nil, err
	}

	cmdb := tencentcloud.NewCMDBClient()

	var client *tencentcloud.Client
	useMock := opts.Mock
	if !useMock {
		secretID := os.Getenv("TENCENTCLOUD_SECRET_ID")
		secretKey := os.Getenv("TENCENTCLOUD_SECRET_KEY")
		region := os.Getenv("TENCENTCLOUD_REGION")
		if region == "" {
			region = "ap-guangzhou"
		}
		if secretID != "" && secretKey != "" {
			client = tencentcloud.NewClient(secretID, secretKey, region)
		} else {
			useMock = true
		}
	}

	window := opts.WindowMinutes
	if window <= 0 {
		window = 30
	}

	results := make([]models.ProcessedAlert, 0, len(alerts))
	for _, alert := range alerts {
		resource := fetchResource(client, alert, useMock)
		lineage := fetchLineage(cmdb, resource, alert, useMock)
		metrics := fetchMetrics(client, alert, window, useMock)
		perception := AssessPerception(metrics)
		matched := MatchScenarios(alert, resource, scenarios)

		results = append(results, models.ProcessedAlert{
			Alert:            alert,
			Resource:         resource,
			Lineage:          lineage,
			Metrics:          metrics,
			PerceptionLevel:  perception,
			MatchedScenarios: matched,
		})
	}

	groups := CorrelateAlerts(results)
	for i := range results {
		bs := results[i].Lineage.BusinessSystem
		if bs == "" {
			bs = "unknown"
		}
		vpc := results[i].Resource.VpcID
		if vpc == "" {
			vpc = "unknown"
		}
		key := bs + "|" + vpc
		count := len(groups[key])
		priority, reason := DeterminePriority(
			results[i].Alert,
			results[i].Metrics,
			results[i].PerceptionLevel,
			results[i].MatchedScenarios,
			count,
		)
		results[i].Priority = priority
		results[i].PriorityReason = reason
	}

	return results, nil
}

func fetchResource(client *tencentcloud.Client, alert models.AlertInput, mock bool) models.Resource {
	if mock {
		return mockdata.GetResource(alert.ResourceType, alert.ResourceID)
	}
	rt := NormalizeResourceType(alert.ResourceType)
	switch rt {
	case "cvm":
		r, err := tencentcloud.DescribeInstance(client, alert.ResourceID)
		if err != nil {
			return models.Resource{InstanceID: alert.ResourceID, Found: false, Tags: map[string]string{}}
		}
		return r
	case "clb":
		r, err := tencentcloud.DescribeLoadBalancer(client, alert.ResourceID)
		if err != nil {
			return models.Resource{InstanceID: alert.ResourceID, Found: false, Tags: map[string]string{}}
		}
		return r
	case "cdb":
		r, err := tencentcloud.DescribeDBInstance(client, alert.ResourceID)
		if err != nil {
			return models.Resource{InstanceID: alert.ResourceID, Found: false, Tags: map[string]string{}}
		}
		return r
	default:
		return models.Resource{InstanceID: alert.ResourceID, Found: false, Tags: map[string]string{}}
	}
}

func fetchLineage(cmdb *tencentcloud.CMDBClient, resource models.Resource, alert models.AlertInput, mock bool) models.Lineage {
	if mock {
		return mockdata.GetLineage(alert.ResourceID)
	}
	lineage := cmdb.GetResourceLineage(alert.ResourceType, alert.ResourceID)
	if lineage.Available {
		return lineage
	}
	return tencentcloud.LineageFromTags(resource.Tags, alert.ResourceID)
}

func fetchMetrics(client *tencentcloud.Client, alert models.AlertInput, window int, mock bool) models.PerceptionMetrics {
	if mock {
		return mockdata.GetMetrics(alert.ResourceID)
	}
	m, err := tencentcloud.FetchPerceptionMetrics(client, NormalizeResourceType(alert.ResourceType), alert.ResourceID, window)
	if err != nil {
		return models.PerceptionMetrics{WindowMinutes: window}
	}
	return m
}
