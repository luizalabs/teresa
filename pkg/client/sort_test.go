package client

import (
	"testing"

	appb "github.com/luizalabs/teresa/pkg/protobuf/app"
)

func TestSortEnvsByKey(t *testing.T) {
	items := []*appb.InfoResponse_EnvVar{
		{
			Key:   "B",
			Value: "1",
		},
		{
			Key:   "A",
			Value: "2",
		},
	}

	SortEnvsByKey(items)

	if items[0].Key != "A" {
		t.Errorf("expected A, got %s", items[0].Key)
	}
}
