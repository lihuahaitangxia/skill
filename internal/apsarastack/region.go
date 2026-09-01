package apsarastack

import (
	"os"
	"strings"
)

// Default region and availability zone for Apsara Stack deployments.
const (
	DefaultRegion     = "cn-hangzhou-1"
	DefaultZoneSuffix = "a"
	DefaultZoneID     = DefaultRegion + "-" + DefaultZoneSuffix
)

// ZoneIDFromEnv resolves the full zone ID from APSARASTACK_AZ and region.
func ZoneIDFromEnv(region string) string {
	az := os.Getenv("APSARASTACK_AZ")
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
