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

func TestFakeOpertionsList(t *testing.T) {
	f := NewFakeOperations()
	items, err := f.List("foo", nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(items) != 2 && items[0].Name != "v1.0.0" { //see fake.go
		t.Errorf(
			"expected [v1.0.0 v1.1.0-rc1], got [%s %s]",
			items[0].Name,
			items[1].Name,
		)
	}
}
