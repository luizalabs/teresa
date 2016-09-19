package k8s

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/luizalabs/tapi/helpers"
	"github.com/luizalabs/tapi/models"
	"k8s.io/kubernetes/pkg/api"
	k8s_errors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// AppsInterface is used to allow mock testing
type AppsInterface interface {
	Apps() AppInterface
}

// AppInterface is used to interact with Kubernetes and also to allow mock testing
type AppInterface interface {
	Create(app *models.App, storage helpers.Storage, tk *Token, l *log.Entry) error
	Update(app *models.App, tk *Token, l *log.Entry) error
	UpdateEnvVars(appName string, operations []*models.PatchAppRequest, tk *Token, l *log.Entry) (app *models.App, err error)
	Get(appName string, tk *Token, l *log.Entry) (app *models.App, err error)
}

type apps struct {
	k *k8sHelper
}

func newApps(c *k8sHelper) *apps {
	return &apps{k: c}
}

// Create creates an App inside kubernetes
// Inside kubernetes, the App is represented as an namespace
func (c apps) Create(app *models.App, storage helpers.Storage, tk *Token, l *log.Entry) error {
	l.Debug("creating app")
	// validating input params
	if err := validateBeforeCreate(app, tk, l); err != nil {
		return err
	}
	// check if user can create apps for the team
	if tk.IsAuthorized(*app.Team) == false {
		msg := "token not allowed to create apps for this team"
		l.Info(msg)
		return NewUnauthorizedError(msg)
	}
	app.Creator = &models.User{
		Email: tk.Email,
	}
	// creating namespace
	if err := c.createNamespace(app, *tk.Email, l); err != nil {
		return err
	}
	// creating quota (limit ranges) for namespace
	l.Debug("creating quota (limit range) for namespace")
	if err := c.createQuota(app, l); err != nil {
		return err
	}
	l.Debug("namespace quota created with success for namespace")
	// creating storage secret. this will be used to store the built App
	l.Debug("creating storage secret for namespace")
	if err := c.createStorageSecret(*app.Name, storage, l); err != nil {
		return err
	}
	l.Debug("secret created with success for namespace")
	l.Info("app created with success")

	return nil
}

func (c apps) Update(app *models.App, tk *Token, l *log.Entry) error {
	l.Debug("updating app")

	// TODO: update the deployment here if exists...
	// TODO: update namespace quota here if exists...

	if err := c.updateNamespace(app, *tk.Email, l); err != nil {
		return err
	}
	l.Info("app updated with success")
	return nil
}

func (c apps) UpdateEnvVars(appName string, operations []*models.PatchAppRequest, tk *Token, l *log.Entry) (app *models.App, err error) {
	l.Debug("updating env vars for App")
	// getting app
	app, err = c.Get(appName, tk, l)
	if err != nil {
		if IsUnauthorizedError(err) {
			l.WithError(err).Warn("token is not allowed to update env vars")
		}
		return nil, err
	}

	// applying the operations to App
	if err = updateAppEnvVars(app, operations, l); err != nil {
		return nil, err
	}
	if tk.IsAuthorized(*app.Team) == false {
		log.Printf(`token "%s" is not allowed to update env vars for the App "%s/%s"`, *tk.Email, *app.Team, *app.Name)
		return nil, NewUnauthorizedErrorf(`token not allowed to make changes for the App "%s"`, *app.Name)
	}

	if err := c.Update(app, tk, l); err != nil {
		return nil, err
	}
	l.Info("app env vars updated successfully")
	return
}

// Get returns an App by the name
func (c apps) Get(appName string, tk *Token, l *log.Entry) (app *models.App, err error) {
	l.Debug("trying to get app")
	ns, err := c.getNamespace(appName, l)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, NewUnauthorizedErrorf(`app "%s" not found or user not allowed to see it`, appName)
		}
		return nil, err
	}
	app, err = unmarshalAppFromNamespace(ns, l)
	if err != nil {
		return nil, err
	}
	// check if the user is authorized to get this App
	if tk.IsAuthorized(*app.Team) == false {
		return nil, NewUnauthorizedErrorf(`app "%s" not found or user not allowed to see it`, appName)
	}
	return
}

// createQuota creates an k8s Limit Range for the App (namespace)
func (c apps) createQuota(app *models.App, l *log.Entry) error {
	l.Debug("creating namespace quota")
	quota, err := newQuotaYaml(app, l)
	if err != nil {
		return err
	}
	if _, err = c.k.k8sClient.LimitRanges(*app.Name).Create(quota); err != nil {
		l.WithError(err).Error(`error when creating "quotas" for the namespace`)
		return err
	}
	return nil
}

