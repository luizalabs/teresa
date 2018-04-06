package build

import (
	"testing"

	context "golang.org/x/net/context"
)

func TestFakeOpertionsCreateByOpts(t *testing.T) {
	f := NewFakeOperations()
	if err := f.CreateByOpts(context.Background(), new(CreateOptions)); err != nil {
		t.Errorf("expected non error, got %v", err)
	}
}
