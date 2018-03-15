package service

import (
	"testing"
)

func TestNewInfoResponse(t *testing.T) {
	info := &Info{
		ServicePorts: []*ServicePort{{Port: 1}, {Port: 2}},
		SSLInfo: &SSLInfo{
			Cert:        "cert",
			ServicePort: &ServicePort{Port: 2},
		},
	}

	resp := newInfoResponse(info)

	if len(resp.ServicePorts) != len(info.ServicePorts) {
		t.Errorf("got %d, want %d", len(resp.ServicePorts), len(info.ServicePorts))
	}
	for i := range info.ServicePorts {
		got := resp.ServicePorts[i].Port
		want := info.ServicePorts[i].Port
		if got != int32(want) {
			t.Errorf("got %d, want %d", got, want)
		}
	}
	if resp.Ssl.Cert != info.SSLInfo.Cert {
		t.Errorf("got %s, want %s", resp.Ssl.Cert, info.SSLInfo.Cert)
	}
	got := resp.Ssl.ServicePort.Port
	want := info.SSLInfo.ServicePort.Port
	if got != int32(want) {
		t.Errorf("got %d, want %d", got, want)
	}
}
