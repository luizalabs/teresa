package cmd

import (
	"fmt"

	appb "github.com/luizalabs/teresa/pkg/protobuf/app"

	"k8s.io/apimachinery/pkg/api/resource"
)

func ValidateLimits(lims *appb.CreateRequest_Limits) error {
	max, err := parseLimitRangeQuantities(lims.Default)
	if err != nil {
		return err
	}
	min, err := parseLimitRangeQuantities(lims.DefaultRequest)
	if err != nil {
		return err
	}
	for k, q := range max {
		minQ, ok := min[k]
		if ok && q.Cmp(minQ) < 0 {
			return fmt.Errorf("min %s=%s greater than max=%s", k, minQ.String(), q.String())
		}
	}
	return nil
}

func parseLimitRangeQuantities(lrq []*appb.CreateRequest_Limits_LimitRangeQuantity) (map[string]resource.Quantity, error) {
	m := map[string]resource.Quantity{}
	for _, item := range lrq {
		q, err := resource.ParseQuantity(item.Quantity)
		if err != nil {
			return nil, err
		}
		m[item.Resource] = q
	}
	return m, nil
}
