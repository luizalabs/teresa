package user

import (
	"sync"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

type FakeOperations struct {
	mutex   *sync.RWMutex
	Storage map[string]string
}

func (f *FakeOperations) Login(email, password string) (string, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if pass, ok := f.Storage[email]; !ok || pass != password {
		return "", auth.ErrPermissionDenied
	}
	return "good token", nil
}

func (f *FakeOperations) GetUser(email string) (*storage.User, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	pass, ok := f.Storage[email]
	if !ok {
		return nil, ErrNotFound
	}
	return &storage.User{Email: email, Password: pass}, nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{
		mutex:   &sync.RWMutex{},
		Storage: make(map[string]string)}
}
