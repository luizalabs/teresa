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
		return handlers.CreateUserHandler(params, principal)
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
		return handlers.GetUserDetailsHandler(params, principal)
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
