package apsarastack

import (
	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

// DescribeLoadBalancer queries SLB via DescribeLoadBalancerAttribute (2014-05-15).
func DescribeLoadBalancer(client *Client, lbID string) (models.Resource, error) {
	endpoint := ServiceEndpoint("slb")
	resp, err := client.CallRPC(endpoint, "DescribeLoadBalancerAttribute", "2014-05-15", map[string]string{
		"LoadBalancerId": lbID,
	})
	if err != nil {
		return models.Resource{}, err
	}

	vpcID := strVal(resp["VpcId"], "")
	tags := extractTagMap(resp["Tags"])

	return models.Resource{
		ResourceType: "slb",
		InstanceID:   strVal(resp["LoadBalancerId"], lbID),
		VpcID:        vpcID,
		ZoneID:       strVal(resp["MasterZoneId"], ""),
		Region:       client.Region,
		Tags:         tags,
		Found:        true,
	}, nil
}
