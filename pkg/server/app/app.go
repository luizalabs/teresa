package app

import (
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/team"
)

type Operations interface {
	Create(user *storage.User, app *App) error
}

type K8sOperations interface {
	Create(app *App, st st.Storage) error
}

type AppOperations struct {
	tops team.Operations
	kops K8sOperations
	st   st.Storage
}

func (ops *AppOperations) hasPerm(user *storage.User, app *App) bool {
	teams, err := ops.tops.ListByUser(user.Email)
	if err != nil {
		return false
	}
	var found bool
	for _, team := range teams {
		if team.Name == app.Team {
			found = true
			break
		}
	}
	return found
}

func (ops *AppOperations) Create(user *storage.User, app *App) error {
	if !ops.hasPerm(user, app) {
		return auth.ErrPermissionDenied
	}
	return ops.kops.Create(app, ops.st)
}

func NewAppOperations(tops team.Operations, kops K8sOperations, st st.Storage) Operations {
	return &AppOperations{tops: tops, kops: kops, st: st}
}
