package user

import (
	"sync"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

type FakeOperations struct {
	mutex   *sync.RWMutex
	Storage map[string]*storage.User
}

func (f *FakeOperations) Login(email, password string) (string, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if user, ok := f.Storage[email]; !ok || user.Password != password {
		return "", auth.ErrPermissionDenied
	}
	return "good token", nil
}

func (f *FakeOperations) GetUser(email string) (*storage.User, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	user, found := f.Storage[email]
	if !found {
		return nil, ErrNotFound
	}
	return &storage.User{Email: user.Email, Password: user.Password}, nil
}

func (f *FakeOperations) SetPassword(email, newPassword string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, found := f.Storage[email]; !found {
		return ErrNotFound
	}
	f.Storage[email] = &storage.User{Password: newPassword, Email: email}
	return nil
}

func (f *FakeOperations) Delete(email string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, found := f.Storage[email]; !found {
		return ErrNotFound
	}
	delete(f.Storage, email)
	return nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{
		mutex:   &sync.RWMutex{},
		Storage: make(map[string]*storage.User)}
}
