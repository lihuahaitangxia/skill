package tencentcloud

import (
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

func DescribeDBInstance(client *Client, instanceID string) (models.Resource, error) {
	client.Service = "cdb"
	client.Host = "cdb.tencentcloudapi.com"

	resp, err := client.Call("DescribeDBInstances", map[string]interface{}{
		"InstanceIds": []string{instanceID},
		"Limit":       1,
	}, "2017-03-20")
	if err != nil {
		return models.Resource{}, err
	}

	items, ok := resp["Items"].([]interface{})
	if !ok || len(items) == 0 {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}

	db, ok := items[0].(map[string]interface{})
	if !ok {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}

	return models.Resource{
		ResourceType: "cdb",
		InstanceID:   strVal(db["InstanceId"], instanceID),
		VpcID:        strVal(db["UniqVpcId"], ""),
		Tags:         extractTagMap(db["TagList"]),
		Found:        true,
	}, nil
}
