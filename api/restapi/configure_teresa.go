package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/luizalabs/paas/api/handlers"
	"github.com/luizalabs/paas/api/restapi/operations"
	"github.com/luizalabs/paas/api/restapi/operations/apps"
	"github.com/luizalabs/paas/api/restapi/operations/auth"
	"github.com/luizalabs/paas/api/restapi/operations/deployments"
	"github.com/luizalabs/paas/api/restapi/operations/teams"
	"github.com/luizalabs/paas/api/restapi/operations/users"
)

// This file is safe to edit. Once it exists it will not be overwritten

func configureFlags(api *operations.TeresaAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.TeresaAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// s.api.Logger = log.Printf

	api.MultipartformConsumer = runtime.DiscardConsumer

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// authentication
	api.APIKeyAuth = handlers.TokenAuthHandler
	api.TokenHeaderAuth = handlers.TokenAuthHandler

	// create an app
	api.AppsCreateAppHandler = apps.CreateAppHandlerFunc(func(params apps.CreateAppParams, principal interface{}) middleware.Responder {
		return handlers.CreateAppHandler(params, principal)
	})

	// create deployment
	api.DeploymentsCreateDeploymentHandler = deployments.CreateDeploymentHandlerFunc(func(params deployments.CreateDeploymentParams, principal interface{}) middleware.Responder {
		return handlers.CreateDeploymentHandler(params, principal)
	})
	// create a team
	api.TeamsCreateTeamHandler = teams.CreateTeamHandlerFunc(func(params teams.CreateTeamParams, principal interface{}) middleware.Responder {
		return handlers.CreateTeamHandler(params, principal)
	})
	// create a user
	api.UsersCreateUserHandler = users.CreateUserHandlerFunc(func(params users.CreateUserParams, principal interface{}) middleware.Responder {
		return handlers.CreateUserHandler(params, principal)
	})
	// delete team
	api.TeamsDeleteTeamHandler = teams.DeleteTeamHandlerFunc(func(params teams.DeleteTeamParams, principal interface{}) middleware.Responder {
		return handlers.DeleteTeamHandler(params, principal)
	})
	api.AppsGetAppDetailsHandler = apps.GetAppDetailsHandlerFunc(func(params apps.GetAppDetailsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation apps.GetAppDetails has not yet been implemented")
	})
	api.AppsGetAppsHandler = apps.GetAppsHandlerFunc(func(params apps.GetAppsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation apps.GetApps has not yet been implemented")
	})
	api.UsersGetCurrentUserHandler = users.GetCurrentUserHandlerFunc(func(principal interface{}) middleware.Responder {
		return handlers.GetCurrentUserHandler(principal)
	})
	api.DeploymentsGetDeploymentsHandler = deployments.GetDeploymentsHandlerFunc(func(params deployments.GetDeploymentsParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation deployments.GetDeployments has not yet been implemented")
	})
	// get team details
	api.TeamsGetTeamDetailHandler = teams.GetTeamDetailHandlerFunc(func(params teams.GetTeamDetailParams, principal interface{}) middleware.Responder {
		return handlers.GetTeamDetailsHandler(params, principal)
	})
	api.TeamsGetTeamsHandler = teams.GetTeamsHandlerFunc(func(params teams.GetTeamsParams, principal interface{}) middleware.Responder {
		return handlers.GetTeamsHandler(params, principal)
	})
	// get a single user from db
	api.UsersGetUserDetailsHandler = users.GetUserDetailsHandlerFunc(func(params users.GetUserDetailsParams, principal interface{}) middleware.Responder {
		return handlers.GetUserDetailsHandler(params, principal)
	})
	api.UsersGetUsersHandler = users.GetUsersHandlerFunc(func(params users.GetUsersParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation users.GetUsers has not yet been implemented")
	})
	api.AppsUpdateAppHandler = apps.UpdateAppHandlerFunc(func(params apps.UpdateAppParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation apps.UpdateApp has not yet been implemented")
	})
	// update a team
	api.TeamsUpdateTeamHandler = teams.UpdateTeamHandlerFunc(func(params teams.UpdateTeamParams, principal interface{}) middleware.Responder {
		return handlers.UpdateTeamHandler(params, principal)
	})
	api.UsersUpdateUserHandler = users.UpdateUserHandlerFunc(func(params users.UpdateUserParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation users.UpdateUser has not yet been implemented")
	})
	// login handler
	api.AuthUserLoginHandler = auth.UserLoginHandlerFunc(func(params auth.UserLoginParams) middleware.Responder {
		return handlers.LoginHandler(params)
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
