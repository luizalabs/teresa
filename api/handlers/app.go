package handlers

import (
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/apps"
)

// CreateAppHandler create apps
func CreateAppHandler(params apps.CreateAppParams, principal interface{}) middleware.Responder {
	a := models.App{
		Name:  params.Body.Name,
		Scale: params.Body.Scale,
	}
	sa := storage.Application{
		Name:  *params.Body.Name,
		Scale: int16(*params.Body.Scale),
	}
	if err := storage.DB.Create(&sa).Error; err != nil {
		log.Printf("CreateAppHandler failed: %s\n", err)
		return apps.NewCreateAppDefault(500)
	}
	a.ID = int64(sa.ID)
	r := apps.NewCreateAppCreated()
	r.SetPayload(&a)
	return r
}
