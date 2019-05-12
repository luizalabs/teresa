package service

import (
	"errors"
	"testing"

	svcpb "github.com/luizalabs/teresa/pkg/protobuf/service"
	"github.com/luizalabs/teresa/pkg/server/database"

	context "golang.org/x/net/context"
)

func TestEnableSSLSuccess(t *testing.T) {
	fake := &FakeOperations{}
	user := &database.User{}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)
	req := &svcpb.EnableSSLRequest{}

	if _, err := svc.EnableSSL(ctx, req); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestEnableSSLFail(t *testing.T) {
	fake := &FakeOperations{EnableSSLErr: errors.New("test")}
	user := &database.User{}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)
	req := &svcpb.EnableSSLRequest{}

	if _, err := svc.EnableSSL(ctx, req); err == nil {
		t.Error("got nil; want error")
	}
}

func TestSetStaticIpSuccess(t *testing.T) {
	fake := &FakeOperations{}
	user := &database.User{}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)
	req := &svcpb.SetStaticIpRequest{}

	if _, err := svc.SetStaticIp(ctx, req); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestSetStaticIpFail(t *testing.T) {
	fake := &FakeOperations{SetStaticIpErr: errors.New("test")}
	user := &database.User{}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)
	req := &svcpb.SetStaticIpRequest{}

	if _, err := svc.SetStaticIp(ctx, req); err == nil {
		t.Error("got nil; want error")
	}
}

func TestInfoSuccess(t *testing.T) {
	fake := &FakeOperations{}
	user := &database.User{}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)
	req := &svcpb.InfoRequest{}

	if _, err := svc.Info(ctx, req); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestInfoFail(t *testing.T) {
	fake := &FakeOperations{InfoErr: errors.New("test")}
	user := &database.User{}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)
	req := &svcpb.InfoRequest{}

	if _, err := svc.Info(ctx, req); err == nil {
		t.Error("got nil; want error")
	}
}

func TestWhitelistSourceRangesSuccess(t *testing.T) {
	fake := &FakeOperations{}
	user := &database.User{}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)
	req := &svcpb.WhitelistSourceRangesRequest{}

	if _, err := svc.WhitelistSourceRanges(ctx, req); err != nil {
		t.Error("got unexpected error:", err)
	}
}

func TestWhitelistSourceRangesFail(t *testing.T) {
	fake := &FakeOperations{WhitelistSourceRangesErr: errors.New("test")}
	user := &database.User{}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)
	req := &svcpb.WhitelistSourceRangesRequest{}

	if _, err := svc.WhitelistSourceRanges(ctx, req); err == nil {
		t.Error("got nil; want error")
	}
}
