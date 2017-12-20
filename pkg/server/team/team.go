package team

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teamext"
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
	Rename(oldName, newName string) error
	SetTeamExt(ext teamext.TeamExt)
	Contains(name, userEmail string) (bool, error)
}

type DatabaseOperations struct {
	DB      *gorm.DB
	UserOps user.Operations
	Ext     teamext.TeamExt
}

func (dbt *DatabaseOperations) Create(name, email, url string) error {
	t := new(database.Team)
	if !dbt.DB.Where(&database.Team{Name: name}).First(t).RecordNotFound() {
		return ErrTeamAlreadyExists
	}

	t.Name = name
	t.Email = email
	t.URL = url

	return dbt.save(t)
}

func (dbt *DatabaseOperations) save(t *database.Team) error {
	if err := dbt.DB.Save(t).Error; err != nil {
		return teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, fmt.Sprintf("saving team %s", t.Name)),
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

	// FIXME: Two call for method `dbt.getTeam`
	if ait, err := dbt.Contains(name, userEmail); err != nil || ait {
		if err != nil {
			return err
		}
		return ErrUserAlreadyInTeam
	}

	return dbt.DB.Model(t).Association("Users").Append(u).Error
}

func (dbt *DatabaseOperations) Contains(name, userEmail string) (bool, error) {
	t, err := dbt.getTeam(name)
	if err != nil {
		return false, err
	}

	usersOfTeam := []database.User{}
	if err := dbt.DB.Model(t).Association("Users").Find(&usersOfTeam).Error; err != nil {
		return false, err
	}

	for _, userOfTeam := range usersOfTeam {
		if userOfTeam.Email == userEmail {
			return true, nil
		}
	}

	return false, nil
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
	t, err := dbt.getTeam(name)
	if err != nil {
		return err
	}

	u, err := dbt.UserOps.GetUser(userEmail)
	if err != nil {
		return err
	}

	if err := dbt.DB.Model(t).Association("Users").Find(u).Error; err != nil {
		return ErrUserNotInTeam
	}

	if err := dbt.DB.Model(t).Association("Users").Delete(u).Error; err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func (dbt *DatabaseOperations) Rename(oldName, newName string) error {
	t := new(database.Team)
	if !dbt.DB.Where(&database.Team{Name: newName}).First(t).RecordNotFound() {
		return ErrTeamAlreadyExists
	}

	t, err := dbt.getTeam(oldName)
	if err != nil {
		return err
	}

	t.Name = newName
	if err = dbt.save(t); err != nil {
		return err
	}

	apps, err := dbt.Ext.ListByTeam(oldName)
	if err != nil {
		return err
	}

	for _, a := range apps {
		if err = dbt.Ext.ChangeTeam(a, newName); err != nil {
			return err
		}
	}
	return nil
}

func (dbt *DatabaseOperations) SetTeamExt(ext teamext.TeamExt) {
	dbt.Ext = ext
}

func NewDatabaseOperations(db *gorm.DB, uOps user.Operations) Operations {
	db.AutoMigrate(&database.Team{})
	return &DatabaseOperations{DB: db, UserOps: uOps}
}
