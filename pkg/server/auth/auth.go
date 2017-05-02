package auth

import (
	"crypto/rsa"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Auth interface {
	GenerateToken(email string) (string, error)
}

type JWTAuth struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func (a *JWTAuth) GenerateToken(email string) (string, error) {
	jwtClaims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24 * 14).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	return token.SignedString(a.privateKey)
}

func New(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) Auth {
	return &JWTAuth{privateKey: privateKey, publicKey: publicKey}
}
