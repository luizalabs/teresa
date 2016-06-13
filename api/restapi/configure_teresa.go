package restapi

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/luizalabs/paas/api/restapi/operations"
	"github.com/luizalabs/paas/api/restapi/operations/apps"
	"github.com/luizalabs/paas/api/restapi/operations/auth"
	"github.com/luizalabs/paas/api/restapi/operations/deployments"
	"github.com/luizalabs/paas/api/restapi/operations/teams"
	"github.com/luizalabs/paas/api/restapi/operations/users"

	"github.com/astaxie/beego/orm"
	"github.com/luizalabs/paas/api/models"
	storage "github.com/luizalabs/paas/api/models/storage"
)

// This file is safe to edit. Once it exists it will not be overwritten

func configureFlags(api *operations.TeresaAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

// FIXME: hardcoded for test purposes
func hardcodedValidateToken(token string) (v bool) {
	if token == "teresa" {
		return true
	} else {
		return false
	}
}

func configureAPI(api *operations.TeresaAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	api.JSONConsumer = runtime.JSONConsumer()

	api.MultipartformConsumer = runtime.DiscardConsumer

	api.JSONProducer = runtime.JSONProducer()

	api.APIKeyAuth = func(token string) (interface{}, error) {
		if hardcodedValidateToken(token) {
			return token, nil
		}
		return nil, errors.NotImplemented("api key auth (api_key) token from query has not yet been implemented")
	}

	api.TokenHeaderAuth = func(token string) (interface{}, error) {
		if hardcodedValidateToken(token) {
			return token, nil
		}
		return nil, errors.NotImplemented("api key auth (token_header) Authorization from header has not yet been implemented")
	}

	api.AppsCreateAppHandler = apps.CreateAppHandlerFunc(func(params apps.CreateAppParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation apps.CreateApp has not yet been implemented")
	})
	api.DeploymentsCreateDeploymentHandler = deployments.CreateDeploymentHandlerFunc(func(params deployments.CreateDeploymentParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation deployments.CreateDeployment has not yet been implemented")
	})
	api.TeamsCreateTeamHandler = teams.CreateTeamHandlerFunc(func(params teams.CreateTeamParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation teams.CreateTeam has not yet been implemented")
	})

	// create a user
	api.UsersCreateUserHandler = users.CreateUserHandlerFunc(func(params users.CreateUserParams, principal interface{}) middleware.Responder {
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
	})
	api.AppsGetAppDetailsHandler = apps.GetAppDetailsHandlerFunc(func(params apps.GetAppDetailsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation apps.GetAppDetails has not yet been implemented")
	})
	api.AppsGetAppsHandler = apps.GetAppsHandlerFunc(func(params apps.GetAppsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation apps.GetApps has not yet been implemented")
	})
	api.UsersGetCurrentUserHandler = users.GetCurrentUserHandlerFunc(func(principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation users.GetCurrentUser has not yet been implemented")
	})
	api.DeploymentsGetDeploymentsHandler = deployments.GetDeploymentsHandlerFunc(func(params deployments.GetDeploymentsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation deployments.GetDeployments has not yet been implemented")
	})
	api.TeamsGetTeamDetailHandler = teams.GetTeamDetailHandlerFunc(func(params teams.GetTeamDetailParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation teams.GetTeamDetail has not yet been implemented")
	})
	api.TeamsGetTeamsHandler = teams.GetTeamsHandlerFunc(func(params teams.GetTeamsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation teams.GetTeams has not yet been implemented")
	})

	// get a single user from db
	api.UsersGetUserDetailsHandler = users.GetUserDetailsHandlerFunc(func(params users.GetUserDetailsParams, principal interface{}) middleware.Responder {
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
	})

	api.UsersGetUsersHandler = users.GetUsersHandlerFunc(func(params users.GetUsersParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation users.GetUsers has not yet been implemented")
	})
	api.AppsUpdateAppHandler = apps.UpdateAppHandlerFunc(func(params apps.UpdateAppParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation apps.UpdateApp has not yet been implemented")
	})
	api.UsersUpdateUserHandler = users.UpdateUserHandlerFunc(func(params users.UpdateUserParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation users.UpdateUser has not yet been implemented")
	})

	api.AuthUserLoginHandler = auth.UserLoginHandlerFunc(func(params auth.UserLoginParams) middleware.Responder {
		n := "arnaldo"
		e := "arnaldo@luizalabs.com"
		u := models.User{Name: &n, Email: &e}
		uo := auth.NewUserLoginOK()
		uo.SetPayload(&u)
		return uo
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
