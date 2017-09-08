package deploy

import (
	"testing"

	dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"
	"github.com/luizalabs/teresa/pkg/server/test"
)

func newRollbackRequest(name, revision string) *dpb.RollbackRequest {
	return &dpb.RollbackRequest{
		AppName:  name,
		Revision: revision,
	}
}

func TestNewRollback(t *testing.T) {
	req := newRollbackRequest("teresa", "5")
	rb := newRollback(req)

	if !test.DeepEqual(rb, req) {
		t.Errorf("expected %v, got %v", req, rb)
	}
}
