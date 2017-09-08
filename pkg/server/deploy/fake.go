package deploy

import (
	"io"
	"sync"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
)

type FakeOperations struct {
	mutex   *sync.RWMutex
	Storage map[string]bool
}

func hasPerm(email string) bool {
	return email != "bad-user@luizalabs.com"
}

func (f *FakeOperations) List(user *database.User, appName string) ([]*ReplicaSetListItem, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if !hasPerm(user.Email) {
		return nil, auth.ErrPermissionDenied
	}

	if _, found := f.Storage[appName]; !found {
		return nil, app.ErrNotFound
	}

	return []*ReplicaSetListItem{}, nil
}

func (f *FakeOperations) Deploy(user *database.User, appName string, tarBall io.ReadSeeker, description string, opts *Options) (io.ReadCloser, error) {
	return nil, nil
}

func (f *FakeOperations) Rollback(user *database.User, appName, revision string) error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if !hasPerm(user.Email) {
		return auth.ErrPermissionDenied
	}

	if _, found := f.Storage[appName]; !found {
		return app.ErrNotFound
	}

	return nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{mutex: &sync.RWMutex{}, Storage: make(map[string]bool)}
}
