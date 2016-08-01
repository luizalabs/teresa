package handlers

import (
	"log"

	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/apps"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	storageRegion     = "us-east-1"
	storageBucketName = "teresa-staging"
	storageAccessKey  = "AKIAIUARH63XWZUMCFWA"
	storageSecretKey  = "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"
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
	// save to DB
	if err := storage.DB.Create(&sa).Error; err != nil {
		log.Printf("CreateAppHandler failed: %s\n", err)
		return apps.NewCreateAppDefault(500)
	}

	// get app and team info
	if storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Team").First(&sa).RecordNotFound() {
		log.Println("app info not found")
		return apps.NewCreateAppDefault(500)
	}

	// namespaces yaml
	nsy := api.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: getNamespaceName(sa.Team.Name, sa.Name),
		},
	}
	// creating namespace
	ns, err := k8sClient.Namespaces().Create(&nsy)
	if err != nil {
		log.Printf("Error when create the namespace [%s] for the app. Err: %s\n", nsy.GetName(), err)
		return apps.NewCreateAppDefault(500)
	}

	// secret yaml
	svcy := api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Type: api.SecretTypeOpaque,
		ObjectMeta: api.ObjectMeta{
			Name:      "s3-storage",
			Namespace: ns.GetName(),
		},
		Data: map[string][]byte{
			"region":         []byte(storageRegion),
			"builder-bucket": []byte(storageBucketName),
			"accesskey":      []byte(storageAccessKey),
			"secretkey":      []byte(storageSecretKey),
		},
	}
	// creating secret
	_, err = k8sClient.Secrets(ns.GetName()).Create(&svcy)
	if err != nil {
		log.Printf("Error creating the storage secret for the namespace [%s] . Err: %s\n", nsy.GetName(), err)
		return apps.NewCreateAppDefault(500)
	}

	a.ID = int64(sa.ID)
	r := apps.NewCreateAppCreated()
	r.SetPayload(&a)
	return r
}

// parseAppFromStorageToResponse receives a storage object and return an response object
func parseAppFromStorageToResponse(sa *storage.Application) (app *models.App) {
	scale := int64(sa.Scale)
	app = &models.App{
		ID:    int64(sa.ID),
		Name:  &sa.Name,
		Scale: &scale,
	}
	app.AddressList = make([]string, len(sa.Addresses))
	for i, x := range sa.Addresses {
		app.AddressList[i] = x.Address
	}

	app.EnvVars = make([]*models.EnvVar, len(sa.EnvVars))
	for i, x := range sa.EnvVars {
		k := x.Key
		v := x.Value
		e := models.EnvVar{
			Key:   &k,
			Value: &v,
		}
		app.EnvVars[i] = &e
	}

	app.DeploymentList = make([]*models.Deployment, len(sa.Deployments))
	for i, x := range sa.Deployments {
		id := x.UUID
		w, _ := strfmt.ParseDateTime(x.CreatedAt.String())
		d := models.Deployment{
			UUID: &id,
			When: w,
		}
		if des := x.Description; des != "" {
			d.Description = &des
		}
		app.DeploymentList[i] = &d
	}
	return
}

// GetAppDetailsHandler return app details
func GetAppDetailsHandler(params apps.GetAppDetailsParams, principal interface{}) middleware.Responder {
	// FIXME: check if the token have permission on this team and app; maybe it's a good idea to centralize this check
	sa := storage.Application{}
	sa.ID = uint(params.AppID)
	if storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Addresses").Preload("EnvVars").Preload("Deployments").First(&sa).RecordNotFound() {
		log.Println("app info not found")
		return apps.NewGetAppDetailsForbidden()
	}

	a := parseAppFromStorageToResponse(&sa)
	r := apps.NewGetAppDetailsOK()
	r.SetPayload(a)
	return r
}

// GetAppsHandler return apps for a team
func GetAppsHandler(params apps.GetAppsParams, principal interface{}) middleware.Responder {
	tk := principal.(*Token)

	// get user teams to check before continue
	rows, err := storage.DB.Table("teams_users").Where("user_id = ?", tk.UserID).Select("team_id as id").Rows()
	if err != nil {
		log.Printf("ERROR querying user teams: %s", err)
		return apps.NewGetAppsDefault(500)
	}
	defer rows.Close()
	userTeams := []int{}
	for rows.Next() {
		var teamID int
		rows.Scan(&teamID)
		userTeams = append(userTeams, teamID)
	}
	// check if user can se this team
	tf := false
	for _, x := range userTeams {
		if x == int(params.TeamID) {
			tf = true
			break
		}
	}
	if tf == false {
		log.Printf("ERROR user can see info about this team. Teams allowed: [%v]. Team provided: [%d]", userTeams, params.TeamID)
		return apps.NewGetAppsUnauthorized()
	}

	// TODO: admin user can see all teams... change here

	// FIXME: we can use this solution bellow to get more than one team from DB
	// if storage.DB.Where("team_id in (?)", userTeams).Preload("Addresses").Preload("EnvVars").Find(&storageAppList).RecordNotFound() {

	storageAppList := []*storage.Application{}
	if err = storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Addresses").Preload("EnvVars").Find(&storageAppList).Error; err != nil {
		log.Printf("ERROR when trying to recover apps from db: %s", err)
		return apps.NewGetAppsDefault(500)
	}
	if len(storageAppList) == 0 {
		log.Printf("No apps found for this team: %d", params.TeamID)
		return apps.NewGetAppsDefault(404)
	}

	appsList := []*models.App{}
	for _, sa := range storageAppList {
		a := parseAppFromStorageToResponse(sa)
		appsList = append(appsList, a)
	}

	r := apps.NewGetAppsOK()

	rb := apps.GetAppsOKBodyBody{}
	rb.Items = appsList
	r.SetPayload(rb)

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
