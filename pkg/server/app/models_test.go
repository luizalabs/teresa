package app

import (
	"reflect"
	"testing"

	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
)

// Shamelessly adapted from the standard library
func deepEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}

	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)

	return deepValueEqual(v1, v2)
}

func deepValueEqual(v1, v2 reflect.Value) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}

	switch v1.Kind() {
	case reflect.Slice:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !deepValueEqual(v1.Index(i), v2.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Ptr:
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		return deepValueEqual(v1.Elem(), v2.Elem())
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !deepValueEqual(v1.Field(i), v2.Field(i)) {
				return false
			}
		}
		return true
	default:
		return v1.Interface() == v2.Interface()
	}
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

	if !deepEqual(app, req) {
		t.Errorf("expected %v, got %v", req, app)
	}
}
