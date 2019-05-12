package cloudprovider

import (
	"testing"

	"github.com/pkg/errors"
)

func TestNewOperationsSuccess(t *testing.T) {
	k8s := &FakeK8sOperations{CloudProviderNameValue: "aws"}

	ops := NewOperations(k8s)
	if opsc, ok := ops.(*awsOperations); !ok {
		t.Errorf("got %v; want aws", opsc)
	}
	k8s = &FakeK8sOperations{CloudProviderNameValue: "gce"}

	ops = NewOperations(k8s)
	if opsc, ok := ops.(*gceOperations); !ok {
		t.Errorf("got %v; want gce", opsc)
	}
}

func TestNewOperationsCloudProviderNameFail(t *testing.T) {
	k8s := &FakeK8sOperations{CloudProviderNameErr: errors.New("test")}

	ops := NewOperations(k8s)
	if opsc, ok := ops.(*fallbackOperations); !ok {
		t.Errorf("got %v; want fallback", opsc)
	}
}

func TestNewOperationsReturnsFallbackOperations(t *testing.T) {
	k8s := &FakeK8sOperations{CloudProviderNameValue: "test"}

	ops := NewOperations(k8s)
	if _, ok := ops.(*fallbackOperations); !ok {
		t.Error("expected fallbackOperations, but another struct was created")
	}
}
