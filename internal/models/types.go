package models

// AlertInput represents a single alarm event.
type AlertInput struct {
	AlarmID       string   `json:"alarmId"`
	AlarmName     string   `json:"alarmName"`
	ResourceType  string   `json:"resourceType"`
	ResourceID    string   `json:"resourceId"`
	Region        string   `json:"region"`
	Severity      string   `json:"severity"`
	TriggerTime   string   `json:"triggerTime"`
	MetricName    string   `json:"metricName"`
	CurrentValue  *float64 `json:"currentValue"`
	Threshold     *float64 `json:"threshold"`
}

// AlertFile is the JSON input envelope.
type AlertFile struct {
	Alerts  []AlertInput `json:"alerts"`
	Options AlertOptions `json:"options"`
}

type AlertOptions struct {
	MetricWindowMinutes      int  `json:"metricWindowMinutes"`
	CorrelationWindowMinutes int  `json:"correlationWindowMinutes"`
	IncludeCustomerScript    bool `json:"includeCustomerScript"`
}

// Resource describes a cloud resource snapshot.
type Resource struct {
	ResourceType string            `json:"resourceType,omitempty"`
	InstanceID   string            `json:"InstanceId,omitempty"`
	VpcID        string            `json:"VpcId,omitempty"`
	Tags         map[string]string `json:"Tags"`
	Found        bool              `json:"found"`
}

// Lineage is the business ownership chain.
type Lineage struct {
	Available      bool   `json:"available"`
	Source         string `json:"source"`
	Instance       string `json:"instance"`
	Application    string `json:"application"`
	BusinessSystem string `json:"businessSystem"`
	Customer       string `json:"customer"`
	Owner          string `json:"owner"`
	OwnerContact   string `json:"ownerContact,omitempty"`
	Error          string `json:"error,omitempty"`
}

// MetricSeries summarizes a time series.
type MetricSeries struct {
	Current *float64 `json:"current"`
	Avg     *float64 `json:"avg"`
	Trend   string   `json:"trend"`
	Label   string   `json:"label"`
}

// PerceptionMetrics holds customer-facing metrics for ~30 minutes.
type PerceptionMetrics struct {
	QPS           MetricSeries `json:"qps"`
	ErrorRate     MetricSeries `json:"errorRate"`
	LatencyP99    MetricSeries `json:"latencyP99"`
	WindowMinutes int          `json:"windowMinutes"`
}

// ScenarioMatch describes a runbook scenario from YAML.
type ScenarioMatch struct {
	ResourceTypes      []string `yaml:"resourceTypes"`
	MetricNames        []string `yaml:"metricNames"`
	ThresholdMin       *float64 `yaml:"thresholdMin"`
	ThresholdMax       *float64 `yaml:"thresholdMax"`
	TagBusinessSystem  string   `yaml:"tagBusinessSystem"`
	CorrelatedTypes    []string `yaml:"correlatedTypes"`
	MinCorrelatedCount int      `yaml:"minCorrelatedCount"`
	SameVpc            bool     `yaml:"sameVpc"`
}

type Scenario struct {
	ID              string        `yaml:"id"`
	Name            string        `yaml:"name"`
	Match           ScenarioMatch `yaml:"match"`
	Description     string        `yaml:"description"`
	DefaultPriority string        `yaml:"defaultPriority"`
	CustomerNote    string        `yaml:"customerNote"`
}

type RunbookConfig struct {
	Scenarios []Scenario `yaml:"scenarios"`
}

// ProcessedAlert is a fully evaluated alarm.
type ProcessedAlert struct {
	Alert            AlertInput
	Resource         Resource
	Lineage          Lineage
	Metrics          PerceptionMetrics
	PerceptionLevel  string
	MatchedScenarios []Scenario
	Priority         string
	PriorityReason   string
}
