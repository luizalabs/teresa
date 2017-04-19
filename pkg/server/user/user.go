package user

import (
	"crypto/rsa"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/common"
)

type Operations interface {
	Login(email, password string) (string, error)
}

type DatabaseOperations struct {
	DB      *gorm.DB
	signKey *rsa.PrivateKey
}

func (dbu *DatabaseOperations) Login(email, password string) (string, error) {
	u := new(storage.User)
	if dbu.DB.Where(&storage.User{Email: email}).First(u).RecordNotFound() {
		log.WithField("user", email).Info("unauthorized")
		return "", common.ErrPermissionDenied
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		log.Printf("Authentication failed for user [%s]\n", email)
		return "", common.ErrPermissionDenied
	}
	log.Printf("Authentication succeeded for user [%s]\n", email)

	token, err := dbu.generateUserToken(email)
	if err != nil {
		log.WithError(err).Error("Failed to sign JWT token")
		return "", common.ErrPermissionDenied
	}
	return token, nil
}

func (dbu *DatabaseOperations) generateUserToken(email string) (string, error) {
	jwtClaims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24 * 14).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	return token.SignedString(dbu.signKey)
}

func NewDatabaseOperations(db *gorm.DB, signKey *rsa.PrivateKey) Operations {
	db.AutoMigrate(&storage.User{})
	return &DatabaseOperations{DB: db, signKey: signKey}
}
