package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/luizalabs/paas/api/restapi/operations"
	"github.com/luizalabs/paas/api/restapi/operations/apps"
	"github.com/luizalabs/paas/api/restapi/operations/auth"
	"github.com/luizalabs/paas/api/restapi/operations/deployments"
	"github.com/luizalabs/paas/api/restapi/operations/teams"
	"github.com/luizalabs/paas/api/restapi/operations/users"

	"github.com/luizalabs/paas/api/models"
)

// This file is safe to edit. Once it exists it will not be overwritten

func configureFlags(api *operations.TeresaAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.TeresaAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	api.MultipartformConsumer = runtime.DiscardConsumer

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.AppsCreateAppHandler = apps.CreateAppHandlerFunc(func(params apps.CreateAppParams) middleware.Responder {
		return middleware.NotImplemented("operation apps.CreateApp has not yet been implemented")
	})
	api.DeploymentsCreateDeploymentHandler = deployments.CreateDeploymentHandlerFunc(func(params deployments.CreateDeploymentParams) middleware.Responder {
		return middleware.NotImplemented("operation deployments.CreateDeployment has not yet been implemented")
	})
	api.TeamsCreateTeamHandler = teams.CreateTeamHandlerFunc(func(params teams.CreateTeamParams) middleware.Responder {
		return middleware.NotImplemented("operation teams.CreateTeam has not yet been implemented")
	})
	api.UsersCreateUserHandler = users.CreateUserHandlerFunc(func(params users.CreateUserParams) middleware.Responder {
		return middleware.NotImplemented("operation users.CreateUser has not yet been implemented")
	})
	api.AppsGetAppDetailsHandler = apps.GetAppDetailsHandlerFunc(func(params apps.GetAppDetailsParams) middleware.Responder {
		return middleware.NotImplemented("operation apps.GetAppDetails has not yet been implemented")
	})
	api.AppsGetAppsHandler = apps.GetAppsHandlerFunc(func(params apps.GetAppsParams) middleware.Responder {
		return middleware.NotImplemented("operation apps.GetApps has not yet been implemented")
	})
	api.UsersGetCurrentUserHandler = users.GetCurrentUserHandlerFunc(func() middleware.Responder {
		return middleware.NotImplemented("operation users.GetCurrentUser has not yet been implemented")
	})
	api.DeploymentsGetDeploymentsHandler = deployments.GetDeploymentsHandlerFunc(func(params deployments.GetDeploymentsParams) middleware.Responder {
		return middleware.NotImplemented("operation deployments.GetDeployments has not yet been implemented")
	})
	api.TeamsGetTeamDetailHandler = teams.GetTeamDetailHandlerFunc(func(params teams.GetTeamDetailParams) middleware.Responder {
		return middleware.NotImplemented("operation teams.GetTeamDetail has not yet been implemented")
	})
	api.TeamsGetTeamsHandler = teams.GetTeamsHandlerFunc(func(params teams.GetTeamsParams) middleware.Responder {
		return middleware.NotImplemented("operation teams.GetTeams has not yet been implemented")
	})
	api.UsersGetUserDetailsHandler = users.GetUserDetailsHandlerFunc(func(params users.GetUserDetailsParams) middleware.Responder {
		return middleware.NotImplemented("operation users.GetUserDetails has not yet been implemented")
	})
	api.UsersGetUsersHandler = users.GetUsersHandlerFunc(func(params users.GetUsersParams) middleware.Responder {
		return middleware.NotImplemented("operation users.GetUsers has not yet been implemented")
	})
	api.AppsUpdateAppHandler = apps.UpdateAppHandlerFunc(func(params apps.UpdateAppParams) middleware.Responder {
		return middleware.NotImplemented("operation apps.UpdateApp has not yet been implemented")
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