// createNamespace creates an k8s namespace
// Inside kubernetes, every app is a k8s namespaces (1:1) with the App information inside
func (c apps) createNamespace(app *models.App, userEmail string, l *log.Entry) error {
	l.Debug("creating namespace")
	nsy := newAppNamespaceYaml(app, userEmail, l)
	if err := addAppToNamespaceYaml(app, nsy, l); err != nil {
		return err
	}
	if _, err := c.k.k8sClient.Namespaces().Create(nsy); err != nil {
		if k8s_errors.IsAlreadyExists(err) {
			msg := fmt.Sprintf(`a namespace with the name "%s" already exists`, *app.Name)
			l.Info(msg)
			return NewAlreadyExistsError(msg)
		}
		l.WithError(err).Error("error found when creating the namespace")
		return err
	}
	l.Debug("namespace created with success")
	return nil
}

// updateNamespace updates App information inside the namespace
func (c apps) updateNamespace(app *models.App, userEmail string, l *log.Entry) error {
	l.Debug("updating namespace")
	ns, err := c.getNamespace(*app.Name, l)
	if err != nil {
		return err
	}
	ai, err := json.Marshal(app)
	if err != nil {
		l.WithError(err).Error("error when marshalling the app")
		return err
	}
	ns.Annotations["teresa.io/app"] = string(ai)
	ns.Annotations["teresa.io/last-user"] = userEmail
	if _, err := c.k.k8sClient.Namespaces().Update(ns); err != nil {
		l.WithError(err).Error("error when updating the namespace")
		return err
	}
	return nil
}

// createStorageSecret creates a K8s Secret that will be used by the Builder and Runner processes
func (c apps) createStorageSecret(appName string, storage helpers.Storage, l *log.Entry) error {
	l.Debug("creating storage secret")
	svc := &api.Secret{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Type: api.SecretTypeOpaque,
		ObjectMeta: api.ObjectMeta{
			Name:      storage.GetK8sSecretName(),
			Namespace: appName,
		},
		Data: storage.GetAccessData(),
	}
	_, err := c.k.k8sClient.Secrets(appName).Create(svc)
	if err != nil {
		l.WithError(err).Error("error found when creating the storage secret for the namespace")
		return err
	}
	l.Debug("storage secret created with success")
	return nil
}

// getNamespace returns a namespace or an error if any
func (c apps) getNamespace(name string, l *log.Entry) (ns *api.Namespace, err error) {
	l.Debug("getting namespace from k8s")
	ns, err = c.k.k8sClient.Namespaces().Get(name)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			msg := "namespace not found"
			l.Debug(msg)
			return nil, NewNotFoundError(msg)
		}
		l.Error("error when trying to get namespace from k8s")
		return nil, err
	}
	return
}

// validateBeforeCreate validade all App parameters and return an InputError if any
func validateBeforeCreate(app *models.App, tk *Token, l *log.Entry) error {
	// TODO: validate...
	// - name
	// - team
	// - userEmail
	if app.Limits == nil {
		l.Debug("quota not specified for the app")
		return NewInputError("limits where not provided")
	}
	return nil
}

// newAppNamespaceYaml is a helper to create k8s namespace (parameters) for an App
func newAppNamespaceYaml(app *models.App, userEmail string, l *log.Entry) (ns *api.Namespace) {
	l.Debug("creating namespace params (yaml)")
	ns = &api.Namespace{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: *app.Name,
			Labels: map[string]string{
				"teresa.io/team": *app.Team,
			},
			Annotations: map[string]string{
				"teresa.io/last-user": userEmail,
			},
		},
	}
	l.Debug("namespace params created with success")
	return
}

// addAppToNamespaceYaml marshall the entire App struct and put inside the namesapce Yaml
func addAppToNamespaceYaml(app *models.App, ns *api.Namespace, l *log.Entry) error {
	l.Debug("marshalling app to put inside namespace")
	ai, err := json.Marshal(app)
	if err != nil {
		l.WithError(err).Error("error found when marshalling the app to put inside the namespace annotation.")
		return err
	}
	ns.Annotations["teresa.io/app"] = string(ai)
	l.Debug("app string inserted to namespace")
	return nil
}

