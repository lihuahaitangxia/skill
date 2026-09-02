package aliyun

import (
	"fmt"
	"os"
	"strings"
)

// ConfigStatus holds environment validation results.
type ConfigStatus struct {
	Ready     bool              `json:"ready"`
	Mode      string            `json:"mode"` // mock | readonly
	Region    string            `json:"region"`
	ZoneID    string            `json:"zoneId"`
	HasAK     bool              `json:"hasAccessKey"`
	Endpoints map[string]string `json:"endpoints"`
	Missing   []string          `json:"missing,omitempty"`
	Warnings  []string          `json:"warnings,omitempty"`
}

// ValidateConfig checks credentials and endpoints for real API mode.
func ValidateConfig() ConfigStatus {
	ak, _, region, hasAK := LoadCredentialsFromEnv()
	status := ConfigStatus{
		Region:    region,
		ZoneID:    ZoneIDFromEnv(region),
		HasAK:     hasAK,
		Endpoints: map[string]string{},
	}

	services := []string{"ecs", "slb", "rds", "cms"}
	for _, svc := range services {
		status.Endpoints[svc] = ServiceEndpoint(svc)
	}

	if !hasAK {
		status.Mode = "mock"
		status.Warnings = append(status.Warnings, "未配置 ALIYUN_ACCESS_KEY_ID/SECRET，将使用 Mock 模式")
		status.Ready = true
		return status
	}

	status.Mode = "readonly"
	missing := []string{}
	for _, svc := range services {
		ep := status.Endpoints[svc]
		if strings.Contains(ep, "example.") || strings.Contains(ep, "placeholder") {
			missing = append(missing, "ALIYUN_"+strings.ToUpper(svc)+"_ENDPOINT")
		}
	}
	if len(missing) > 0 {
		status.Missing = missing
		status.Warnings = append(status.Warnings,
			fmt.Sprintf("AccessKey 已配置（%s...），但 Endpoint 未显式配置", maskKey(ak)))
		status.Ready = true
	} else {
		status.Ready = true
	}

	cmdbURL := os.Getenv("CMDB_API_URL")
	if cmdbURL == "" {
		status.Warnings = append(status.Warnings, "CMDB_API_URL 未配置，链路将降级为云标签")
	}

	return status
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "****"
}
