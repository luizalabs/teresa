package handlers

import (
	"log"

	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/apps"
	"k8s.io/kubernetes/pkg/api"
)

// CreateAppHandler create apps
func CreateAppHandler(params apps.CreateAppParams, principal interface{}) middleware.Responder {
	a := models.App{
		Name:  params.Body.Name,
		Scale: params.Body.Scale,
	}
	sa := storage.Application{
		Name:   *params.Body.Name,
		Scale:  int16(*params.Body.Scale),
		TeamID: uint(params.TeamID),
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

// PartialUpdateAppHandler partial updating app... only envvars for now
func PartialUpdateAppHandler(params apps.PartialUpdateAppParams, principal interface{}) middleware.Responder {
	log.Printf("executing partial update for app [%d] envvars\n", params.AppID)
	// TODO: find a better place to put this centralized
	slugRunnerEnvVars := []string{"SLUG_URL", "PORT", "DEIS_DEBUG", "BUILDER_STORAGE"}

	// get info about the app
	appID := uint(params.AppID)
	sApp := storage.Application{}
	sApp.ID = appID
	if storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Team").Preload("EnvVars").First(&sApp).RecordNotFound() {
		log.Printf("app [%d] not found in db.\n", appID)
		return apps.NewPartialUpdateAppDefault(500)
	}

	sEnvVars := &sApp.EnvVars

	// start transaction
	t := storage.DB.Begin()
	var te error
	// checking operations
	for _, op := range params.Body {
		createUpdateEnvVars := func() error {
			for _, opv := range op.Value {
				kf := false // key found controll
				for _, e := range *sEnvVars {
					if *opv.Key != e.Key {
						continue
					}
					// envvar already exists... update
					kf = true
					if err := t.Model(&e).UpdateColumns(storage.EnvVar{Value: opv.Value}).Error; err != nil {
						return err
					}
					break
				}
				if kf == false {
					newEnv := storage.EnvVar{ // envvar not found... create
						AppID: appID,
						Key:   *opv.Key,
						Value: opv.Value,
					}
					if err := t.Create(&newEnv).Error; err != nil {
						return err
					}
				}
			}
			return nil
		}
		deleteEnvVars := func() error {
			for _, o := range op.Value {
				for _, e := range *sEnvVars {
					if *o.Key != e.Key {
						continue
					}
					if err := t.Delete(&e).Error; err != nil {
						return err
					}
					break
				}
			}
			return nil
		}
		// checking for invalid keys to be changed manually
		for i, opv := range op.Value {
			for _, k := range slugRunnerEnvVars {
				if *opv.Key != k {
					continue
				}
				log.Printf(`invalid key "%s" to be changed manually... discarding this before do something\n`, k)
				op.Value = append(op.Value[:i], op.Value[i+1:]...)
				break
			}
		}
		if *op.Op == "add" {
			te = createUpdateEnvVars()
		} else if *op.Op == "remove" {
			te = deleteEnvVars()
		}
		if te != nil {
			break
		}
	}
	// check for errors
	if te != nil {
		t.Rollback()
		log.Printf("error doing a partial update to app envvars. %s\n", te)
		return apps.NewPartialUpdateAppDefault(500)
	}
	// commit transaction
	if err := t.Commit().Error; err != nil {
		log.Printf("error doing a partial update to app envvars. %s\n", err)
		return apps.NewPartialUpdateAppDefault(500)
	}

	// check if the k8s deploy exists and update his envvars
	dp := newDeployParams(&sApp)

	// get k8s deploy
	d, err := getDeploy(dp)
	if err != nil {
		log.Printf("error when trying to collect info about the k8s_deploy. %s\n", err)
		return apps.NewPartialUpdateAppDefault(500)
	}
	if d != nil { // deploy exists
		// get and update envvars
		sEnvVars := []storage.EnvVar{}
		if r := storage.DB.Where(&storage.EnvVar{AppID: appID}).Find(&sEnvVars); r.Error != nil && r.RecordNotFound() == false {
			log.Printf("error getting env_vars for the app [%d] from db to update deployment. %s\n", params.AppID, r.Error)
			return apps.NewPartialUpdateAppDefault(500)
		}
		// extract slugrunner env vars from deployment
		newEnvVars := []api.EnvVar{}
		for _, se := range slugRunnerEnvVars {
			for _, e := range d.Spec.Template.Spec.Containers[0].Env {
				if e.Name == se {
					newEnvVars = append(newEnvVars, e)
					break
				}
			}
		}
		// insert app env vars
		for _, ne := range sEnvVars {
			e := api.EnvVar{
				Name:  ne.Key,
				Value: ne.Value,
			}
			newEnvVars = append(newEnvVars, e)
		}
		d.Spec.Template.Spec.Containers[0].Env = newEnvVars
		// update k8s deploy
		if _, err := updateDeploy(d, "update on envvars"); err != nil {
			log.Printf("error when updating the k8s deploy [%s/%s]. %s\n", d.GetNamespace(), d.GetName(), err)
			return apps.NewPartialUpdateAppDefault(500)
		}
	}

	return apps.NewPartialUpdateAppOK()
}
