package client

import (
	"testing"

	"golang.org/x/net/context"
)

func TestGetRequestMetadaReturnToken(t *testing.T) {
	expectedToken := "gopher"
	ta := &tokenAuth{token: expectedToken}

	metadata, err := ta.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal("Error on get request metadata: ", err)
	}
	token, ok := metadata["token"]
	if !ok {
		t.Fatal("doen't returns token key")
	}
	if token != expectedToken {
		t.Errorf("expected %s, got %s", expectedToken, token)
	}
}
