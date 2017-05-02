package handlers

import (
	"fmt"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/teresa-api/k8s"
	"github.com/luizalabs/teresa-api/models"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/restapi/operations/users"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserHandler ...
func CreateUserHandler(params users.CreateUserParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	// need to have admin permissions to do this action
	if !*tk.IsAdmin {
		log.Printf("User [%s] doesn't have permission to create a user", *tk.Email)
		return users.NewCreateUserUnauthorized()
	}

	h, err := bcrypt.GenerateFromPassword([]byte(*params.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return users.NewCreateUserDefault(500) // FIXME: better handling
	}
	hashedPassword := string(h)
	u := models.User{
		Name:     params.Body.Name,
		Email:    params.Body.Email,
		Password: &hashedPassword,
		IsAdmin:  params.Body.IsAdmin,
	}
	su := storage.User{
		Name:     *u.Name,
		Email:    *u.Email,
		Password: *u.Password,
		IsAdmin:  *u.IsAdmin,
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

// GetCurrentUserHandler ...
func GetCurrentUserHandler(principal interface{}) middleware.Responder {
	r := users.NewGetCurrentUserOK()
	return r
}

// DeleteUserHandler ...
func DeleteUserHandler(params users.DeleteUserParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	// need to have admin permissions to do this action
	if *tk.IsAdmin == false {
		log.Printf("User [%s] doesn't have permission to delete a user", *tk.Email)
		return users.NewDeleteUserDefault(401)
	}

	su := storage.User{}
	su.ID = uint(params.UserID)
	if storage.DB.Delete(&su).Error != nil {
		return users.NewGetUsersDefault(500)
	}
	return users.NewDeleteUserNoContent()
}
