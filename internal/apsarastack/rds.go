package apsarastack

import (
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

// DescribeDBInstance queries RDS via DescribeDBInstanceAttribute (2014-08-15).
func DescribeDBInstance(client *Client, instanceID string) (models.Resource, error) {
	endpoint := ServiceEndpoint("rds")
	resp, err := client.CallRPC(endpoint, "DescribeDBInstanceAttribute", "2014-08-15", map[string]string{
		"DBInstanceId": instanceID,
	})
	if err != nil {
		return models.Resource{}, err
	}

	items, ok := resp["Items"].(map[string]interface{})
	if !ok {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}
	dbList, ok := items["DBInstanceAttribute"].([]interface{})
	if !ok || len(dbList) == 0 {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}

	db, ok := dbList[0].(map[string]interface{})
	if !ok {
		return models.Resource{InstanceID: instanceID, Found: false, Tags: map[string]string{}}, nil
	}

	return models.Resource{
		ResourceType: "rds",
		InstanceID:   strVal(db["DBInstanceId"], instanceID),
		VpcID:        strVal(db["VpcId"], ""),
		Tags:         extractTagMap(db["Tags"]),
		Found:        true,
	}, nil
}
