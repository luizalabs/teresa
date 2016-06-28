package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/paas/api/restapi/operations/deployments"
)

func CreateDeploymentHandler(params deployments.CreateDeploymentParams, principal interface{}) middleware.Responder {
	return middleware.NotImplemented("operation deployments.CreateDeployment has not yet been implemented")
}
