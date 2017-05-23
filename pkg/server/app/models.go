package app

import (
	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
)

type LimitRangeQuantity struct {
	Quantity string `json:"quantity"`
	Resource string `json:"resource"`
}

type Limits struct {
	Default           []*LimitRangeQuantity `json:"default"`
	DefaultRequest    []*LimitRangeQuantity `json:"defaultRequest"`
	LimitRequestRatio []*LimitRangeQuantity `json:"limitRequestRatio,omitempty"`
	Max               []*LimitRangeQuantity `json:"max,omitempty"`
	Min               []*LimitRangeQuantity `json:"min,omitempty"`
}

type AutoScale struct {
	CPUTargetUtilization int64 `json:"cpuTargetUtilization,omitempty"`
	Max                  int64 `json:"max,omitempty"`
	Min                  int64 `json:"min,omitempty"`
}

type App struct {
	AutoScale   *AutoScale `json:"autoScale,omitempty"`
	Limits      *Limits    `json:"limits,omitempty"`
	Name        string     `json:"name"`
	ProcessType string     `json:"processType,omitempty"`
	Team        string     `json:"team"`
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
	var def, defReq, limReqRatio, min, max []*LimitRangeQuantity
	if req.Limits != nil {
		def = newSliceLrq(req.Limits.Default)
		defReq = newSliceLrq(req.Limits.DefaultRequest)
		limReqRatio = newSliceLrq(req.Limits.LimitRequestRatio)
		min = newSliceLrq(req.Limits.Min)
		max = newSliceLrq(req.Limits.Max)
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
			Default:           def,
			DefaultRequest:    defReq,
			LimitRequestRatio: limReqRatio,
			Max:               max,
			Min:               min,
		},
		Name:        req.Name,
		ProcessType: req.ProcessType,
		Team:        req.Team,
	}
	return app
}
