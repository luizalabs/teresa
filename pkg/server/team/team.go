package team

import (
	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/user"
)

type Operations interface {
	Create(name, email, url string) error
	AddUser(name, userEmail string) error
}

type DatabaseOperations struct {
	DB      *gorm.DB
	UserOps user.Operations
}

func (dbt *DatabaseOperations) Create(name, email, url string) error {
	t := new(storage.Team)
	if !dbt.DB.Where(&storage.Team{Name: name}).First(t).RecordNotFound() {
		return ErrTeamAlreadyExists
	}

	t.Name = name
	t.Email = email
	t.URL = url
	return dbt.DB.Save(t).Error
}

func (dbt *DatabaseOperations) AddUser(name, userEmail string) error {
	t, err := dbt.getTeam(name)
	if err != nil {
		return err
	}
	u, err := dbt.UserOps.GetUser(userEmail)
	if err != nil {
		return err
	}

	usersOfTeam := []storage.User{}
	dbt.DB.Model(t).Association("Users").Find(&usersOfTeam)
	for _, userOfTeam := range usersOfTeam {
		if userOfTeam.Email == userEmail {
			return ErrUserAlreadyInTeam
		}
	}

	return dbt.DB.Model(t).Association("Users").Append(u).Error
}

func (dbt *DatabaseOperations) getTeam(name string) (*storage.Team, error) {
	t := new(storage.Team)
	if dbt.DB.Where(&storage.Team{Name: name}).First(t).RecordNotFound() {
		return nil, ErrNotFound
	}
	return t, nil
}

func NewDatabaseOperations(db *gorm.DB, uOps user.Operations) Operations {
	db.AutoMigrate(&storage.Team{})
	return &DatabaseOperations{DB: db, UserOps: uOps}
}
