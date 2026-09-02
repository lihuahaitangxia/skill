package aliyun

import (
	"os"
	"strings"
)

// Default region and availability zone for Aliyun deployments.
const (
	DefaultRegion     = "cn-hangzhou"
	DefaultZoneSuffix = "a"
	DefaultZoneID     = DefaultRegion + "-" + DefaultZoneSuffix
)

// ZoneIDFromEnv resolves the full zone ID from ALIYUN_AZ and region.
func ZoneIDFromEnv(region string) string {
	az := os.Getenv("ALIYUN_AZ")
	if az == "" {
		az = DefaultZoneSuffix
	}
	return ZoneID(region, az)
}

// ZoneID returns the full zone ID, combining region and AZ suffix when needed.
func ZoneID(region, az string) string {
	az = strings.TrimPrefix(az, "-")
	if strings.Contains(az, "-") {
		return az
	}
	if region == "" {
		region = DefaultRegion
	}
	return region + "-" + az
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
