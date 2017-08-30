package deploy

import (
	"testing"

	dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"
	"github.com/luizalabs/teresa/pkg/server/test"
)

func cmpRollbackWithRollbackRequest(rb *Rollback, req *dpb.RollbackRequest) bool {
	var tmp = struct {
		A string
		B string
	}{
		rb.AppName,
		rb.Revision,
	}

	return test.DeepEqual(&tmp, req)
}

func newRollbackRequest(name, revision string) *dpb.RollbackRequest {
	return &dpb.RollbackRequest{
		AppName:  name,
		Revision: revision,
	}
}

func TestNewAutoscale(t *testing.T) {
	req := newRollbackRequest("teresa", "5")
	rb := newRollback(req)

	if !cmpRollbackWithRollbackRequest(rb, req) {
		t.Errorf("expected %v, got %v", req, rb)
	}
}
