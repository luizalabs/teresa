package auth

import (
	"crypto/rsa"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Auth interface {
	GenerateToken(email string) (string, error)
	ValidateToken(token string) (string, error)
}

type tokenClaim struct {
	Email string `json:"email"`
	jwt.StandardClaims
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

func (a *JWTAuth) ValidateToken(token string) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &tokenClaim{}, func(*jwt.Token) (interface{}, error) {
		return a.publicKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return "", ErrPermissionDenied
	}
	claims, ok := parsedToken.Claims.(*tokenClaim)
	if !ok {
		return "", ErrPermissionDenied
	}
	return claims.Email, nil
}

func New(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) Auth {
	return &JWTAuth{privateKey: privateKey, publicKey: publicKey}
}
