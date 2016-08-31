package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/luizalabs/teresa-api/models"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/restapi/operations/apps"
	"k8s.io/kubernetes/pkg/api"
	k8s_errors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

func addQuantityToResourceList(r *api.ResourceList, quota []*models.LimitRangeQuantity) error {
	if quota != nil {
		rl := api.ResourceList{}
		for _, item := range quota {
			name := api.ResourceName(*item.Resource)
			q, err := resource.ParseQuantity(*item.Quantity)
			if err != nil {
				log.Printf(`error when trying to parse limits value "%s:%s". Err: %s`, *item.Resource, *item.Quantity, err)
				return err
			}
			rl[name] = q
		}
		*r = rl
	}
	return nil
}

func parseLimitsParams(limitRangeItem *api.LimitRangeItem, limits *models.AppInLimits) error {
	if err := addQuantityToResourceList(&limitRangeItem.Default, limits.Default); err != nil {
		return err
	}
	if err := addQuantityToResourceList(&limitRangeItem.DefaultRequest, limits.DefaultRequest); err != nil {
		return err
	}
	if err := addQuantityToResourceList(&limitRangeItem.Max, limits.Max); err != nil {
		return err
	}
	if err := addQuantityToResourceList(&limitRangeItem.Min, limits.Min); err != nil {
		return err
	}
	if err := addQuantityToResourceList(&limitRangeItem.MaxLimitRequestRatio, limits.LimitRequestRatio); err != nil {
		return err
	}
	return nil
}

// CreateAppHandler create apps
func CreateAppHandler(params apps.CreateAppParams, principal interface{}) middleware.Responder {
	tk := principal.(*Token)

	// FIXME: wee should mode this "team checking" to a middleware ASAP!!!
	var teamsFound int
	if params.Body.Team != "" {
		if err := storage.DB.Raw("select count(*) as count from teams where name = ?", params.Body.Team).Count(&teamsFound).Error; err != nil {
			log.Printf(`error when trying to check if the team "%s" is valid for the user "%s". Err: %s`, params.Body.Team, tk.Email, err)
			return NewInternalServerError()
		}
		if teamsFound == 0 {
			log.Printf(`team "%s" not found`, params.Body.Team)
			return NewUnauthorizedError("team not found or user dont have permission to do actions with the team provided")
		}
	} else {
		if tk.IsAdmin {
			log.Printf(`user "%s" is admin and not provided a team`, tk.Email)
			return NewBadRequestError("team is required when user is in more than one team")
		}
		q := "select * from teams inner join teams_users on teams.id = teams_users.team_id inner join users on teams_users.user_id = users.id where users.email = ?"
		if err := storage.DB.Raw(q, tk.Email).Count(&teamsFound).Error; err != nil {
			log.Printf(`user "%s" is in more than one team and provided none`, tk.Email)
			return NewBadRequestError("team is required when user is in more than one team")
		}
	}
	// creating namespace (aka App) params...
	nsParams := api.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: *params.Body.Name,
			Labels: map[string]string{
				"teresa.io/team": params.Body.Team,
			},
			Annotations: map[string]string{
				"teresa.io/last-user": tk.Email,
			},
		},
	}
	// marshalling the params appIn to store inside namespace annotations...
	ai, err := json.Marshal(params.Body)
	if err != nil {
		log.Printf(`error when trying to marshal the parameters for the namespace "%s". Err: %s`, *params.Body.Name, err)
		return NewInternalServerError()
	}
	nsParams.Annotations["teresa.io/app"] = string(ai)
	// checking for quota specifications...
	if params.Body.Limits == nil {
		log.Printf(`error when trying to create a namespace "%s". limits is not provide`, *params.Body.Name)
		return NewBadRequestError("limits is not provided")
	}
	// creating quota specifications...
	lrItem := api.LimitRangeItem{
		Type: api.LimitTypeContainer,
	}
	// parse limits params to k8s params
	if err := parseLimitsParams(&lrItem, params.Body.Limits); err != nil {
		log.Printf(`error when trying to parse "limits" for the app "%s"`, *params.Body.Name)
		return NewBadRequestError(fmt.Sprintf(`error found when parsing "limits". err.: %s`, err))
	}
	// creating namespace limits params
	limitsParams := api.LimitRange{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "LimitRange",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: "limits",
		},
		Spec: api.LimitRangeSpec{
			Limits: []api.LimitRangeItem{
				lrItem,
			},
		},
	}
	// create namespace
	if _, err := k8sClient.Namespaces().Create(&nsParams); err != nil {
		if k8s_errors.IsAlreadyExists(err) {
			log.Printf(`already exists a namespace with this name "%s". Err: %s`, *params.Body.Name, err)
			return NewConflictError("team already exists")
		}
		log.Printf(`error when trying to create the namespace "%s"`, *params.Body.Name)
		return NewInternalServerError()
	}
	log.Printf(`namespace "%s" created with success`, *params.Body.Name)
	// create quota (aka limit range)
	if _, err := k8sClient.LimitRanges(*params.Body.Name).Create(&limitsParams); err != nil {
		log.Printf(`error when trying to create the "limit range" for the namespace "%s". Err: %s `, *params.Body.Name, err)
		return NewInternalServerError()
	}
	log.Printf(`limit ranges created with success for the namespace "%s"`, *params.Body.Name)
	// creating k8s secrets for the namespace.
	// this will be used by the building and runner proccess to access the storage
	// FIXME: maybe we need only one of this secret to everybody
	svcParams := api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Type: api.SecretTypeOpaque,
		ObjectMeta: api.ObjectMeta{
			Name:      "s3-storage",
			Namespace: *params.Body.Name,
		},
		Data: map[string][]byte{
			"region":         []byte(builderConfig.AwsRegion),
			"builder-bucket": []byte(builderConfig.AwsBucket),
			"accesskey":      []byte(builderConfig.AwsKey),
			"secretkey":      []byte(builderConfig.AwsSecret),
		},
	}
	// creating the secret
	if _, err := k8sClient.Secrets(*params.Body.Name).Create(&svcParams); err != nil {
		log.Printf(`error when creating the storage secret for the namespace "%s" . Err: %s`, *params.Body.Name, err)
		return NewInternalServerError()
	}
	log.Printf(`secret created with success for the namespace "%s"`, *params.Body.Name)

	log.Printf(`namespace (aka App) "%s" created with success by the user "%s" for the team "%s"`, *params.Body.Name, tk.Email, params.Body.Team)

	app := models.App{AppIn: *params.Body}
	creator := models.User{
		Name: &tk.Email,
	}
	app.Creator = &creator
	res := apps.NewCreateAppCreated()
	res.SetPayload(&app)
	return res
}

// parseAppFromStorageToResponse receives a storage object and return an response object
func parseAppFromStorageToResponse(sa *storage.Application) (app *models.App) {
	scale := int64(sa.Scale)
	app = &models.App{}
	app.Name = &sa.Name
	app.Scale = scale

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
