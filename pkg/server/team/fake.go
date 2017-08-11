package team

import (
	"sync"

	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/user"
)

type FakeOperations struct {
	mutex   *sync.RWMutex
	Storage map[string]*database.Team

	UserOps user.Operations
}

func (f *FakeOperations) Create(name, email, url string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, found := f.Storage[name]; found {
		return ErrTeamAlreadyExists
	}

	f.Storage[name] = &database.Team{Name: name, Email: email, URL: url}
	return nil
}

func (f *FakeOperations) AddUser(name, userEmail string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	t, found := f.Storage[name]
	if !found {
		return ErrNotFound
	}

	u, err := f.UserOps.GetUser(userEmail)
	if err != nil {
		return err
	}

	for _, userOfTeam := range t.Users {
		if userOfTeam.Email == userEmail {
			return ErrUserAlreadyInTeam
		}
	}

	t.Users = append(t.Users, *u)
	return nil
}

func (f *FakeOperations) List() ([]*database.Team, error) {
	var teams []*database.Team
	for _, v := range f.Storage {
		teams = append(teams, v)
	}
	return teams, nil
}

func (f *FakeOperations) ListByUser(userEmail string) ([]*database.Team, error) {
	var teams []*database.Team
	for _, v := range f.Storage {
		for _, u := range v.Users {
			if u.Email == userEmail {
				teams = append(teams, v)
			}
		}
	}
	return teams, nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{
		mutex:   &sync.RWMutex{},
		Storage: make(map[string]*database.Team),
		UserOps: user.NewFakeOperations()}
}
