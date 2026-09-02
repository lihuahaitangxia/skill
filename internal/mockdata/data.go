package mockdata

import (
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

var resources = map[string]models.Resource{
	"i-order-svc-001": {
		ResourceType: "ecs",
		InstanceID:   "i-order-svc-001",
		VpcID:        "vpc-core-prod",
		ZoneID:       "cn-hangzhou-a",
		Region:       "cn-hangzhou",
		Tags: map[string]string{
			"Application":    "order-service",
			"BusinessSystem": "ecommerce-core",
			"Customer":       "华东零售集团",
			"Owner":          "张三",
			"OwnerContact":   "138****5678",
		},
		Found: true,
	},
	"lb-gateway-prod": {
		ResourceType: "slb",
		InstanceID:   "lb-gateway-prod",
		VpcID:        "vpc-core-prod",
		ZoneID:       "cn-hangzhou-a",
		Region:       "cn-hangzhou",
		Tags: map[string]string{
			"Application":    "api-gateway",
			"BusinessSystem": "ecommerce-core",
			"Customer":       "华东零售集团",
			"Owner":          "李四",
		},
		Found: true,
	},
	"rm-order-primary": {
		ResourceType: "rds",
		InstanceID:   "rm-order-primary",
		VpcID:        "vpc-core-prod",
		ZoneID:       "cn-hangzhou-a",
		Region:       "cn-hangzhou",
		Tags: map[string]string{
			"Application":    "order-db",
			"BusinessSystem": "ecommerce-core",
			"Customer":       "华东零售集团",
			"Owner":          "王五",
		},
		Found: true,
	},
}

var metrics = map[string]models.PerceptionMetrics{
	"i-order-svc-001": {
		QPS:           models.MetricSeries{Current: floatPtr(1250), Avg: floatPtr(1180), Trend: "rising", Label: "QPS"},
		ErrorRate:     models.MetricSeries{Current: floatPtr(6.2), Avg: floatPtr(3.1), Trend: "rising", Label: "ErrorRate"},
		LatencyP99:    models.MetricSeries{Current: floatPtr(820), Avg: floatPtr(450), Trend: "rising", Label: "Latency"},
		WindowMinutes: 30,
	},
	"lb-gateway-prod": {
		QPS:           models.MetricSeries{Current: floatPtr(8500), Avg: floatPtr(8200), Trend: "stable", Label: "QPS"},
		ErrorRate:     models.MetricSeries{Current: floatPtr(1.2), Avg: floatPtr(0.8), Trend: "stable", Label: "ErrorRate"},
		LatencyP99:    models.MetricSeries{Current: floatPtr(120), Avg: floatPtr(115), Trend: "stable", Label: "Latency"},
		WindowMinutes: 30,
	},
	"rm-order-primary": {
		QPS:           models.MetricSeries{Current: floatPtr(320), Avg: floatPtr(300), Trend: "stable", Label: "QPS"},
		ErrorRate:     models.MetricSeries{Current: floatPtr(0.5), Avg: floatPtr(0.3), Trend: "stable", Label: "ErrorRate"},
		LatencyP99:    models.MetricSeries{Current: floatPtr(45), Avg: floatPtr(40), Trend: "stable", Label: "Latency"},
		WindowMinutes: 30,
	},
}

var lineages = map[string]models.Lineage{
	"i-order-svc-001": {
		Available:      true,
		Source:         "cmdb",
		Instance:       "i-order-svc-001",
		Application:    "order-service",
		BusinessSystem: "ecommerce-core",
		Customer:       "华东零售集团",
		Owner:          "张三",
		OwnerContact:   "138****5678",
	},
	"lb-gateway-prod": {
		Available:      true,
		Source:         "cmdb",
		Instance:       "lb-gateway-prod",
		Application:    "api-gateway",
		BusinessSystem: "ecommerce-core",
		Customer:       "华东零售集团",
		Owner:          "李四",
		OwnerContact:   "139****1234",
	},
	"rm-order-primary": {
		Available:      true,
		Source:         "cmdb",
		Instance:       "rm-order-primary",
		Application:    "order-db",
		BusinessSystem: "ecommerce-core",
		Customer:       "华东零售集团",
		Owner:          "王五",
		OwnerContact:   "137****9876",
	},
}

func floatPtr(v float64) *float64 {
	return &v
}

func GetResource(resourceType, resourceID string) models.Resource {
	if r, ok := resources[resourceID]; ok {
		return r
	}
	return models.Resource{
		ResourceType: resourceType,
		InstanceID:   resourceID,
		VpcID:        "vpc-unknown",
		Tags:         map[string]string{},
		Found:        false,
	}
}

func GetMetrics(resourceID string) models.PerceptionMetrics {
	if m, ok := metrics[resourceID]; ok {
		return m
	}
	return models.PerceptionMetrics{
		QPS:           models.MetricSeries{Current: floatPtr(100), Avg: floatPtr(95), Trend: "stable", Label: "QPS"},
		ErrorRate:     models.MetricSeries{Current: floatPtr(0.5), Avg: floatPtr(0.4), Trend: "stable", Label: "ErrorRate"},
		LatencyP99:    models.MetricSeries{Current: floatPtr(80), Avg: floatPtr(75), Trend: "stable", Label: "Latency"},
		WindowMinutes: 30,
	}
}

func GetLineage(resourceID string) models.Lineage {
	if l, ok := lineages[resourceID]; ok {
		return l
	}
	return models.Lineage{
		Available:      true,
		Source:         "tags",
		Instance:       resourceID,
		Application:    "unknown-app",
		BusinessSystem: "unknown-system",
		Customer:       "未知客户",
		Owner:          "待确认",
	}
}
