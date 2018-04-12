package service

import (
	svcpb "github.com/luizalabs/teresa/pkg/protobuf/service"
	"github.com/luizalabs/teresa/pkg/server/spec"
)

type SSLInfo struct {
	ServicePort *spec.ServicePort
	Cert        string
}

type Info struct {
	ServicePorts []*spec.ServicePort
	SSLInfo      *SSLInfo
}

func newInfoResponse(info *Info) *svcpb.InfoResponse {
	if info == nil {
		return nil
	}
	ports := make([]*svcpb.InfoResponse_ServicePort, len(info.ServicePorts))
	for i := range info.ServicePorts {
		ports[i] = &svcpb.InfoResponse_ServicePort{int32(info.ServicePorts[i].Port)}
	}
	ssl := &svcpb.InfoResponse_SSL{
		Cert: info.SSLInfo.Cert,
		ServicePort: &svcpb.InfoResponse_ServicePort{
			int32(info.SSLInfo.ServicePort.Port),
		},
	}
	return &svcpb.InfoResponse{Ssl: ssl, ServicePorts: ports}
}
