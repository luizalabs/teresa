package handlers

import (
	"crypto/rsa"
	"io/ioutil"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/teresa-api/k8s"
	"github.com/luizalabs/teresa-api/models"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/restapi/operations/auth"
)

// location of the files used for signing and verification
const (
	privKeyPath = "keys/teresa.rsa"     // openssl genrsa -out teresa.rsa keysize
	pubKeyPath  = "keys/teresa.rsa.pub" // openssl rsa -in teresa.rsa -pubout > teresa.rsa.pub
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

// read the key files before starting http handlers
func init() {
	signBytes, err := ioutil.ReadFile(privKeyPath)
	fatal(err)

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	fatal(err)

	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	fatal(err)

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	fatal(err)
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// LoginHandler validates a user
func LoginHandler(params auth.UserLoginParams) middleware.Responder {
	su := storage.User{}
	if storage.DB.Where(&storage.User{Email: params.Body.Email.String()}).First(&su).RecordNotFound() {
		log.Printf("Login unauthorized for user: [%s]\n", params.Body.Email)
		return auth.NewUserLoginUnauthorized()
	}
	p := params.Body.Password.String()
	err := su.Authenticate(&p)
	if err != nil {
		return auth.NewUserLoginUnauthorized()
	}

	jwtClaims := jwt.MapClaims{
		"email": su.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 14).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(signKey)
	if err != nil {
		log.Printf("Failed to sign jwt token, err: %s\n", err)
		return auth.NewUserLoginDefault(500)
	}
	r := auth.NewUserLoginOK()
	t := models.LoginToken{Token: tokenString}
	r.SetPayload(&t)
	return r
}

// TokenAuthHandler try and validate the jwt token
func TokenAuthHandler(t string) (interface{}, error) {
	token, err := jwt.ParseWithClaims(t, &k8s.Token{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	// branch out into the possible error from signing
	switch err.(type) {
	case nil: // no error
		if !token.Valid { // but may still be invalid
			log.Printf("JWT token validation - invalid token received: %+v\n", token)
			return nil, errors.Unauthenticated("Invalid credentials")
		}
		// see stdout and watch for the CustomUserInfo, nicely unmarshalled
		log.Printf("JWT token validation - granting access with token: %+v", token)
		tc, _ := token.Claims.(*k8s.Token)

		err := k8s.Client.Users().LoadUserToToken(tc)
		if err != nil {
			log.Printf(`error when trying to load user "%s" informations for the token`, *tc.Email)
		}
		return tc, nil
	case *jwt.ValidationError: // something was wrong during the validation
		vErr := err.(*jwt.ValidationError)

		switch vErr.Errors {
		case jwt.ValidationErrorExpired:
			log.Printf("JWT token validation - token expired: %+v\n", token)
			return nil, errors.Unauthenticated("Invalid credentials")
		default:
			log.Printf("JWT token validation - ValidationError error on token: %+v\n", token)
			return nil, errors.Unauthenticated("Invalid credentials")
		}
	default: // something else went wrong
		log.Printf("JWT token validation - parse error: %v\n", err)
		return nil, errors.Unauthenticated("Invalid credentials")
	}
}
