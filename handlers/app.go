package handlers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/luizalabs/teresa-api/helpers"
	"github.com/luizalabs/teresa-api/k8s"
	"github.com/luizalabs/teresa-api/models"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/restapi/operations/apps"
)

// CreateAppHandler handler for "-X POST /apps"
var CreateAppHandler apps.CreateAppHandlerFunc = func(params apps.CreateAppParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	// App informations
	app := &models.App{AppIn: *params.Body}

	l := log.WithField("app", *app.Name).WithField("team", *app.Team).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())

	if err := k8s.Client.Apps().Create(app, helpers.FileStorage, tk); err != nil {
		if k8s.IsInputError(err) {
			l.WithError(err).Warn("error when creating app")
			return NewBadRequestError(err)
		} else if k8s.IsAlreadyExistsError(err) {
			l.WithError(err).Debug("error when creating app")
			return NewConflictError("app already exists")
		}
		l.WithError(err).Error("error when creating app")
		return NewInternalServerError(err)
	}
	return apps.NewCreateAppCreated().WithPayload(app)
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
var GetAppDetailsHandler apps.GetAppDetailsHandlerFunc = func(params apps.GetAppDetailsParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)

	l := log.WithField("app", params.AppName).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())

	// FIXME: implements a functions GetFull, that will return the App + LB address + Deployments

	app, err := k8s.Client.Apps().Get(params.AppName, tk)
	if err != nil {
		if k8s.IsNotFoundError(err) {
			l.WithError(err).Debug("error when getting detail for the app")
			return NewNotFoundError(err)
		} else if k8s.IsUnauthorizedError(err) {
			l.WithError(err).Info("error when getting detail for the app")
			return NewUnauthorizedError(err)
		}
		l.WithError(err).Error("error when getting detail for the app")
		return NewInternalServerError(err)
	}
	return apps.NewGetAppDetailsOK().WithPayload(app)
}

// GetAppsHandler return apps for a team
func GetAppsHandler(params apps.GetAppsParams, principal interface{}) middleware.Responder {
	// tk := k8s.IToToken(principal)
	//
	// // get user teams to check before continue
	// rows, err := storage.DB.Table("teams_users").Where("user_id = ?", tk.UserID).Select("team_id as id").Rows()
	// if err != nil {
	// 	log.Printf("ERROR querying user teams: %s", err)
	// 	return apps.NewGetAppsDefault(500)
	// }
	// defer rows.Close()
	// userTeams := []int{}
	// for rows.Next() {
	// 	var teamID int
	// 	rows.Scan(&teamID)
	// 	userTeams = append(userTeams, teamID)
	// }
	// // check if user can se this team
	// tf := false
	// for _, x := range userTeams {
	// 	if x == int(params.TeamID) {
	// 		tf = true
	// 		break
	// 	}
	// }
	// if tf == false {
	// 	log.Printf("ERROR user can see info about this team. Teams allowed: [%v]. Team provided: [%d]", userTeams, params.TeamID)
	// 	return apps.NewGetAppsUnauthorized()
	// }
	//
	// // TODO: admin user can see all teams... change here
	//
	// // FIXME: we can use this solution bellow to get more than one team from DB
	// // if storage.DB.Where("team_id in (?)", userTeams).Preload("Addresses").Preload("EnvVars").Find(&storageAppList).RecordNotFound() {
	//
	// storageAppList := []*storage.Application{}
	// if err = storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Addresses").Preload("EnvVars").Find(&storageAppList).Error; err != nil {
	// 	log.Printf("ERROR when trying to recover apps from db: %s", err)
	// 	return apps.NewGetAppsDefault(500)
	// }
	// if len(storageAppList) == 0 {
	// 	log.Printf("No apps found for this team: %d", params.TeamID)
	// 	return apps.NewGetAppsDefault(404)
	// }
	//
	// appsList := []*models.App{}
	// for _, sa := range storageAppList {
	// 	a := parseAppFromStorageToResponse(sa)
	// 	appsList = append(appsList, a)
	// }
	//
	// r := apps.NewGetAppsOK()
	//
	// rb := apps.GetAppsOKBodyBody{}
	// rb.Items = appsList
	// r.SetPayload(rb)
	//
	// return r
	return apps.NewGetAppsOK()
}

// PartialUpdateAppHandler partial updating app... only envvars for now
var PartialUpdateAppHandler apps.PartialUpdateAppHandlerFunc = func(params apps.PartialUpdateAppParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	l := log.WithField("app", params.AppName).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())

	app, err := k8s.Client.Apps().UpdateEnvVars(params.AppName, params.Body, tk)
	if err != nil {
		if k8s.IsInputError(err) {
			l.WithError(err).Warn("error during partial update for the app")
			return NewBadRequestError(err)
		} else if k8s.IsNotFoundError(err) {
			l.WithError(err).Debug("error during partial update for the app")
			return NewNotFoundError(err)
		} else if k8s.IsUnauthorizedError(err) {
			l.WithError(err).Warn("error during partial update for the app")
			return NewUnauthorizedError(err)
		}
		l.WithError(err).Error("error during partial update for the app")
		return NewInternalServerError(err)
	}
	return apps.NewPartialUpdateAppOK().WithPayload(app)
}
