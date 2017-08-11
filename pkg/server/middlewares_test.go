package server

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/luizalabs/teresa/pkg/server/user"
)

var (
	privateKey, _ = rsa.GenerateKey(rand.Reader, 1024)
	publicKey     = &privateKey.PublicKey
	authenticator = auth.New(privateKey, publicKey)
)

func TestAuthorize(t *testing.T) {
	validEmail := "gopher@luizalabs.com"
	validToken, err := authenticator.GenerateToken(validEmail)
	if err != nil {
		t.Fatal("error on generate token: ", err)
	}
	tokenForInvalidUser, err := authenticator.GenerateToken("invalid@luizalabs.com")
	if err != nil {
		t.Fatal("error on generate token: ", err)
	}

	uOps := user.NewFakeOperations()
	uOps.(*user.FakeOperations).Storage[validEmail] = &database.User{
		Password: "secret",
		Email:    validEmail,
	}

	var testCases = []struct {
		token          string
		testResultFunc func(*database.User, error)
	}{
		{
			validToken,
			func(u *database.User, err error) {
				if err != nil {
					t.Fatal("error on validate token: ", err)
				}
				if u.Email != validEmail {
					t.Errorf("expected %s, got %s", validEmail, u.Email)
				}
			},
		},
		{
			"invalidToken",
			func(u *database.User, err error) {
				if err != auth.ErrPermissionDenied {
					t.Errorf("expected ErrPermissionDenied, got %v", err)
				}
			},
		},
		{
			tokenForInvalidUser,
			func(u *database.User, err error) {
				if err != user.ErrNotFound {
					t.Errorf("expected user.ErrNotFound, got %v", err)
				}
			},
		},
	}

	for _, tc := range testCases {
		md := metadata.Pairs("token", tc.token)
		ctx := metadata.NewIncomingContext(context.Background(), md)
		u, err := authorize(ctx, authenticator, uOps)
		tc.testResultFunc(u, err)
	}
}

func TestLoginUnaryInterceptorIgnoreLoginRoute(t *testing.T) {
	handler := func(context.Context, interface{}) (interface{}, error) {
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "Login"}
	if _, err := loginUnaryInterceptor(nil, nil)(context.Background(), nil, info, handler); err != nil {
		t.Error("error on process unaryInterceptor: ", err)
	}
}

func TestLoginUnaryInterceptor(t *testing.T) {
	expectedUserEmail := "gopher@luizalabs.com"
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		u, ok := ctx.Value("user").(*database.User)
		if !ok {
			return false, errors.New("Context without User")
		}
		if u.Email != expectedUserEmail {
			return false, errors.New(fmt.Sprintf("expected %s, got %s", expectedUserEmail, u.Email))
		}
		return true, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "Test"}

	uOps := user.NewFakeOperations()
	uOps.(*user.FakeOperations).Storage[expectedUserEmail] = &database.User{
		Password: "secret",
		Email:    expectedUserEmail,
	}

	validToken, err := authenticator.GenerateToken(expectedUserEmail)
	if err != nil {
		t.Fatal("error on generate token: ", err)
	}
	md := metadata.Pairs("token", validToken)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	if ok, err := loginUnaryInterceptor(authenticator, uOps)(ctx, nil, info, handler); err != nil || !ok.(bool) {
		t.Errorf("expected successful execution, got error %v", err)
	}
}

func TestLoginStreamInterceptorIgnoreLoginRoute(t *testing.T) {
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}
	info := &grpc.StreamServerInfo{FullMethod: "Login"}
	if err := loginStreamInterceptor(nil, nil)(nil, nil, info, handler); err != nil {
		t.Error("error on process StreamInterceptor: ", err)
	}
}

func TestLoginStreamInterceptor(t *testing.T) {
	expectedUserEmail := "gopher@luizalabs.com"
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		ctx := stream.Context()
		u, ok := ctx.Value("user").(*database.User)
		if !ok {
			return errors.New("Context without User")
		}
		if u.Email != expectedUserEmail {
			return errors.New(fmt.Sprintf("expected %s, got %s", expectedUserEmail, u.Email))
		}
		return nil
	}
	info := &grpc.StreamServerInfo{FullMethod: "Test"}

	uOps := user.NewFakeOperations()
	uOps.(*user.FakeOperations).Storage[expectedUserEmail] = &database.User{
		Password: "secret",
		Email:    expectedUserEmail,
	}

	validToken, err := authenticator.GenerateToken(expectedUserEmail)
	if err != nil {
		t.Fatal("error on generate token: ", err)
	}
	md := metadata.Pairs("token", validToken)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	stream := &serverStreamWrapper{ctx: ctx}
	if err := loginStreamInterceptor(authenticator, uOps)(nil, stream, info, handler); err != nil {
		t.Errorf("expected successful execution, got error %v", err)
	}
}

func TestLogStreamInterceptor(t *testing.T) {
	rawErr := errors.New("error")
	grpcErr := status.Errorf(codes.Unknown, "grpc error")

	var testCases = []struct {
		rawError      error
		expectedError error
	}{
		{teresa_errors.New(grpcErr, rawErr), grpcErr},
		{grpcErr, grpcErr},
		{rawErr, rawErr},
		{nil, nil},
	}

	for _, tc := range testCases {
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			return tc.rawError
		}
		info := &grpc.StreamServerInfo{FullMethod: "Test"}

		ss := &serverStreamWrapper{ctx: context.Background()}
		actualError := logStreamInterceptor(nil, ss, info, handler)
		if actualError != tc.expectedError {
			t.Errorf("expected %v, got %v", tc.expectedError, actualError)
		}
	}
}

func TestLogUnaryInterceptor(t *testing.T) {
	rawErr := errors.New("error")
	grpcErr := status.Errorf(codes.Unknown, "grpc error")

	var testCases = []struct {
		expectedResult string
		rawError       error
		expectedError  error
	}{
		{"result", rawErr, rawErr},
		{"result", teresa_errors.New(grpcErr, rawErr), grpcErr},
		{"result", nil, nil},
		{"", grpcErr, grpcErr},
	}

	for _, tc := range testCases {
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return tc.expectedResult, tc.rawError
		}
		info := &grpc.UnaryServerInfo{FullMethod: "Test"}

		actualResult, actualError := logUnaryInterceptor(context.Background(), nil, info, handler)
		if actualResult != tc.expectedResult {
			t.Errorf("expected %s, got %s", tc.expectedResult, actualResult)
		}
		if actualError != tc.expectedError {
			t.Errorf("expected %v, got %v", tc.expectedError, actualError)
		}
	}
}
