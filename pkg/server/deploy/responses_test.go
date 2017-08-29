package deploy

import (
	"sort"
	"testing"

	dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"
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

func TestSortListResponseByRevision(t *testing.T) {
	items := []*dpb.ListResponse_Deploy{
		{
			Revision:    "2",
			Description: "Test 2",
			Age:         2,
			Current:     false,
		},
		{
			Revision:    "1",
			Description: "Test 1",
			Age:         1,
			Current:     true,
		},
	}

	sort.Sort(ByRevision(items))

	if items[0].Revision != "1" {
		t.Errorf("expected 1, got %s", items[0].Revision)
	}
}
