package server

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/user"
)

var (
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
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
	uOps.(*user.FakeOperations).Storage[validEmail] = "secret"

	var testCases = []struct {
		token          string
		testResultFunc func(*storage.User, error)
	}{
		{
			validToken,
			func(u *storage.User, err error) {
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
			func(u *storage.User, err error) {
				if err != auth.ErrPermissionDenied {
					t.Errorf("expected ErrPermissionDenied, got %v", err)
				}
			},
		},
		{
			tokenForInvalidUser,
			func(u *storage.User, err error) {
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

func TestUnaryInterceptorIgnoreLoginRoute(t *testing.T) {
	handler := func(context.Context, interface{}) (interface{}, error) {
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "Login"}
	if _, err := unaryInterceptor(nil, nil)(context.Background(), nil, info, handler); err != nil {
		t.Error("error on process unaryInterceptor: ", err)
	}
}

func TestUnaryInterceptor(t *testing.T) {
	expectedUserEmail := "gopher@luizalabs.com"
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		u, ok := ctx.Value("user").(*storage.User)
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
	uOps.(*user.FakeOperations).Storage[expectedUserEmail] = "secret"

	validToken, err := authenticator.GenerateToken(expectedUserEmail)
	if err != nil {
		t.Fatal("error on generate token: ", err)
	}
	md := metadata.Pairs("token", validToken)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	if ok, err := unaryInterceptor(authenticator, uOps)(ctx, nil, info, handler); err != nil || !ok.(bool) {
		t.Errorf("expected successful execution, got error %v", err)
	}
}
