package k8s

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/luizalabs/tapi/helpers"
	"github.com/luizalabs/tapi/models"
	"github.com/luizalabs/tapi/models/storage"
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
	Create(app *models.App, userEmail string, storage helpers.Storage) error
	UpdateEnvVars(appName, userEmail string, userIsAdmin bool, operations []*models.PatchAppRequest) (app *models.App, err error)
}

type apps struct {
	k *k8sHelper
}

func newApps(c *k8sHelper) *apps {
	return &apps{k: c}
}

// Create creates an App inside kubernetes
// Inside kubernetes, the App is represented as an namespace
func (c apps) Create(app *models.App, userEmail string, storage helpers.Storage) error {
	if err := validateBeforeCreate(app, userEmail); err != nil {
		log.Printf(`error found when validating the input params for the app "%s". Err: %s`, *app.Name, err)
		return err
	}
	app.Creator = &models.User{
		Email: &userEmail,
	}
	nsYaml := newAppNamespaceYaml(app, userEmail)
	if err := addAppToNamespaceYaml(app, nsYaml); err != nil {
		log.Printf(`error when trying to add app information to the namespace yaml. App name: %s; Err: %s`, *app.Name, err)
		return err
	}
	appQuota, err := newAppQuotaYaml(app)
	if err != nil {
		log.Printf(`error when trying to create the "quota yaml" for the app "%s", Err: %s`, *app.Name, err)
		return err
	}
	if err := c.createAppNamespace(nsYaml); err != nil {
		log.Printf(`error creating the namespace for the app "%s". Err: %s`, *app.Name, err)
		return err
	}
	log.Printf(`namespace created with success for the app "%s"`, *app.Name)
	if err := c.createAppQuota(*app.Name, appQuota); err != nil {
		log.Printf(`error when trying to create quotas for the app "%s". Err: %s `, *app.Name, err)
		return err
	}
	log.Printf(`"quota" created with success for the app "%s"`, *app.Name)
	if err := c.createAppStorageSecret(*app.Name, storage); err != nil {
		log.Printf(`error when trying to create a secrect for the app "%s". Err: %s`, *app.Name, err)
		return err
	}
	log.Printf(`secret created with success for the app "%s"`, *app.Name)
	log.Printf(`app (a.k.a. namespace in k8s) "%s" created with success by the user "%s" for the team "%s"`, *app.Name, userEmail, *app.Team)
	return nil
}

func (c apps) UpdateEnvVars(appName, userEmail string, userIsAdmin bool, operations []*models.PatchAppRequest) (app *models.App, err error) {
	ns, err := c.getNamespace(appName)
	if err != nil {
		log.Printf(`error found when trying to get the namespace "%s". Err: %s`, appName, err)
		return nil, err
	}
	// extracting and unmarshalling the App from the namespace
	if app, err = unmarshalAppFromNamespace(ns); err != nil {
		log.Printf(`error when unmarshalling the app (annotation "teresa.io/app") from the namespace "%s". Err: %s`, appName, err)
		return nil, err
	}

	// FIXME: remove this to somewhere... i don't know where ¯\_(ツ)_/¯
	var (
		dbQuery  string
		dbParams []interface{}
	)
	if userIsAdmin {
		dbQuery = "select * from teams where name = ?"
		dbParams = append(dbParams, []interface{}{*app.Team})
	} else {
		dbQuery = "select * from teams inner join teams_users on teams.id = teams_users.team_id inner join users on teams_users.user_id = users.id where users.email = ? and teams.name = ?"
		dbParams = append(dbParams, []interface{}{userEmail, *app.Team})
	}
	if storage.DB.Exec(dbQuery, dbParams...).RecordNotFound() {
		return nil, NewUnauthorizedError("team not found or user dont have permission to do actions with the team provided")
	}

	// applying the operations to App
	if err = updateAppEnvVars(app, operations); err != nil {
		log.Printf(`error found when updating the env vars for the App "%s". Err: %s`, appName, err)
		return nil, err
	}

	// TODO: update the deployment here if exists...

	if err = c.updateNamespace(ns, app, userEmail); err != nil {
		log.Printf(`error when updating the namespace "%s". Err: %s`, appName, err)
		return nil, err
	}
	log.Printf(`env vars updated with success for the App "%s" by the user "%s"`, appName, userEmail)
	return
}

// createAppQuota creates an k8s Limit Range for the App (namespace)
func (c apps) createAppQuota(appName string, lr *api.LimitRange) error {
	_, err := c.k.k8sClient.LimitRanges(appName).Create(lr)
	return err
}

// createAppNamespace creates an k8s namespace
// Inside kubernetes, every app is a k8s namespaces (1:1) with the App information inside
func (c apps) createAppNamespace(ns *api.Namespace) error {
	if _, err := c.k.k8sClient.Namespaces().Create(ns); err != nil {
		if k8s_errors.IsAlreadyExists(err) {
			return NewAlreadyExistsErrorf(`already exists an app (aka namespace) with this name "%s"`, ns.GetName())
		}
		return err
	}
	return nil
}

