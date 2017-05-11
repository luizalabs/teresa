package restapi

import (
	"crypto/tls"
	"net/http"

	log "github.com/Sirupsen/logrus"
	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/teresa-api/handlers"
	"github.com/luizalabs/teresa-api/restapi/operations"
	"github.com/luizalabs/teresa-api/restapi/operations/deployments"
	"github.com/luizalabs/teresa-api/restapi/operations/teams"
	"github.com/luizalabs/teresa-api/restapi/operations/users"
	"github.com/x-cray/logrus-prefixed-formatter"
)

// This file is safe to edit. Once it exists it will not be overwritten

func configureFlags(api *operations.TeresaAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(new(prefixed.TextFormatter))
}

func configureAPI(api *operations.TeresaAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	api.Logger = log.Infof

	api.MultipartformConsumer = runtime.DiscardConsumer

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.BinProducer = runtime.ByteStreamProducer()

	// authentication
	api.APIKeyAuth = handlers.TokenAuthHandler
	api.TokenHeaderAuth = handlers.TokenAuthHandler

	// create an app
	api.AppsCreateAppHandler = handlers.CreateAppHandler

	// create deployment
	api.DeploymentsCreateDeploymentHandler = handlers.CreateDeploymentHandler

	// create team
	api.TeamsCreateTeamHandler = teams.CreateTeamHandlerFunc(func(params teams.CreateTeamParams, principal interface{}) middleware.Responder {
		return handlers.CreateTeamHandler(params, principal)
	})
	api.TeamsAddUserToTeamHandler = teams.AddUserToTeamHandlerFunc(func(params teams.AddUserToTeamParams, principal interface{}) middleware.Responder {
		return handlers.AddUserToTeam(params, principal)
	})
	// create user
	api.UsersCreateUserHandler = users.CreateUserHandlerFunc(func(params users.CreateUserParams, principal interface{}) middleware.Responder {
		return handlers.CreateUserHandler(params, principal)
	})
	// delete user
	api.UsersDeleteUserHandler = users.DeleteUserHandlerFunc(func(params users.DeleteUserParams, principal interface{}) middleware.Responder {
		return handlers.DeleteUserHandler(params, principal)
	})
	// delete team
	api.TeamsDeleteTeamHandler = teams.DeleteTeamHandlerFunc(func(params teams.DeleteTeamParams, principal interface{}) middleware.Responder {
		return handlers.DeleteTeamHandler(params, principal)
	})
	// app details
	api.AppsGetAppDetailsHandler = handlers.GetAppDetailsHandler

	// app logs
	api.AppsGetAppLogsHandler = handlers.GetAppLogsHandler

	// list apps
	api.AppsGetAppsHandler = handlers.GetAppsHandler

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
	// partial update app... update envVars
	api.AppsPartialUpdateAppHandler = handlers.PartialUpdateAppHandler

	// update app
	api.AppsUpdateAppHandler = handlers.UpdateAppHandler

	// update app autoscale info
	api.AppsUpdateAppAutoScaleHandler = handlers.UpdateAppAutoScaleHandler

	api.AppsUpdateAppScaleHandler = handlers.UpdateAppScaleHandler

	// update a team
	api.TeamsUpdateTeamHandler = teams.UpdateTeamHandlerFunc(func(params teams.UpdateTeamParams, principal interface{}) middleware.Responder {
		return handlers.UpdateTeamHandler(params, principal)
	})
	api.UsersUpdateUserHandler = users.UpdateUserHandlerFunc(func(params users.UpdateUserParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation users.UpdateUser has not yet been implemented")
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
