package app

import (
	"fmt"
	"io"
	"math/rand"
	"sync"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

type FakeOperations struct {
	mutex   *sync.RWMutex
	Storage map[string]*App
}

func hasPerm(email string) bool {
	return email != "bad-user@luizalabs.com"
}

func (f *FakeOperations) HasPermission(user *storage.User, appName string) bool {
	return hasPerm(user.Email)
}

func (f *FakeOperations) Create(user *storage.User, app *App) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !hasPerm(user.Email) {
		return auth.ErrPermissionDenied
	}
	if _, found := f.Storage[app.Name]; found {
		return ErrAlreadyExists
	}

	f.Storage[app.Name] = app
	return nil
}

func (f *FakeOperations) Logs(user *storage.User, appName string, lines int64, follow bool) (io.ReadCloser, error) {
	if _, found := f.Storage[appName]; !found {
		return nil, ErrNotFound
	}

	if !hasPerm(user.Email) {
		return nil, auth.ErrPermissionDenied
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		for i := 0; int64(i) < lines; i++ {
			fmt.Fprintf(w, "line %d of log\n", i)
		}
		if follow {
			rand.Seed(42) // The Answser
			for i := 0; i <= rand.Intn(5); i++ {
				fmt.Fprintln(w, "extra random lines")
			}
		}
	}()
	return r, nil
}

func (f *FakeOperations) Info(user *storage.User, appName string) (*Info, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !hasPerm(user.Email) {
		return nil, newAppErr(auth.ErrPermissionDenied, fmt.Errorf("error"))
	}

	if _, found := f.Storage[appName]; !found {
		return nil, newAppErr(ErrNotFound, fmt.Errorf("error"))
	}

	return &Info{}, nil
}

func (f *FakeOperations) TeamName(appName string) (string, error) {
	return "luizalabs", nil
}

func (f *FakeOperations) Get(appName string) (*App, error) {
	a := &App{
		Name:        "teresa",
		ProcessType: "web",
		EnvVars: []*EnvVar{
			&EnvVar{Key: "KEY", Value: "Value"},
		},
	}
	return a, nil
}

func (f *FakeOperations) SetEnv(user *storage.User, appName string, envVars []*EnvVar) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !hasPerm(user.Email) {
		return auth.ErrPermissionDenied
	}

	if _, found := f.Storage[appName]; !found {
		return ErrNotFound
	}

	return nil
}

func (f *FakeOperations) UnsetEnv(user *storage.User, appName string, envVars []string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !hasPerm(user.Email) {
		return auth.ErrPermissionDenied
	}

	if _, found := f.Storage[appName]; !found {
		return ErrNotFound
	}

	return nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{
		mutex:   &sync.RWMutex{},
		Storage: make(map[string]*App),
	}
}
