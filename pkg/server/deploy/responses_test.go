package deploy

import (
	"reflect"
	"sort"
	"testing"

	dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"
)

func TestNewListResponse(t *testing.T) {
	items := []*ReplicaSetListItem{
		{
			Revision:    "1",
			Description: "Test 1",
			CreatedAt:   "1",
			Current:     false,
		},
		{
			Revision:    "2",
			Description: "Test 2",
			CreatedAt:   "2",
			Current:     true,
		},
	}
	want := []*dpb.ListResponse_Deploy{
		{
			Revision:    "1",
			Description: "Test 1",
			CreatedAt:   "1",
			Current:     false,
		},
		{
			Revision:    "2",
			Description: "Test 2",
			CreatedAt:   "2",
			Current:     true,
		},
	}

	resp := newListResponse(items)

	if !reflect.DeepEqual(resp.Deploys, want) {
		t.Fatalf("expected %v items, got %v", want, resp.Deploys)
	}
}

func TestSortListResponseByRevision(t *testing.T) {
	items := []*dpb.ListResponse_Deploy{
		{
			Revision:    "2",
			Description: "Test 2",
			CreatedAt:   "2",
			Current:     false,
		},
		{
			Revision:    "1",
			Description: "Test 1",
			CreatedAt:   "1",
			Current:     true,
		},
	}

	sort.Sort(ByRevision(items))

	if items[0].Revision != "1" {
		t.Errorf("expected 1, got %s", items[0].Revision)
	}
}
