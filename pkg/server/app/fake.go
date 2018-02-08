package app

import (
	"fmt"
	"io"
	"math/rand"
	"sync"

	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

type FakeOperations struct {
	mutex   *sync.RWMutex
	Storage map[string]*App
}

func hasPerm(email string) bool {
	return email != "bad-user@luizalabs.com"
}

func (f *FakeOperations) HasPermission(user *database.User, appName string) bool {
	return hasPerm(user.Email)
}

func (f *FakeOperations) Create(user *database.User, app *App) error {
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

func (f *FakeOperations) Logs(user *database.User, appName string, opts *LogOptions) (io.ReadCloser, error) {
	if _, found := f.Storage[appName]; !found {
		return nil, ErrNotFound
	}

	if !hasPerm(user.Email) {
		return nil, auth.ErrPermissionDenied
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		for i := 0; int64(i) < opts.Lines; i++ {
			fmt.Fprintf(w, "line %d of log\n", i)
		}
		if opts.Follow {
			rand.Seed(42) // The Answser
			for i := 0; i <= rand.Intn(5); i++ {
				fmt.Fprintln(w, "extra random lines")
			}
		}
	}()
	return r, nil
}

func (f *FakeOperations) Info(user *database.User, appName string) (*Info, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !hasPerm(user.Email) {
		return nil, teresa_errors.New(auth.ErrPermissionDenied, fmt.Errorf("error"))
	}

	if _, found := f.Storage[appName]; !found {
		return nil, teresa_errors.New(ErrNotFound, fmt.Errorf("error"))
	}

	return &Info{}, nil
}

func (f *FakeOperations) List(user *database.User) ([]*AppListItem, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	items := make([]*AppListItem, 0)
	for k, v := range f.Storage {
		items = append(items, &AppListItem{
			Team:      v.Team,
			Name:      k,
			Addresses: []*Address{{Hostname: "localhost"}},
		})
	}
	return items, nil
}

func (f *FakeOperations) ListByTeam(teamName string) ([]string, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	items := make([]string, 0)
	for k, v := range f.Storage {
		if v.Team == teamName {
			items = append(items, k)
		}
	}

	return items, nil
}

func (f *FakeOperations) Delete(user *database.User, appName string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !hasPerm(user.Email) {
		return auth.ErrPermissionDenied
	}

	if _, found := f.Storage[appName]; !found {
		return ErrNotFound
	}
	delete(f.Storage, appName)

	return nil
}

func (f *FakeOperations) TeamName(appName string) (string, error) {
	return "luizalabs", nil
}

func (f *FakeOperations) Get(appName string) (*App, error) {
	if appName != "teresa" {
		return nil, ErrNotFound
	}

	a := &App{
		Name:        "teresa",
		ProcessType: "web",
		EnvVars: []*EnvVar{
			{Key: "KEY", Value: "Value"},
		},
	}
	return a, nil
}

func (f *FakeOperations) SetEnv(user *database.User, appName string, envVars []*EnvVar) error {
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

func (f *FakeOperations) UnsetEnv(user *database.User, appName string, envVars []string) error {
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

func (f *FakeOperations) SetAutoscale(user *database.User, appName string, as *Autoscale) error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if !hasPerm(user.Email) {
		return auth.ErrPermissionDenied
	}

	if _, found := f.Storage[appName]; !found {
		return ErrNotFound
	}

	return nil
}

func (f *FakeOperations) CheckPermAndGet(user *database.User, appName string) (*App, error) {
	if !hasPerm(user.Email) {
		return nil, auth.ErrPermissionDenied
	}

	if appName != "teresa" {
		return nil, ErrNotFound
	}

	a := &App{
		Name:        "teresa",
		ProcessType: "web",
		EnvVars: []*EnvVar{
			{Key: "KEY", Value: "Value"},
		},
	}
	return a, nil
}

func (f *FakeOperations) SaveApp(app *App, lastUser string) error {
	if app.Name != "teresa" {
		return ErrNotFound
	}

	return nil
}

func (f *FakeOperations) SetReplicas(user *database.User, appName string, replicas int32) error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if !hasPerm(user.Email) {
		return auth.ErrPermissionDenied
	}

	if _, found := f.Storage[appName]; !found {
		return ErrNotFound
	}

	return nil
}

func (f *FakeOperations) ChangeTeam(appName, teamName string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Storage[appName].Team = teamName
	return nil
}

func (f *FakeOperations) DeletePods(user *database.User, appName string, podsNames []string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !hasPerm(user.Email) {
		return teresa_errors.New(auth.ErrPermissionDenied, fmt.Errorf("error"))
	}

	if _, found := f.Storage[appName]; !found {
		return teresa_errors.New(ErrNotFound, fmt.Errorf("error"))
	}

	return nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{
		mutex:   &sync.RWMutex{},
		Storage: make(map[string]*App),
	}
}
