package app

import appb "github.com/luizalabs/teresa/pkg/protobuf/app"

const (
	ProcessTypeWeb        = "web"
	ProcessTypeCronPrefix = "cron"
	defaultAppProtocol    = "http"
)

type LimitRangeQuantity struct {
	Quantity string
	Resource string
}

type Limits struct {
	Default        []*LimitRangeQuantity
	DefaultRequest []*LimitRangeQuantity
}

type Autoscale struct {
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
	VirtualHost string     `json:"virtualHost"`
	Limits      *Limits    `json:"-"`
	Autoscale   *Autoscale `json:"-"`
	EnvVars     []*EnvVar  `json:"envVars"`
	Internal    bool       `json:"internal"`
	Secrets     []string   `json:"secrets"`
	Protocol    string     `json:"protocol"`
}

type Pod struct {
	Name     string
	State    string
	Age      int64
	Restarts int32
	Ready    bool
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
	Autoscale *Autoscale
	Limits    *Limits
	Protocol  string
}

type AppListItem struct {
	Team      string
	Name      string
	Addresses []*Address
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

	var as *Autoscale
	if req.Autoscale != nil {
		as = &Autoscale{
			CPUTargetUtilization: req.Autoscale.CpuTargetUtilization,
			Max:                  req.Autoscale.Max,
			Min:                  req.Autoscale.Min,
		}
	}

	processType := req.ProcessType
	if processType == "" {
		processType = ProcessTypeWeb
	}
	protocol := req.Protocol
	if processType == ProcessTypeWeb && protocol == "" {
		protocol = defaultAppProtocol
	}

	app := &App{
		Autoscale: as,
		Limits: &Limits{
			Default:        def,
			DefaultRequest: defReq,
		},
		Name:        req.Name,
		ProcessType: processType,
		VirtualHost: req.VirtualHost,
		Team:        req.Team,
		EnvVars:     []*EnvVar{},
		Internal:    req.Internal,
		Protocol:    protocol,
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
				Name:     item.Name,
				State:    item.State,
				Age:      item.Age,
				Restarts: item.Restarts,
				Ready:    item.Ready,
			}
			pods = append(pods, pod)
		}
		stat = &appb.InfoResponse_Status{
			Cpu:  info.Status.CPU,
			Pods: pods,
		}
	}

	var as *appb.InfoResponse_Autoscale
	if info.Autoscale != nil {
		as = &appb.InfoResponse_Autoscale{
			CpuTargetUtilization: info.Autoscale.CPUTargetUtilization,
			Max:                  info.Autoscale.Max,
			Min:                  info.Autoscale.Min,
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
		Autoscale: as,
		Limits:    lim,
		Protocol:  info.Protocol,
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

func setSecretsOnApp(app *App, secrets []string) {
	for _, secret := range secrets {
		found := false
		for _, tmp := range app.Secrets {
			if tmp == secret {
				found = true
				break
			}
		}
		if !found {
			app.Secrets = append(app.Secrets, secret)
		}
	}
}

func unsetSecretsOnApp(app *App, secrets []string) {
	for _, secret := range secrets {
		for i := range app.Secrets {
			if app.Secrets[i] == secret {
				app.Secrets = append(app.Secrets[:i], app.Secrets[i+1:]...)
				break
			}
		}
	}
}

func newListResponse(items []*AppListItem) *appb.ListResponse {
	if items == nil {
		return nil
	}

	apps := make([]*appb.ListResponse_App, 0)
	for _, item := range items {
		addresses := make([]string, 0)
		for _, addr := range item.Addresses {
			addresses = append(addresses, addr.Hostname)
		}
		apps = append(apps, &appb.ListResponse_App{
			Urls: addresses,
			Name: item.Name,
			Team: item.Team,
		})
	}

	return &appb.ListResponse{Apps: apps}
}

func newAutoscale(req *appb.SetAutoscaleRequest) *Autoscale {
	return &Autoscale{
		CPUTargetUtilization: req.Autoscale.CpuTargetUtilization,
		Max:                  req.Autoscale.Max,
		Min:                  req.Autoscale.Min,
	}
}
