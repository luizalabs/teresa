package app

import (
	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
)

const (
	processTypeWeb = "web"
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

type App struct {
	Name        string     `json:"name"`
	Team        string     `json:"-"`
	ProcessType string     `json:"processType"`
	Limits      *Limits    `json:"-"`
	AutoScale   *AutoScale `json:"-"`
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

	processType := req.ProcessType
	if processType == "" {
		processType = processTypeWeb
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
	}
	return app
}