// updateNamespace updates App information inside the namespace
func (c apps) updateNamespace(ns *api.Namespace, app *models.App, userEmail string) error {
	ai, err := json.Marshal(app)
	if err != nil {
		return err
	}
	ns.Annotations["teresa.io/app"] = string(ai)
	ns.Annotations["teresa.io/last-user"] = userEmail
	if _, err := c.k.k8sClient.Namespaces().Update(ns); err != nil {
		return fmt.Errorf(`error when updating the namespace "%s". Err: %s`, *app.Name, err)
	}
	return nil
}

// createAppStorageSecret creates a K8s Secret that will be used by the Builder and Runner processes
func (c apps) createAppStorageSecret(appName string, storage helpers.Storage) error {
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
	return err
}

// getNamespace returns a namespace or an error if any
func (c apps) getNamespace(name string) (ns *api.Namespace, err error) {
	ns, err = c.k.k8sClient.Namespaces().Get(name)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return nil, NewNotFoundErrorf(`namespace "%s" not found`, name)
		}
		return nil, err
	}
	return
}

// validateBeforeCreate validade all App parameters and return an InputError if any
func validateBeforeCreate(app *models.App, userEmail string) error {
	// TODO: validate...
	// - name
	// - team
	// - userEmail
	if app.Limits == nil {
		return NewInputError("limits where not provided")
	}
	return nil
}

// newAppNamespaceYaml is a helper to create k8s namespace (parameters) for an App
func newAppNamespaceYaml(app *models.App, userEmail string) (ns *api.Namespace) {
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
	return
}

// addAppToNamespaceYaml marshall the entire App struct and put inside the namesapce Yaml
func addAppToNamespaceYaml(app *models.App, ns *api.Namespace) error {
	ai, err := json.Marshal(app)
	if err != nil {
		return err
	}
	ns.Annotations["teresa.io/app"] = string(ai)
	return nil
}

// addLimitRangeQuantityToResourceList is a helper to the function parseLimitRangeParams, used to add a limit range
// to a specific limit range list
func addLimitRangeQuantityToResourceList(r *api.ResourceList, limitRangeQuantity []*models.LimitRangeQuantity) error {
	if limitRangeQuantity == nil {
		return nil
	}
	rl := api.ResourceList{}
	for _, item := range limitRangeQuantity {
		name := api.ResourceName(*item.Resource)
		q, err := resource.ParseQuantity(*item.Quantity)
		if err != nil {
			return fmt.Errorf(`error when trying to parse limits value "%s:%s". Err: %s`, *item.Resource, *item.Quantity, err)
		}
		rl[name] = q
	}
	*r = rl
	return nil
}

// parseLimitRangeParams is a helper to parse all the limit range types
func parseLimitRangeParams(limitRangeItem *api.LimitRangeItem, limits *models.AppInLimits) error {
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.Default, limits.Default); err != nil {
		return err
	}
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.DefaultRequest, limits.DefaultRequest); err != nil {
		return err
	}
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.Max, limits.Max); err != nil {
		return err
	}
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.Min, limits.Min); err != nil {
		return err
	}
	if err := addLimitRangeQuantityToResourceList(&limitRangeItem.MaxLimitRequestRatio, limits.LimitRequestRatio); err != nil {
		return err
	}
	return nil
}

// newAppQuotaYaml is a helper to create an k8s LimitRange based on App Limits
func newAppQuotaYaml(app *models.App) (lr *api.LimitRange, err error) {
	lrItem := api.LimitRangeItem{
		Type: api.LimitTypeContainer,
	}
	// parse limits params to k8s params
	if err = parseLimitRangeParams(&lrItem, app.Limits); err != nil {
		return nil, NewInputErrorf(`error found when parsing "limits". Err.: %s`, err)
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
	return
}

// unmarshalAppFromNamespace extract and unmarshal the App from the namespace
func unmarshalAppFromNamespace(ns *api.Namespace) (app *models.App, err error) {
	s, ok := ns.GetAnnotations()["teresa.io/app"]
	if ok == false {
		return nil, fmt.Errorf(`app not found on this provided namespace "%s"`, ns.Name)
	}
	app = &models.App{}
	err = json.Unmarshal([]byte(s), app)
	return
}

// checkForProtectedEnvVars check if is there any Operation trying to modify some of the protected env vars
func checkForProtectedEnvVars(operations []*models.PatchAppRequest) error {
	protectedEnvVars := [...]string{"SLUG_URL", "PORT", "DEIS_DEBUG", "BUILDER_STORAGE"}
	for _, operation := range operations {
		for _, operationValue := range operation.Value {
			for _, pv := range protectedEnvVars {
				if *operationValue.Key == pv {
					return NewInputErrorf(`it is not allowed to manual change the env var "%s"`)
				}
			}
		}
	}
	return nil
}

// updateAppEnvVars updates App env vars based on the operations
func updateAppEnvVars(app *models.App, operations []*models.PatchAppRequest) error {
	if err := checkForProtectedEnvVars(operations); err != nil {
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
	return nil
}
