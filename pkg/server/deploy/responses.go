package deploy

import (
	dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"
)

type ReplicaSetListItem struct {
	Revision    string
	Description string
	Age         int64
	Current     bool
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
