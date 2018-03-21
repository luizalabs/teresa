package cloudprovider

import (
	"testing"

	"github.com/pkg/errors"
)

func TestNewOperationsSuccess(t *testing.T) {
	k8s := &FakeK8sOperations{CloudProviderNameValue: "aws"}

	if _, err := NewOperations(k8s); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestNewOperationsCloudProviderNameFail(t *testing.T) {
	wantErr := errors.New("test")
	k8s := &FakeK8sOperations{CloudProviderNameErr: wantErr}

	if _, err := NewOperations(k8s); errors.Cause(err) != wantErr {
		t.Errorf("got %v; want %v", errors.Cause(err), wantErr)
	}
}

func TestNewOperationsReturnsFallbackOperations(t *testing.T) {
	k8s := &FakeK8sOperations{CloudProviderNameValue: "test"}

	ops, err := NewOperations(k8s)
	if err != nil {
		t.Fatal("error creating a new operation")
	}
	if _, ok := ops.(*fallbackOperations); !ok {
		t.Error("expected fallbackOperations, but another struct was created")
	}
}
