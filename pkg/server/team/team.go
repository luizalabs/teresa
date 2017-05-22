package team

import (
	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/user"
)

type Operations interface {
	Create(name, email, url string) error
	AddUser(name, userEmail string) error
	List() ([]*storage.Team, error)
	ListByUser(userEmail string) ([]*storage.Team, error)
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

func (dbt *DatabaseOperations) List() ([]*storage.Team, error) {
	var teams []*storage.Team
	if err := dbt.DB.Find(&teams).Error; err != nil {
		return nil, err
	}

	for _, t := range teams {
		if err := dbt.DB.Model(t).Association("Users").Find(&t.Users).Error; err != nil {
			return nil, err
		}
	}
	return teams, nil
}

func (dbt *DatabaseOperations) ListByUser(userEmail string) ([]*storage.Team, error) {
	u, err := dbt.UserOps.GetUser(userEmail)
	if err != nil {
		return nil, err
	}

	var teams []*storage.Team
	if err = dbt.DB.Model(u).Association("Teams").Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
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