// addLimitRangeQuantityToResourceList is a helper to the function parseLimitRangeParams, used to add a limit range
// to a specific limit range list
func addLimitRangeQuantityToResourceList(r *api.ResourceList, limitRangeQuantity []*models.LimitRangeQuantity, l *log.Entry) error {
	if limitRangeQuantity == nil {
		return nil
	}
	rl := api.ResourceList{}
	for _, item := range limitRangeQuantity {
		name := api.ResourceName(*item.Resource)
		q, err := resource.ParseQuantity(*item.Quantity)
		if err != nil {
			l.WithError(err).Errorf(`error when trying to parse limits value "%s:%s".`, *item.Resource, *item.Quantity)
			return err
		}
		rl[name] = q
	}
	*r = rl
	return nil
}

// parseLimitRangeParams is a helper to parse all the limit range types
func parseLimitRangeParams(limitRangeItem *api.LimitRangeItem, limits *models.AppInLimits, l *log.Entry) error {
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.Default, limits.Default, l); err != nil {
		return err
	}
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.DefaultRequest, limits.DefaultRequest, l); err != nil {
		return err
	}
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.Max, limits.Max, l); err != nil {
		return err
	}
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.Min, limits.Min, l); err != nil {
		return err
	}
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.MaxLimitRequestRatio, limits.LimitRequestRatio, l); err != nil {
		return err
	}
	return nil
}

// newQuotaYaml is a helper to create an k8s LimitRange based on App Limits
func newQuotaYaml(app *models.App, l *log.Entry) (lr *api.LimitRange, err error) {
	l.Debug("create quota for the namespace")
	lrItem := api.LimitRangeItem{
		Type: api.LimitTypeContainer,
	}
	// parse limits params to k8s params
	l.Debug("parse limit ranges...")
	if err = parseLimitRangeParams(&lrItem, app.Limits, l); err != nil {
		msg := `found error when parsing app "limits"`
		l.Error(msg)
		return nil, NewInputError(msg)
	}
	lr = &api.LimitRange{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "LimitRange",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: "limits",
		},
		Spec: api.LimitRangeSpec{
			Limits: []api.LimitRangeItem{lrItem},
		},
	}
	l.Debug("quota created with success for the namespace")
	return
}

// unmarshalAppFromNamespace extract and unmarshal the App from the namespace
func unmarshalAppFromNamespace(ns *api.Namespace, l *log.Entry) (app *models.App, err error) {
	s, ok := ns.GetAnnotations()["teresa.io/app"]
	if ok == false {
		msg := fmt.Sprintf(`annotation "teresa.io/app" not found on this provided namespace`)
		l.WithField("namespace", ns.Name).Error(msg)
		return nil, errors.New(msg)
	}
	app = &models.App{}
	err = json.Unmarshal([]byte(s), app)
	if err != nil {
		l.WithError(err).WithField("namespace", ns.Name).Error("error when trying to unmarshal the app from namespace.")
		return nil, err
	}
	return
}

// checkForProtectedEnvVars check if is there any Operation trying to modify some of the protected env vars
func checkForProtectedEnvVars(operations []*models.PatchAppRequest, l *log.Entry) error {
	protectedEnvVars := [...]string{"SLUG_URL", "PORT", "DEIS_DEBUG", "BUILDER_STORAGE"}
	for _, operation := range operations {
		for _, operationValue := range operation.Value {
			for _, pv := range protectedEnvVars {
				if *operationValue.Key == pv {
					msg := fmt.Sprintf(`manually changing the env var "%s" isn't allowed`, pv)
					l.Warn(msg)
					return NewInputError(msg)
				}
			}
		}
	}
	return nil
}

// updateAppEnvVars updates App env vars based on the operations
func updateAppEnvVars(app *models.App, operations []*models.PatchAppRequest, l *log.Entry) error {
	l.Debug("updating env vars for the app")
	if err := checkForProtectedEnvVars(operations, l); err != nil {
		return err
	}
	// applying the operations to App
	for _, operation := range operations {
		for _, operationValue := range operation.Value {
			if *operation.Op == "add" { // create or update env var
				evarFound := false
				// App env vars
				for _, evar := range app.EnvVars {
					if *evar.Key == *operationValue.Key {
						evar.Value = &operationValue.Value
						evarFound = true
						break
					}
				}
				// env var not found on App, register this new one
				if evarFound == false {
					app.EnvVars = append(app.EnvVars, &models.EnvVar{
						Key:   operationValue.Key,
						Value: &operationValue.Value,
					})
				}
			} else if *operation.Op == "remove" { // remove env var
				for i, evar := range app.EnvVars {
					if *evar.Key == *operationValue.Key {
						app.EnvVars = append(app.EnvVars[:i], app.EnvVars[i+1:]...)
						break
					}
				}
			}
		}
	}
	l.Debug("env vars updated with success for the app")
	return nil
}
