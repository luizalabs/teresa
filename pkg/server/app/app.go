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

type K8sOperations interface {
	Create(app *App, st st.Storage) error
}

type AppOperations struct {
	DB   *gorm.DB
	kops K8sOperations
	st   st.Storage
}

func (ops *AppOperations) Create(user *storage.User, app *App) error {
	var teams []*storage.Team
	err := ops.DB.Model(user).Association("Teams").Find(&teams).Error
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
	return ops.kops.Create(app, ops.st)
}

func NewAppOperations(db *gorm.DB, kops K8sOperations, st st.Storage) Operations {
	return &AppOperations{DB: db, kops: kops, st: st}
}
