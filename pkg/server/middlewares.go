package server

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/teresa_errors"
	"github.com/luizalabs/teresa-api/pkg/server/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type serverStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *serverStreamWrapper) Context() context.Context {
	return w.ctx
}

func loginStreamInterceptor(a auth.Auth, uOps user.Operations) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if strings.HasSuffix(info.FullMethod, "Login") {
			return handler(srv, stream)
		}

		ctx := stream.Context()
		user, err := authorize(ctx, a, uOps)
		if err != nil {
			return err
		}

		ctx = context.WithValue(ctx, "user", user)
		wrap := &serverStreamWrapper{stream, ctx}
		return handler(srv, wrap)
	}
}

func loginUnaryInterceptor(a auth.Auth, uOps user.Operations) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.HasSuffix(info.FullMethod, "Login") {
			return handler(ctx, req)
		}

		user, err := authorize(ctx, a, uOps)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, "user", user)
		return handler(ctx, req)
	}
}

func authorize(ctx context.Context, a auth.Auth, uOps user.Operations) (*storage.User, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, auth.ErrPermissionDenied
	}
	if len(md["token"]) < 1 || md["token"][0] == "" {
		return nil, auth.ErrPermissionDenied
	}
	email, err := a.ValidateToken(md["token"][0])
	if err != nil {
		return nil, err
	}
	return uOps.GetUser(email)
}

func recFunc(p interface{}) (err error) {
	log.WithField("panic", p).Error("teresa-server recovered")
	return status.Errorf(codes.Unknown, "Internal Server Error")
}

func logUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		logger := log.WithField("route", info.FullMethod).WithField("request", req).WithError(err)
		u, ok := ctx.Value("user").(*storage.User)
		if ok {
			logger = logger.WithField("user", u.Email)
		}
		logger.Error("Log Interceptor got an Error")
		return resp, teresa_errors.Get(err)
	}
	return resp, nil
}

func logStreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, stream)
	if err != nil {
		logger := log.WithField("route", info.FullMethod).WithError(err)
		u, ok := stream.Context().Value("user").(*storage.User)
		if ok {
			logger = logger.WithField("user", u.Email)
		}
		logger.Error("Log Interceptor got an Error")
		return teresa_errors.Get(err)
	}
	return nil
}
