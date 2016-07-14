package handlers

import (
	"log"

	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
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

// GetAppDetailsHandler foo bar
func GetAppDetailsHandler(params apps.GetAppDetailsParams, principal interface{}) middleware.Responder {
	// FIXME: check if the token have permission on this team and app; maybe it's a good idea to centralize this check
	sa := storage.Application{}
	sa.ID = uint(params.AppID)
	if storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Addresses").Preload("EnvVars").Preload("Deployments").First(&sa).RecordNotFound() {
		log.Println("app info not found")
		return apps.NewGetAppDetailsForbidden()
	}

	scale := int64(sa.Scale)
	m := models.App{
		ID:    int64(sa.ID),
		Name:  &sa.Name,
		Scale: &scale,
	}
	m.AddressList = make([]string, len(sa.Addresses))
	for i, x := range sa.Addresses {
		m.AddressList[i] = x.Address
	}
	// TODO: add envvars ?!?
	// m.EnvVars
	m.DeploymentList = make([]*models.Deployment, len(sa.Deployments))
	for i, x := range sa.Deployments {
		w, _ := strfmt.ParseDateTime(x.CreatedAt.String())
		d := models.Deployment{
			UUID: &x.UUID,
			When: w,
			// Description: &x.Description
		}
		m.DeploymentList[i] = &d
	}
	r := apps.NewGetAppDetailsOK()
	r.SetPayload(&m)
	return r
}
