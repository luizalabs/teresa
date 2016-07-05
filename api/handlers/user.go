package handlers

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/users"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserHandler ...
func CreateUserHandler(params users.CreateUserParams, principal interface{}) middleware.Responder {

	h, err := bcrypt.GenerateFromPassword([]byte(*params.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return users.NewCreateUserDefault(500) // FIXME: better handling
	}
	hashedPassword := string(h)
	u := models.User{
		Name:     params.Body.Name,
		Email:    params.Body.Email,
		Password: &hashedPassword,
	}
	su := storage.User{
		Name:     *u.Name,
		Email:    *u.Email,
		Password: *u.Password,
	}
	err = storage.DB.Create(&su).Error
	if err != nil {
		fmt.Printf("UsersCreateUserHandler failed: %s\n", err)
		return users.NewCreateUserDefault(422)
	}
	u.ID = int64(su.ID)
	u.Password = nil
	r := users.NewCreateUserCreated()
	r.SetPayload(&u)
	return r
}

// GetUserDetailsHandler ...
func GetUserDetailsHandler(params users.GetUserDetailsParams, principal interface{}) middleware.Responder {
	su := storage.User{}
	su.ID = uint(params.UserID)
	if storage.DB.First(&su).RecordNotFound() {
		fmt.Printf("No user with ID [%d] found\n", params.UserID)
		return users.NewGetUserDetailsNotFound()
	}

	fmt.Printf("Found user with ID [%d] name [%s] email [%s]\n", su.ID, su.Name, su.Email)
	r := users.NewGetUserDetailsOK()
	u := models.User{
		ID:    int64(su.ID),
		Name:  &su.Name,
		Email: &su.Email,
	}
	r.SetPayload(&u)
	return r
}
