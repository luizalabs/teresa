package app

import (
	"reflect"
	"testing"

	appb "github.com/luizalabs/teresa/pkg/protobuf/app"
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

// We need to skip EnvVars
func cmpAppWithCreateRequest(app *App, req *appb.CreateRequest) bool {
	var tmp = struct {
		A string
		B string
		C string
		D *Limits
		E *AutoScale
	}{
		app.Name,
		app.Team,
		app.ProcessType,
		app.Limits,
		app.AutoScale,
	}

	return deepEqual(&tmp, req)
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
		CpuTargetUtilization: 42,
		Max:                  666,
		Min:                  1,
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

	if !cmpAppWithCreateRequest(app, req) {
		t.Errorf("expected %v, got %v", req, app)
	}
}

func TestNewInfoResponse(t *testing.T) {
	lrq1 := &LimitRangeQuantity{Quantity: "1", Resource: "resource1"}
	lrq2 := &LimitRangeQuantity{Quantity: "2", Resource: "resource2"}
	info := &Info{
		Team:      "luizalabs",
		Addresses: []*Address{{Hostname: "host1"}},
		EnvVars: []*EnvVar{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		},
		Status: &Status{
			CPU:  42,
			Pods: []*Pod{{Name: "pod 1", State: "Running"}},
		},
		AutoScale: &AutoScale{CPUTargetUtilization: 33, Max: 10, Min: 1},
		Limits: &Limits{
			Default:        []*LimitRangeQuantity{lrq1},
			DefaultRequest: []*LimitRangeQuantity{lrq2},
		},
	}

	resp := newInfoResponse(info)
	if !deepEqual(info, resp) {
		t.Errorf("expected %v, got %v", info, resp)
	}
}

func TestNewListResponse(t *testing.T) {
	items := []*AppListItem{{
		Team:      "luizalabs",
		Addresses: []*Address{{Hostname: "host1"}},
		Name:      "teste",
	}}

	resp := newListResponse(items)
	if len(items) != len(resp.Apps) {
		t.Fatal("expected %d items, got %d", len(items), len(resp.Apps))
	}
	itemExpected := items[0]
	itemActual := resp.Apps[0]
	expectedUrl := itemExpected.Addresses[0].Hostname
	actualUrl := itemActual.Urls[0]
	if expectedUrl != actualUrl {
		t.Errorf("expected %s, got %s", expectedUrl, actualUrl)
	}
	if itemExpected.Name != itemActual.Name {
		t.Errorf("expected %s, got %s", itemExpected.Name, itemActual.Name)
	}
	if itemExpected.Team != itemActual.Team {
		t.Errorf("expected %s, got %s", itemExpected.Team, itemActual.Team)
	}
}

func TestSetEnvVars(t *testing.T) {
	app := &App{Name: "teresa", Team: "luizalabs"}
	var testCases = []struct {
		evs  []*EnvVar
		want []*EnvVar
	}{
		{
			[]*EnvVar{
				{Key: "key2", Value: "value2"},
			},
			[]*EnvVar{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
		{
			[]*EnvVar{
				{Key: "key1", Value: "new-value1"},
				{Key: "key2", Value: "value2"},
			},
			[]*EnvVar{
				{Key: "key1", Value: "new-value1"},
				{Key: "key2", Value: "value2"},
			},
		},
	}

	for _, tc := range testCases {
		app.EnvVars = []*EnvVar{{Key: "key1", Value: "value1"}}
		setEnvVars(app, tc.evs)
		if !reflect.DeepEqual(app.EnvVars, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, app.EnvVars)
		}
	}
}

func TestUnsetEnvVars(t *testing.T) {
	app := &App{Name: "teresa", Team: "luizalabs"}
	var testCases = []struct {
		evs  []string
		want []*EnvVar
	}{
		{
			[]string{"key2"},
			[]*EnvVar{{Key: "key1", Value: "value1"}},
		},
		{
			[]string{"key1", "key2"},
			[]*EnvVar{},
		},
	}

	for _, tc := range testCases {
		app.EnvVars = []*EnvVar{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		}
		unsetEnvVars(app, tc.evs)
		if !reflect.DeepEqual(app.EnvVars, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, app.EnvVars)
		}
	}
}
