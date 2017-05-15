package team

import (
	"sync"

	"github.com/luizalabs/teresa-api/models/storage"
)

type FakeOperations struct {
	mutex   *sync.RWMutex
	Storage map[string]*storage.Team
}

func (f *FakeOperations) Create(name, email, url string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, found := f.Storage[name]; found {
		return ErrTeamAlreadyExists
	}

	f.Storage[name] = &storage.Team{Name: name, Email: email, URL: url}
	return nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{
		mutex:   &sync.RWMutex{},
		Storage: make(map[string]*storage.Team)}
}
