package apsarastack

import (
	"testing"
)

// 测试 Tag 解析函数的健壮性
func TestExtractTagMap(t *testing.T) {
	mockRawTag := map[string]interface{}{
		"Tag": []interface{}{
			map[string]interface{}{
				"TagKey":   "Environment",
				"TagValue": "Production",
			},
			map[string]interface{}{
				"Key":   "Owner",
				"Value": "DevOpsTeam",
			},
		},
	}

	tags := extractTagMap(mockRawTag)

	if tags["Environment"] != "Production" {
		t.Errorf("Expected Production, got %s", tags["Environment"])
	}
	if tags["Owner"] != "DevOpsTeam" {
		t.Errorf("Expected DevOpsTeam, got %s", tags["Owner"])
	}
}

// 测试字符串安全提取工具函数
func TestStrVal(t *testing.T) {
	val := strVal("ins-123456", "fallback-id")
	if val != "ins-123456" {
		t.Errorf("Expected ins-123456, got %s", val)
	}

	fallbackVal := strVal(nil, "fallback-id")
	if fallbackVal != "fallback-id" {
		t.Errorf("Expected fallback-id, got %s", fallbackVal)
	}
}
