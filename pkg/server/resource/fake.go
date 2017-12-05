package resource

import (
	"sync"

	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
)

type FakeOperations struct {
	mu      sync.Mutex
	Storage map[string]*Resource
}

func hasPerm(email string) bool {
	return email != "bad-user@luizalabs.com"
}

func (f *FakeOperations) Create(user *database.User, res *Resource) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !hasPerm(user.Email) {
		return "", auth.ErrPermissionDenied
	}

	if _, found := f.Storage[res.Name]; found {
		return "", ErrAlreadyExists
	}

	f.Storage[res.Name] = res
	return "", nil
}

func (f *FakeOperations) Delete(user *database.User, resName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !hasPerm(user.Email) {
		return auth.ErrPermissionDenied
	}

	if _, found := f.Storage[resName]; !found {
		return ErrNotFound
	}

	delete(f.Storage, resName)
	return nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{Storage: map[string]*Resource{}}
}
