package deploy

import dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"

type Rollback struct {
	AppName  string
	Revision string
}

func newRollback(req *dpb.RollbackRequest) *Rollback {
	return &Rollback{
		AppName:  req.AppName,
		Revision: req.Revision,
	}
}
