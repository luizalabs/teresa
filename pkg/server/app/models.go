package app

import (
	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
)

type LimitRangeQuantity struct {
	Quantity string `json:"quantity"`
	Resource string `json:"resource"`
}

type Limits struct {
	Default        []*LimitRangeQuantity `json:"default"`
	DefaultRequest []*LimitRangeQuantity `json:"defaultRequest"`
}

type AutoScale struct {
	CPUTargetUtilization int64 `json:"cpuTargetUtilization,omitempty"`
	Max                  int64 `json:"max,omitempty"`
	Min                  int64 `json:"min,omitempty"`
}

type App struct {
	Name        string     `json:"name"`
	Team        string     `json:"team"`
	ProcessType string     `json:"processType,omitempty"`
	Limits      *Limits    `json:"limits,omitempty"`
	AutoScale   *AutoScale `json:"autoScale,omitempty"`
}

type Pod struct {
	Name  string
	State string
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

	app := &App{
		AutoScale: as,
		Limits: &Limits{
			Default:        def,
			DefaultRequest: defReq,
		},
		Name:        req.Name,
		ProcessType: req.ProcessType,
		Team:        req.Team,
	}
	return app
}
