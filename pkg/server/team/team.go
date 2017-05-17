package team

import (
	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
)

type Operations interface {
	Create(name, email, url string) error
	AddUser(name, email string) error
}

type DatabaseOperations struct {
	DB *gorm.DB
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

func (dbt *DatabaseOperations) AddUser(name, email string) error {
	return nil
}

func NewDatabaseOperations(db *gorm.DB) Operations {
	db.AutoMigrate(&storage.Team{})
	return &DatabaseOperations{DB: db}
}
