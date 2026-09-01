package tencentcloud

import (
	"time"

	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

type metricConfig struct {
	Namespace     string
	QPSMetric     string
	ErrorMetric   string
	LatencyMetric string
}

var metricConfigs = map[string]metricConfig{
	"cvm": {
		Namespace:     "QCE/CVM",
		QPSMetric:     "AccOuttraffic",
		ErrorMetric:   "CpuUsage",
		LatencyMetric: "MemUsage",
	},
	"clb": {
		Namespace:     "QCE/LB_PUBLIC",
		QPSMetric:     "ClientConnum",
		ErrorMetric:   "Http5xx",
		LatencyMetric: "RspAvg",
	},
	"cdb": {
		Namespace:     "QCE/CDB",
		QPSMetric:     "Qps",
		ErrorMetric:   "SlowQueries",
		LatencyMetric: "QueryLatency",
	},
}

var dimKeys = map[string]string{
	"cvm": "InstanceId",
	"clb": "loadBalancerId",
	"cdb": "InstanceId",
}

func getMonitorData(client *Client, namespace, metricName string, dimensions []map[string]string, windowMinutes, period int) ([]float64, error) {
	client.Service = "monitor"
	client.Host = "monitor.tencentcloudapi.com"

	end := time.Now().UTC()
	start := end.Add(-time.Duration(windowMinutes) * time.Minute)

	dimInterfaces := make([]map[string]interface{}, len(dimensions))
	for i, d := range dimensions {
		dimInterfaces[i] = map[string]interface{}{
			"Name":  d["Name"],
			"Value": d["Value"],
		}
	}

	resp, err := client.Call("GetMonitorData", map[string]interface{}{
		"Namespace":  namespace,
		"MetricName": metricName,
		"Period":     period,
		"StartTime":  start.Format("2006-01-02T15:04:05+00:00"),
		"EndTime":    end.Format("2006-01-02T15:04:05+00:00"),
		"Instances": []map[string]interface{}{
			{"Dimensions": dimInterfaces},
		},
	}, "2018-07-24")
	if err != nil {
		return nil, err
	}

	points, ok := resp["DataPoints"].([]interface{})
	if !ok || len(points) == 0 {
		return nil, nil
	}
	first, ok := points[0].(map[string]interface{})
	if !ok {
		return nil, nil
	}
	return asFloatSlice(first["Values"]), nil
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
		cfg = metricConfigs["cvm"]
	}
	dimKey := dimKeys[resourceType]
	if dimKey == "" {
		dimKey = "InstanceId"
	}
	dimensions := []map[string]string{{"Name": dimKey, "Value": resourceID}}

	qpsVals, err := getMonitorData(client, cfg.Namespace, cfg.QPSMetric, dimensions, windowMinutes, 60)
	if err != nil {
		return models.PerceptionMetrics{}, err
	}
	errVals, err := getMonitorData(client, cfg.Namespace, cfg.ErrorMetric, dimensions, windowMinutes, 60)
	if err != nil {
		return models.PerceptionMetrics{}, err
	}
	latVals, err := getMonitorData(client, cfg.Namespace, cfg.LatencyMetric, dimensions, windowMinutes, 60)
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
