package user

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

const (
	minPassLength = 8
)

type Operations interface {
	Login(email, password string) (string, error)
	GetUser(email string) (*storage.User, error)
	SetPassword(email, newPassword string) error
	Delete(email string) error
	Create(name, email, pass string, admin bool) error
}

type DatabaseOperations struct {
	DB   *gorm.DB
	auth auth.Auth
}

func (dbu *DatabaseOperations) Login(email, password string) (string, error) {
	u, err := dbu.GetUser(email)
	if err != nil {
		log.WithField("user", email).Info("Not Found")
		return "", auth.ErrPermissionDenied
	}
	if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		log.Printf("Authentication failed for user [%s]\n", email)
		return "", auth.ErrPermissionDenied
	}
	log.Printf("Authentication succeeded for user [%s]\n", email)

	token, err := dbu.auth.GenerateToken(email)
	if err != nil {
		log.WithError(err).Error("Failed to sign JWT token")
		return "", auth.ErrPermissionDenied
	}
	return token, nil
}

func (dbu *DatabaseOperations) GetUser(email string) (*storage.User, error) {
	u := new(storage.User)
	if dbu.DB.Where(&storage.User{Email: email}).First(u).RecordNotFound() {
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
		return err
	}
	u.Password = string(pass)
	return dbu.DB.Save(u).Error
}

func (dbu *DatabaseOperations) Delete(email string) error {
	u, err := dbu.GetUser(email)
	if err != nil {
		return err
	}
	return dbu.DB.Delete(u).Error
}

func (dbu *DatabaseOperations) Create(name, email, pass string, admin bool) error {
	u := new(storage.User)
	if !dbu.DB.Where(&storage.User{Email: email}).First(u).RecordNotFound() {
		return ErrUserAlreadyExists
	}
	if len(pass) < minPassLength {
		return ErrInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Name = name
	u.Email = email
	u.Password = string(hash)
	u.IsAdmin = admin
	return dbu.DB.Save(u).Error
}

func NewDatabaseOperations(db *gorm.DB, a auth.Auth) Operations {
	db.AutoMigrate(&storage.User{})
	return &DatabaseOperations{DB: db, auth: a}
}
