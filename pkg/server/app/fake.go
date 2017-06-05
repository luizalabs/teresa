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
	if !hasPerm(user.Email) {
		return nil, auth.ErrPermissionDenied
	}

	if _, found := f.Storage[appName]; !found {
		return nil, ErrNotFound
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		for i := 0; int64(i) < lines; i++ {
			fmt.Fprintf(w, "line %d of log\n", i)
		}
		if follow {
			rand.Seed(42) // The Anwser
			for i := 0; i <= rand.Intn(5); i++ {
				fmt.Fprintln(w, "extra random lines")
			}
		}
	}()
	return r, nil
}

func NewFakeOperations() Operations {
	return &FakeOperations{
		mutex:   &sync.RWMutex{},
		Storage: make(map[string]*App),
	}
}
