package user

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/luizalabs/teresa/pkg/server/validations"
)

const (
	minPassLength = 8
)

type Operations interface {
	Login(email, password string, exp time.Duration) (string, error)
	GetUser(email string) (*database.User, error)
	SetPassword(email, newPassword string) error
	Delete(email string) error
	Create(name, email, pass string, admin bool) error
}

type DatabaseOperations struct {
	DB   *gorm.DB
	auth auth.Auth
}

func (dbu *DatabaseOperations) Login(email, password string, exp time.Duration) (string, error) {
	u, err := dbu.GetUser(email)
	if err != nil {
		return "", auth.ErrPermissionDenied
	}
	if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", teresa_errors.New(
			auth.ErrPermissionDenied,
			errors.Wrap(err, fmt.Sprintf("Authentication failed for user %s", email)),
		)
	}

	token, err := dbu.auth.GenerateToken(email, exp)
	if err != nil {
		return "", teresa_errors.New(
			auth.ErrPermissionDenied,
			errors.Wrap(err, "Signing JWT token"),
		)
	}
	return token, nil
}

func (dbu *DatabaseOperations) GetUser(email string) (*database.User, error) {
	u := new(database.User)
	if dbu.DB.Where(&database.User{Email: email}).First(u).RecordNotFound() {
		return nil, ErrNotFound
	}
	return u, nil
}

func (dbu *DatabaseOperations) SetPassword(email, newPassword string) error {
	u, err := dbu.GetUser(email)
	if err != nil {
		return err
	}
	pass, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, fmt.Sprintf("Generating the password hash to user %s", email)),
		)
	}
	u.Password = string(pass)
	if err = dbu.DB.Save(u).Error; err != nil {
		return teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, fmt.Sprintf("Updating password of user %s", email)),
		)
	}
	return nil
}

func (dbu *DatabaseOperations) Delete(email string) error {
	u, err := dbu.GetUser(email)
	if err != nil {
		return err
	}
	if err = dbu.DB.Delete(u).Error; err != nil {
		return teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, fmt.Sprintf("Deleting user %s", email)),
		)
	}
	return nil
}

func (dbu *DatabaseOperations) Create(name, email, pass string, admin bool) error {
	if !validations.ValidateEmail(email) {
		return ErrInvalidEmail
	}
	if len(pass) < minPassLength {
		return ErrInvalidPassword
	}

	u := new(database.User)
	if !dbu.DB.Where(&database.User{Email: email}).First(u).RecordNotFound() {
		return ErrUserAlreadyExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, fmt.Sprintf("Generating the password hash to user %s", email)),
		)
	}
	u.Name = name
	u.Email = email
	u.Password = string(hash)
	u.IsAdmin = admin
	if err = dbu.DB.Save(u).Error; err != nil {
		return teresa_errors.New(
			teresa_errors.ErrInternalServerError,
			errors.Wrap(err, fmt.Sprintf("Creating user %s", email)),
		)
	}
	return nil
}

func NewDatabaseOperations(db *gorm.DB, a auth.Auth) Operations {
	db.AutoMigrate(&database.User{})
	return &DatabaseOperations{DB: db, auth: a}
}
