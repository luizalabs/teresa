package deploy

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/test"
)

func TestNewListResponse(t *testing.T) {
	items := []*ReplicaSetListItem{
		{
			Revision:    "1",
			Description: "Test 1",
			Age:         1,
			Current:     false,
		},
		{
			Revision:    "2",
			Description: "Test 2",
			Age:         2,
			Current:     true,
		},
	}

	resp := newListResponse(items)

	if !test.DeepEqual(resp.Deploys, items) {
		t.Fatal("expected %v items, got %v", resp.Deploys, items)
	}
}
