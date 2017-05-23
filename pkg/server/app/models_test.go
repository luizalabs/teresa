package app

import (
	"testing"

	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
)

func slicesLrqEquals(s1 []*LimitRangeQuantity, s2 []*appb.CreateRequest_Limits_LimitRangeQuantity) bool {
	n1 := len(s1)
	n2 := len(s2)
	if n1 != n2 {
		return false
	}
	for i, lrq1 := range s1 {
		lrq2 := s2[i]
		if lrq1.Quantity != lrq2.Quantity {
			return false
		}
		if lrq1.Resource != lrq2.Resource {
			return false
		}
	}
	return true
}

func limitsEquals(l1 *Limits, l2 *appb.CreateRequest_Limits) bool {
	if l1 == nil && l2 == nil {
		return true
	} else if l1 == nil {
		return false
	} else if l2 == nil {
		return false
	}
	if !slicesLrqEquals(l1.Default, l2.Default) {
		return false
	}
	if !slicesLrqEquals(l1.DefaultRequest, l2.DefaultRequest) {
		return false
	}
	if !slicesLrqEquals(l1.LimitRequestRatio, l2.LimitRequestRatio) {
		return false
	}
	if !slicesLrqEquals(l1.Max, l2.Max) {
		return false
	}
	if !slicesLrqEquals(l1.Min, l2.Min) {
		return false
	}
	return true
}

func autoscalesEquals(a1 *AutoScale, a2 *appb.CreateRequest_AutoScale) bool {
	if a1 == nil && a2 == nil {
		return true
	} else if a1 == nil {
		return false
	} else if a2 == nil {
		return false
	}
	if a1.CPUTargetUtilization != a2.CpuTargetUtilization {
		return false
	}
	if a1.Max != a2.Max {
		return false
	}
	if a1.Min != a2.Min {
		return false
	}
	return true
}

func newCreateRequest() *appb.CreateRequest {
	lrq1 := &appb.CreateRequest_Limits_LimitRangeQuantity{
		Quantity: "1",
		Resource: "resource1",
	}
	lrq2 := &appb.CreateRequest_Limits_LimitRangeQuantity{
		Quantity: "2",
		Resource: "resource2",
	}
	lrq3 := &appb.CreateRequest_Limits_LimitRangeQuantity{
		Quantity: "3",
		Resource: "resource3",
	}
	lrq4 := &appb.CreateRequest_Limits_LimitRangeQuantity{
		Quantity: "4",
		Resource: "resource4",
	}
	lim := &appb.CreateRequest_Limits{
		Default: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			lrq1,
			lrq2,
		},
		DefaultRequest: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			lrq3,
			lrq4,
		},
		LimitRequestRatio: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			lrq1,
			lrq3,
		},
		Max: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			lrq2,
			lrq4,
		},
		Min: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			lrq1,
			lrq3,
		},
	}
	as := &appb.CreateRequest_AutoScale{
		CpuTargetUtilization: int64(42),
		Max:                  int64(666),
		Min:                  int64(1),
	}
	return &appb.CreateRequest{
		Name:        "name",
		Team:        "team",
		ProcessType: "process_type",
		AutoScale:   as,
		Limits:      lim,
	}
}

func TestNewApp(t *testing.T) {
	req := newCreateRequest()

	app := newApp(req)

	if app.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, app.Name)
	}
	if app.Team != req.Team {
		t.Errorf("expected team %s, got %s", req.Team, app.Team)
	}
	if app.ProcessType != req.ProcessType {
		t.Errorf("expected process_type %s, got %s", req.ProcessType, app.ProcessType)
	}
	if !autoscalesEquals(app.AutoScale, req.AutoScale) {
		t.Errorf("expected auto_scale %v, got %v", req.AutoScale, app.AutoScale)
	}
	if !limitsEquals(app.Limits, req.Limits) {
		t.Errorf("expected limit %v, got %v", req.Limits, app.Limits)
	}
}
