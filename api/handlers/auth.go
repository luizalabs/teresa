package handlers

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/astaxie/beego/orm"
	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/paas/api/models"
	storage "github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/auth"
)

func LoginHandler(params auth.UserLoginParams) middleware.Responder {
	o := orm.NewOrm()
	o.Using("default")
	su := storage.User{Email: params.Body.Email.String()}
	err := o.Read(&su, "Email")
	if err == orm.ErrNoRows {
		fmt.Printf("Login unauthorized for user: [%s]\n", params.Body.Email)
		return auth.NewUserLoginUnauthorized()
	} else {
		err = bcrypt.CompareHashAndPassword([]byte(su.Password), []byte(*params.Body.Password))
		if err != nil {
			fmt.Printf("Login unauthorized for user: [%s]\n", params.Body.Email)
			return auth.NewUserLoginUnauthorized()
		}
		fmt.Printf("Login OK for user: %s\n", params.Body.Email)
		r := auth.NewUserLoginOK()
		u := models.User{
			ID:    su.Id,
			Name:  &su.Name,
			Email: &su.Email,
		}
		r.SetPayload(&u)
		return r
	}
}
