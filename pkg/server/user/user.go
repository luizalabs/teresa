package user

import (
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

type Operations interface {
	Login(email, password string) (string, error)
}

type DatabaseOperations struct {
	DB   *gorm.DB
	auth auth.Auth
}

func (dbu *DatabaseOperations) Login(email, password string) (string, error) {
	u := new(storage.User)
	if dbu.DB.Where(&storage.User{Email: email}).First(u).RecordNotFound() {
		log.WithField("user", email).Info("unauthorized")
		return "", auth.ErrPermissionDenied
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
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

func NewDatabaseOperations(db *gorm.DB, a auth.Auth) Operations {
	db.AutoMigrate(&storage.User{})
	return &DatabaseOperations{DB: db, auth: a}
}
