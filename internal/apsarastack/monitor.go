package apsarastack

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

type metricConfig struct {
	Namespace     string
	QPSMetric     string
	ErrorMetric   string
	LatencyMetric string
	DimensionKey  string
}

var metricConfigs = map[string]metricConfig{
	"ecs": {
		Namespace:     "acs_ecs_dashboard",
		QPSMetric:     "InternetInRate",
		ErrorMetric:   "CPUUtilization",
		LatencyMetric: "IntranetOutRate",
		DimensionKey:  "instanceId",
	},
	"slb": {
		Namespace:     "acs_slb_dashboard",
		QPSMetric:     "Qps",
		ErrorMetric:   "HttpCode5xx",
		LatencyMetric: "Latency",
		DimensionKey:  "instanceId",
	},
	"rds": {
		Namespace:     "acs_rds_dashboard",
		QPSMetric:     "MySQL_QPS",
		ErrorMetric:   "SlowQueries",
		LatencyMetric: "MySQL_NetworkTraffic",
		DimensionKey:  "instanceId",
	},
}

func describeMetricList(client *Client, namespace, metricName, dimensionKey, resourceID string, windowMinutes int) ([]float64, error) {
	endpoint := ServiceEndpoint("cms")
	end := time.Now().UTC()
	start := end.Add(-time.Duration(windowMinutes) * time.Minute)

	dimensions, _ := json.Marshal([]map[string]string{{dimensionKey: resourceID}})

	resp, err := client.CallRPC(endpoint, "DescribeMetricList", "2019-01-01", map[string]string{
		"Namespace":  namespace,
		"MetricName": metricName,
		"Dimensions": string(dimensions),
		"StartTime":  fmt.Sprintf("%d", start.UnixMilli()),
		"EndTime":    fmt.Sprintf("%d", end.UnixMilli()),
		"Period":     "60",
	})
	if err != nil {
		return nil, err
	}

	datapoints, _ := resp["Datapoints"].(string)
	if datapoints == "" {
		return nil, nil
	}

	var points []map[string]interface{}
	if err := json.Unmarshal([]byte(datapoints), &points); err != nil {
		return nil, err
	}

	values := make([]float64, 0, len(points))
	for _, p := range points {
		if v, ok := p["Average"].(float64); ok {
			values = append(values, v)
		} else if v, ok := p["Value"].(float64); ok {
			values = append(values, v)
		}
	}
	return values, nil
}

func summarizeSeries(values []float64, label string) models.MetricSeries {
	if len(values) == 0 {
		return models.MetricSeries{Trend: "unknown", Label: label}
	}

	current := values[len(values)-1]
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(len(values))

	mid := len(values) / 2
	firstHalf := values[:mid]
	if len(firstHalf) == 0 {
		firstHalf = values
	}
	secondHalf := values[mid:]
	if len(secondHalf) == 0 {
		secondHalf = values
	}

	h1Sum, h2Sum := 0.0, 0.0
	for _, v := range firstHalf {
		h1Sum += v
	}
	for _, v := range secondHalf {
		h2Sum += v
	}
	h1Avg := h1Sum / float64(len(firstHalf))
	h2Avg := h2Sum / float64(len(secondHalf))

	trend := "stable"
	if h1Avg == 0 {
		trend = "stable"
	} else if (h2Avg-h1Avg)/h1Avg > 0.2 {
		trend = "rising"
	} else if (h1Avg-h2Avg)/h1Avg > 0.2 {
		trend = "falling"
	}

	cur, a := round2(current), round2(avg)
	return models.MetricSeries{
		Current: &cur,
		Avg:     &a,
		Trend:   trend,
		Label:   label,
	}
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func FetchPerceptionMetrics(client *Client, resourceType, resourceID string, windowMinutes int) (models.PerceptionMetrics, error) {
	cfg, ok := metricConfigs[resourceType]
	if !ok {
		cfg = metricConfigs["ecs"]
	}

	qpsVals, err := describeMetricList(client, cfg.Namespace, cfg.QPSMetric, cfg.DimensionKey, resourceID, windowMinutes)
	if err != nil {
		return models.PerceptionMetrics{}, err
	}
	errVals, err := describeMetricList(client, cfg.Namespace, cfg.ErrorMetric, cfg.DimensionKey, resourceID, windowMinutes)
	if err != nil {
		return models.PerceptionMetrics{}, err
	}
	latVals, err := describeMetricList(client, cfg.Namespace, cfg.LatencyMetric, cfg.DimensionKey, resourceID, windowMinutes)
	if err != nil {
		return models.PerceptionMetrics{}, err
	}

	return models.PerceptionMetrics{
		QPS:           summarizeSeries(qpsVals, "QPS"),
		ErrorRate:     summarizeSeries(errVals, "ErrorRate"),
		LatencyP99:    summarizeSeries(latVals, "Latency"),
		WindowMinutes: windowMinutes,
	}, nil
}
