package handlers

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/paas/api/models"
	storage "github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/users"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserHandler(params users.CreateUserParams, principal interface{}) middleware.Responder {
	o := orm.NewOrm()
	o.Using("default")
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
	id, err := o.Insert(&su)
	if err != nil {
		fmt.Printf("UsersCreateUserHandler failed: %s\n", err)
		return users.NewCreateUserDefault(422)
	}
	u.ID = id
	u.Password = nil
	r := users.NewCreateUserCreated()
	r.SetPayload(&u)
	return r
}

func GetUserDetailsHandler(params users.GetUserDetailsParams, principal interface{}) middleware.Responder {
	o := orm.NewOrm()
	o.Using("default")
	su := storage.User{Id: params.UserID}
	err := o.Read(&su)
	if err == orm.ErrNoRows {
		fmt.Println("No result found")
		return users.NewGetUserDetailsNotFound()
	} else if err == orm.ErrMissPK {
		fmt.Printf("No user with ID [%s] found\n", params.UserID)
		return users.NewGetUserDetailsNotFound()
	} else {
		fmt.Printf("Found user with ID [%d] name [%s] email [%s]\n", su.Id, su.Name, su.Email)
		r := users.NewGetUserDetailsOK()
		u := models.User{
			ID:    su.Id,
			Name:  &su.Name,
			Email: &su.Email,
		}
		r.SetPayload(&u)
		return r
	}
}
