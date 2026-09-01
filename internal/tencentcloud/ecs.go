package tencentcloud

import (
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

func DescribeInstance(client *Client, instanceID string) (models.Resource, error) {
	client.Service = "cvm"
	client.Host = "cvm.tencentcloudapi.com"

	resp, err := client.Call("DescribeInstances", map[string]interface{}{
		"InstanceIds": []string{instanceID},
		"Limit":       1,
	}, "2017-03-12")
	if err != nil {
		return models.Resource{}, err
	}

	instances, ok := resp["InstanceSet"].([]interface{})
	if !ok || len(instances) == 0 {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}

	inst, ok := instances[0].(map[string]interface{})
	if !ok {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}

	tags := extractTagMap(inst["Tags"])
	vpcID := ""
	if vpc, ok := inst["VirtualPrivateCloud"].(map[string]interface{}); ok {
		vpcID, _ = vpc["VpcId"].(string)
	}

	return models.Resource{
		ResourceType: "cvm",
		InstanceID:   strVal(inst["InstanceId"], instanceID),
		VpcID:        vpcID,
		Tags:         tags,
		Found:        true,
	}, nil
}

func strVal(v interface{}, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func extractTagMap(raw interface{}) map[string]string {
	tags := make(map[string]string)
	items, ok := raw.([]interface{})
	if !ok {
		return tags
	}
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		key, _ := m["Key"].(string)
		if key == "" {
			key, _ = m["TagKey"].(string)
		}
		val, _ := m["Value"].(string)
		if val == "" {
			val, _ = m["TagValue"].(string)
		}
		if key != "" {
			tags[key] = val
		}
	}
	return tags
}

func asStringSlice(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
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
