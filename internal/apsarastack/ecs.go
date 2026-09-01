package apsarastack

import (
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

func strVal(v interface{}, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func extractTagMap(raw interface{}) map[string]string {
	tags := make(map[string]string)
	root, ok := raw.(map[string]interface{})
	if !ok {
		return tags
	}
	tagList, ok := root["Tag"].([]interface{})
	if !ok {
		// flat Tags array
		if flat, ok := raw.([]interface{}); ok {
			tagList = flat
		} else {
			return tags
		}
	}
	for _, item := range tagList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		key, _ := m["TagKey"].(string)
		if key == "" {
			key, _ = m["Key"].(string)
		}
		val, _ := m["TagValue"].(string)
		if val == "" {
			val, _ = m["Value"].(string)
		}
		if key != "" {
			tags[key] = val
		}
	}
	return tags
}

func asFloatSlice(v interface{}) []float64 {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]float64, 0, len(arr))
	for _, item := range arr {
		if n, ok := item.(float64); ok {
			out = append(out, n)
		}
	}
	return out
}

// DescribeInstance queries ECS via DescribeInstances (2014-05-26).
func DescribeInstance(client *Client, instanceID string) (models.Resource, error) {
	endpoint := ServiceEndpoint("ecs")
	resp, err := client.CallRPC(endpoint, "DescribeInstances", "2014-05-26", map[string]string{
		"InstanceIds": `[ "` + instanceID + `" ]`,
	})
	if err != nil {
		return models.Resource{}, err
	}

	instances, ok := resp["Instances"].(map[string]interface{})
	if !ok {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}
	items, ok := instances["Instance"].([]interface{})
	if !ok || len(items) == 0 {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}

	inst, ok := items[0].(map[string]interface{})
	if !ok {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}

	vpcID := strVal(inst["VpcId"], "")
	if vpcID == "" {
		if vpcAttrs, ok := inst["VpcAttributes"].(map[string]interface{}); ok {
			vpcID = strVal(vpcAttrs["VpcId"], "")
		}
	}

	zoneID := strVal(inst["ZoneId"], "")
	if zoneID == "" {
		if placement, ok := inst["Placement"].(map[string]interface{}); ok {
			zoneID = strVal(placement["ZoneId"], "")
		}
	}

	return models.Resource{
		ResourceType: "ecs",
		InstanceID:   strVal(inst["InstanceId"], instanceID),
		VpcID:        vpcID,
		ZoneID:       zoneID,
		Region:       client.Region,
		Tags:         extractTagMap(inst["Tags"]),
		Found:        true,
	}, nil
}
