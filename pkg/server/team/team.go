package team

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/luizalabs/teresa/pkg/server/user"
	"github.com/pkg/errors"
)

type Operations interface {
	Create(name, email, url string) error
	AddUser(name, userEmail string) error
	List() ([]*database.Team, error)
	ListByUser(userEmail string) ([]*database.Team, error)
	RemoveUser(name, userEmail string) error
}

type DatabaseOperations struct {
	DB      *gorm.DB
	UserOps user.Operations
}

func (dbt *DatabaseOperations) Create(name, email, url string) error {
	t := new(database.Team)
	if !dbt.DB.Where(&database.Team{Name: name}).First(t).RecordNotFound() {
		return ErrTeamAlreadyExists
	}

	t.Name = name
	t.Email = email
	t.URL = url

	if err := dbt.DB.Save(t).Error; err != nil {
		return teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, fmt.Sprintf("saving team %s", name)),
		)
	}
	return nil
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

	usersOfTeam := []database.User{}
	dbt.DB.Model(t).Association("Users").Find(&usersOfTeam)
	for _, userOfTeam := range usersOfTeam {
		if userOfTeam.Email == userEmail {
			return ErrUserAlreadyInTeam
		}
	}

	return dbt.DB.Model(t).Association("Users").Append(u).Error
}

func (dbt *DatabaseOperations) List() ([]*database.Team, error) {
	var teams []*database.Team
	if err := dbt.DB.Find(&teams).Error; err != nil {
		return nil, teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, "finding teams"),
		)
	}

	if err := dbt.findTeamUsers(teams); err != nil {
		return nil, err
	}
	return teams, nil
}

func (dbt *DatabaseOperations) ListByUser(userEmail string) ([]*database.Team, error) {
	u, err := dbt.UserOps.GetUser(userEmail)
	if err != nil {
		return nil, err
	}

	var teams []*database.Team
	if err = dbt.DB.Model(u).Association("Teams").Find(&teams).Error; err != nil {
		return nil, teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, fmt.Sprintf("finding teams of user %s", userEmail)),
		)
	}

	if err = dbt.findTeamUsers(teams); err != nil {
		return nil, err
	}
	return teams, nil
}

func (dbt *DatabaseOperations) findTeamUsers(teams []*database.Team) error {
	for _, t := range teams {
		if err := dbt.DB.Model(t).Association("Users").Find(&t.Users).Error; err != nil {
			return teresa_errors.New(
				teresa_errors.ErrInternalServerError,
				errors.Wrap(err, fmt.Sprintf("associating team %s with its users", t.Name)),
			)
		}
	}
	return nil
}

func (dbt *DatabaseOperations) getTeam(name string) (*database.Team, error) {
	t := new(database.Team)
	if dbt.DB.Where(&database.Team{Name: name}).First(t).RecordNotFound() {
		return nil, ErrNotFound
	}
	return t, nil
}

func (dbt *DatabaseOperations) RemoveUser(name, userEmail string) error {
	return nil
}

func NewDatabaseOperations(db *gorm.DB, uOps user.Operations) Operations {
	db.AutoMigrate(&database.Team{})
	return &DatabaseOperations{DB: db, UserOps: uOps}
}
