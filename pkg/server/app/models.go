package app

import (
	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"

	"strings"
)

const (
	ProcessTypeWeb = "web"
)

type LimitRangeQuantity struct {
	Quantity string
	Resource string
}

type Limits struct {
	Default        []*LimitRangeQuantity
	DefaultRequest []*LimitRangeQuantity
}

type AutoScale struct {
	CPUTargetUtilization int32
	Max                  int32
	Min                  int32
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type App struct {
	Name        string     `json:"name"`
	Team        string     `json:"-"`
	ProcessType string     `json:"processType"`
	Limits      *Limits    `json:"-"`
	AutoScale   *AutoScale `json:"-"`
	EnvVars     []*EnvVar  `json:"envVars"`
}

type Pod struct {
	Name  string
	State string
}

type Address struct {
	Hostname string
}

type Status struct {
	CPU  int32
	Pods []*Pod
}

type Info struct {
	Team      string
	Addresses []*Address
	EnvVars   []*EnvVar
	Status    *Status
	AutoScale *AutoScale
	Limits    *Limits
}

type List struct {
	Team      string
	Addresses []*Address
	Name      string
}

type AppList struct {
	AppList string
}

func newSliceLrq(s []*appb.CreateRequest_Limits_LimitRangeQuantity) []*LimitRangeQuantity {
	var t []*LimitRangeQuantity
	for _, tmp := range s {
		lrq := &LimitRangeQuantity{
			Quantity: tmp.Quantity,
			Resource: tmp.Resource,
		}
		t = append(t, lrq)
	}
	return t
}

func newInfoResponseLrq(s []*LimitRangeQuantity) []*appb.InfoResponse_Limits_LimitRangeQuantity {
	var t []*appb.InfoResponse_Limits_LimitRangeQuantity
	for _, tmp := range s {
		if tmp == nil {
			continue
		}
		lrq := &appb.InfoResponse_Limits_LimitRangeQuantity{
			Quantity: tmp.Quantity,
			Resource: tmp.Resource,
		}
		t = append(t, lrq)
	}
	return t
}

func newApp(req *appb.CreateRequest) *App {
	var def, defReq []*LimitRangeQuantity
	if req.Limits != nil {
		def = newSliceLrq(req.Limits.Default)
		defReq = newSliceLrq(req.Limits.DefaultRequest)
	}

	var as *AutoScale
	if req.AutoScale != nil {
		as = &AutoScale{
			CPUTargetUtilization: req.AutoScale.CpuTargetUtilization,
			Max:                  req.AutoScale.Max,
			Min:                  req.AutoScale.Min,
		}
	}

	processType := req.ProcessType
	if processType == "" {
		processType = ProcessTypeWeb
	}

	app := &App{
		AutoScale: as,
		Limits: &Limits{
			Default:        def,
			DefaultRequest: defReq,
		},
		Name:        req.Name,
		ProcessType: processType,
		Team:        req.Team,
		EnvVars:     []*EnvVar{},
	}
	return app
}

func newInfoResponse(info *Info) *appb.InfoResponse {
	if info == nil {
		return nil
	}

	addrs := []*appb.InfoResponse_Address{}
	for _, item := range info.Addresses {
		if item == nil {
			continue
		}
		addr := &appb.InfoResponse_Address{Hostname: item.Hostname}
		addrs = append(addrs, addr)
	}

	evs := []*appb.InfoResponse_EnvVar{}
	for _, item := range info.EnvVars {
		if item == nil {
			continue
		}
		ev := &appb.InfoResponse_EnvVar{
			Key:   item.Key,
			Value: item.Value,
		}
		evs = append(evs, ev)
	}

	var stat *appb.InfoResponse_Status
	if info.Status != nil {
		pods := []*appb.InfoResponse_Status_Pod{}
		for _, item := range info.Status.Pods {
			if item == nil {
				continue
			}
			pod := &appb.InfoResponse_Status_Pod{
				Name:  item.Name,
				State: item.State,
			}
			pods = append(pods, pod)
		}
		stat = &appb.InfoResponse_Status{
			Cpu:  info.Status.CPU,
			Pods: pods,
		}
	}

	var as *appb.InfoResponse_AutoScale
	if info.AutoScale != nil {
		as = &appb.InfoResponse_AutoScale{
			CpuTargetUtilization: info.AutoScale.CPUTargetUtilization,
			Max:                  info.AutoScale.Max,
			Min:                  info.AutoScale.Min,
		}
	}

	var lim *appb.InfoResponse_Limits
	if info.Limits != nil {
		def := newInfoResponseLrq(info.Limits.Default)
		defReq := newInfoResponseLrq(info.Limits.DefaultRequest)
		lim = &appb.InfoResponse_Limits{
			Default:        def,
			DefaultRequest: defReq,
		}
	}

	return &appb.InfoResponse{
		Team:      info.Team,
		Addresses: addrs,
		EnvVars:   evs,
		Status:    stat,
		AutoScale: as,
		Limits:    lim,
	}
}

func newEnvVars(req *appb.SetEnvRequest) []*EnvVar {
	tmp := []*EnvVar{}
	for _, ev := range req.EnvVars {
		if ev == nil {
			continue
		}
		tmp = append(tmp, &EnvVar{Key: ev.Key, Value: ev.Value})
	}
	return tmp
}

func setEnvVars(app *App, evs []*EnvVar) {
	for _, ev := range evs {
		found := false
		for _, tmp := range app.EnvVars {
			if tmp.Key == ev.Key {
				tmp.Value = ev.Value
				found = true
				break
			}
		}
		if !found {
			app.EnvVars = append(app.EnvVars, &EnvVar{
				Key:   ev.Key,
				Value: ev.Value,
			})
		}
	}
}

func unsetEnvVars(app *App, evs []string) {
	for _, ev := range evs {
		for i, tmp := range app.EnvVars {
			if tmp.Key == ev {
				app.EnvVars = append(app.EnvVars[:i], app.EnvVars[i+1:]...)
				break
			}
		}
	}
}

func newListResponse(list []*List) *appb.ListResponse {
	if list == nil {
		return nil
	}

	appNames := []*appb.ListResponse_App{}

	for _, item := range list {
		if item == nil {
			continue
		}
		var tmp []string
		for _, elt := range item.Addresses {
			tmp = append(tmp, elt.Hostname)
		}
		addrs := strings.Join(tmp, ",")

		appName := &appb.ListResponse_App{
			Urls: addrs,
			App:  item.Name,
			Team: item.Team,
		}
		appNames = append(appNames, appName)
	}

	return &appb.ListResponse{
		Apps: appNames,
	}
}
