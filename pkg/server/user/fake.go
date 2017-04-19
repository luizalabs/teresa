package user

import (
	"sync"

	"github.com/luizalabs/teresa-api/pkg/common"
)

type FakeOperations struct {
	mutex   *sync.RWMutex
	Storage map[string]string
}

func (f *FakeOperations) Login(email, password string) (string, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if pass, ok := f.Storage[email]; !ok || pass != password {
		return "", common.ErrPermissionDenied
	}
	return "good token", nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{
		mutex:   &sync.RWMutex{},
		Storage: make(map[string]string)}
}
