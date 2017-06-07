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
	CreateNamespace(*App, string) error
	CreateQuota(*App) error
	CreateSecret(string, string, map[string][]byte) error
	CreateAutoScale(*App) error
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

	if err := ops.kops.CreateNamespace(app, user.Email); err != nil {
		return err
	}

	if err := ops.kops.CreateQuota(app); err != nil {
		return err
	}

	secretName := ops.st.K8sSecretName()
	data := ops.st.AccessData()
	if err := ops.kops.CreateSecret(app.Name, secretName, data); err != nil {
		return err
	}

	return ops.kops.CreateAutoScale(app)
}

func NewAppOperations(tops team.Operations, kops K8sOperations, st st.Storage) Operations {
	return &AppOperations{tops: tops, kops: kops, st: st}
}
