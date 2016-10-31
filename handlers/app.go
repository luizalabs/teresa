package handlers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/go-openapi/runtime/middleware"
	"github.com/luizalabs/teresa-api/helpers"
	"github.com/luizalabs/teresa-api/k8s"
	"github.com/luizalabs/teresa-api/models"
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
		} else if k8s.IsUnauthorizedError(err) {
			l.WithError(err).Debug("error when creating app")
			return NewUnauthorizedError("team doesn't exists or user cant see this team")
		}
		l.WithError(err).Error("error when creating app")
		return NewInternalServerError(err)
	}
	return apps.NewCreateAppCreated().WithPayload(app)
}

// GetAppDetailsHandler return app details
var GetAppDetailsHandler apps.GetAppDetailsHandlerFunc = func(params apps.GetAppDetailsParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)

	l := log.WithField("app", params.AppName).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())

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

// UpdateAppHandler handler for update app
var UpdateAppHandler apps.UpdateAppHandlerFunc = func(params apps.UpdateAppParams, principal interface{}) middleware.Responder {
	// tk := k8s.IToToken(principal)

	// ############################################################
	// FIXME: stopped this because it's not very usefull right now
	// ############################################################

	// update the app body name to force secure update
	// params.Body.Name = &params.AppName
	// app := &models.App{AppIn: *params.Body}
	//
	// l := log.WithField("app", *app.Name).WithField("team", *app.Team).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())
	//
	// if err := k8s.Client.Apps().Update(app, helpers.FileStorage, tk); err != nil {
	// 	if k8s.IsInputError(err) {
	// 		l.WithError(err).Warn("error when updating app")
	// 		return NewBadRequestError(err)
	// 	}
	// 	l.WithError(err).Error("error when creating app")
	// 	return NewInternalServerError(err)
	// }
	// l.Debug("app updated with success")
	// return apps.NewUpdateAppOK().WithPayload(app)
	return middleware.NotImplemented("operation apps.UpdateAppHandlerFunc has not yet been implemented")
}

// GetAppsHandler returns all apps for the user
var GetAppsHandler apps.GetAppsHandlerFunc = func(params apps.GetAppsParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	l := log.WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())
	appList, err := k8s.Client.Apps().List(tk)
	if err != nil {
		if k8s.IsNotFoundError(err) {
			l.WithError(err).Debug("error when getting app list")
			return NewNotFoundError(err)
		}
		l.WithError(err).Error("error when getting app list")
		return NewInternalServerError(err)
	}
	return apps.NewGetAppsOK().WithPayload(appList)
}

// PartialUpdateAppHandler partial updating app... only envvars for now
var PartialUpdateAppHandler apps.PartialUpdateAppHandlerFunc = func(params apps.PartialUpdateAppParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	l := log.WithField("app", params.AppName).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())

	app, err := k8s.Client.Apps().UpdateEnvVars(params.AppName, params.Body, helpers.FileStorage, tk)
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

// UpdateAppScaleHandler handler for app scale -XPUT
var UpdateAppScaleHandler apps.UpdateAppScaleHandlerFunc = func(params apps.UpdateAppScaleParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	l := log.WithField("app", params.AppName).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())
	// updating the app scale
	app, err := k8s.Client.Apps().UpdateScale(params.AppName, *params.Body.Scale, helpers.FileStorage, tk)
	if err != nil {
		if k8s.IsInputError(err) {
			l.WithError(err).Warn("error when updating the app scale")
			return NewBadRequestError(err)
		} else if k8s.IsNotFoundError(err) {
			l.WithError(err).Debug("error when updating the app scale")
			return NewNotFoundError(err)
		} else if k8s.IsUnauthorizedError(err) {
			l.WithError(err).Warn("error when updating the app scale")
			return NewUnauthorizedError(err)
		}
		l.WithError(err).Error("error when updating the app scale")
		return NewInternalServerError(err)
	}
	l.Debugf(`app scale updated with success to: %d`, app.Scale)
	return apps.NewUpdateAppScaleOK().WithPayload(app)
}

// UpdateAppAutoScaleHandler update app autoscale info
var UpdateAppAutoScaleHandler apps.UpdateAppAutoScaleHandlerFunc = func(params apps.UpdateAppAutoScaleParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	l := log.WithField("app", params.AppName).WithField("token", *tk.Email).WithField("requestId", helpers.NewShortUUID())
	app, err := k8s.Client.Apps().UpdateAutoScale(params.AppName, params.Body, helpers.FileStorage, tk)
	if err != nil {
		if k8s.IsInputError(err) {
			l.WithError(err).Warn("error when updating the app auto scale")
			return NewBadRequestError(err)
		} else if k8s.IsNotFoundError(err) {
			l.WithError(err).Debug("error when updating the app auto scale")
			return NewNotFoundError(err)
		} else if k8s.IsUnauthorizedError(err) {
			l.WithError(err).Warn("error when updating the app auto scale")
			return NewUnauthorizedError(err)
		}
		l.WithError(err).Error("error when updating the app auto scale")
		return NewInternalServerError(err)
	}
	l.Debug(`app auto scale updated with success`)
	return apps.NewUpdateAppAutoScaleOK().WithPayload(app)
}
