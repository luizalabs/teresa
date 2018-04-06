package build

import context "golang.org/x/net/context"

type FakeOperations struct {
}

func (f *FakeOperations) CreateByOpts(ctx context.Context, opt *CreateOptions) error {
	return nil
}

func NewFakeOperations() *FakeOperations {
	return new(FakeOperations)
}
