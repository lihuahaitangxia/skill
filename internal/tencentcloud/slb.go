package tencentcloud

import (
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

func DescribeLoadBalancer(client *Client, lbID string) (models.Resource, error) {
	client.Service = "clb"
	client.Host = "clb.tencentcloudapi.com"

	resp, err := client.Call("DescribeLoadBalancers", map[string]interface{}{
		"LoadBalancerIds": []string{lbID},
	}, "2018-03-17")
	if err != nil {
		return models.Resource{}, err
	}

	lbs, ok := resp["LoadBalancerSet"].([]interface{})
	if !ok || len(lbs) == 0 {
		return models.Resource{InstanceID: lbID, Found: false, Tags: map[string]string{}}, nil
	}

	lb, ok := lbs[0].(map[string]interface{})
	if !ok {
		return models.Resource{InstanceID: lbID, Found: false, Tags: map[string]string{}}, nil
	}

	return models.Resource{
		ResourceType: "clb",
		InstanceID:   strVal(lb["LoadBalancerId"], lbID),
		VpcID:        strVal(lb["VpcId"], ""),
		Tags:         extractTagMap(lb["Tags"]),
		Found:        true,
	}, nil
}
