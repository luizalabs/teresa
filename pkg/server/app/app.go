package app

import (
	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
)

type Operations interface {
	Create(user *storage.User, app *App) error
}

type AppOperations interface {
	Create(app *App, st st.Storage) error
}

type K8sOperations struct {
	DB  *gorm.DB
	ops AppOperations
	st  st.Storage
}

func (kops *K8sOperations) Create(user *storage.User, app *App) error {
	var teams []*storage.Team
	err := kops.DB.Model(user).Association("Teams").Find(&teams).Error
	if err != nil {
		return err
	}
	var found bool
	for _, team := range teams {
		if team.Name == app.Team {
			found = true
			break
		}
	}
	if !found {
		return auth.ErrPermissionDenied
	}
	return kops.ops.Create(app, kops.st)
}

func NewK8sOperations(db *gorm.DB, ops AppOperations, st st.Storage) Operations {
	return &K8sOperations{DB: db, ops: ops, st: st}
}
