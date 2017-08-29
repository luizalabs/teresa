package deploy

import (
	"strconv"

	dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"
)

type ReplicaSetListItem struct {
	Revision    string
	Description string
	Age         int64
	Current     bool
}

type ByRevision []*dpb.ListResponse_Deploy

func (s ByRevision) Len() int {
	return len(s)
}

func (s ByRevision) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByRevision) Less(i, j int) bool {
	k, err := strconv.Atoi(s[i].Revision)
	if err != nil {
		return false
	}

	l, err := strconv.Atoi(s[j].Revision)
	if err != nil {
		return false
	}

	return k < l
}

func newListResponse(items []*ReplicaSetListItem) *dpb.ListResponse {
	if items == nil {
		return nil
	}

	resp := &dpb.ListResponse{Deploys: make([]*dpb.ListResponse_Deploy, len(items))}

	for i, item := range items {
		resp.Deploys[i] = &dpb.ListResponse_Deploy{
			Revision:    item.Revision,
			Description: item.Description,
			Age:         item.Age,
			Current:     item.Current,
		}
	}

	return resp
}
