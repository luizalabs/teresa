package k8s

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	// "github.com/luizalabs/tapi/handlers"

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
	Create(app *models.App, storage helpers.Storage, tk *Token) error
	Update(app *models.App, tk *Token) error
	UpdateEnvVars(appName string, operations []*models.PatchAppRequest, tk *Token) (app *models.App, err error)
	Get(appName string, tk *Token) (app *models.App, err error)
}

type apps struct {
	k *k8sHelper
}

func newApps(c *k8sHelper) *apps {
	return &apps{k: c}
}

// Create creates an App inside kubernetes
// Inside kubernetes, the App is represented as an namespace
func (c apps) Create(app *models.App, storage helpers.Storage, tk *Token) error {
	log.Printf(`creating App "%s"`, *app.Name)
	// validating input params
	if err := validateBeforeCreate(app, tk); err != nil {
		return err
	}
	// check if user can create apps for the team
	if tk.IsAuthorized(*app.Team) == false {
		log.Printf(`token "%s" is not allowed to create Apps for team "%s"`, *tk.Email, *app.Team)
		return NewUnauthorizedError("token not allowed to create apps for this team")
	}
	app.Creator = &models.User{
		Email: tk.Email,
	}

	// creating namespace
	log.Printf(`creating namespace "%s"`, *app.Name)
	if err := c.createNamespace(app, *tk.Email); err != nil {
		return err
	}
	log.Printf(`namespace "%s" created with success`, *app.Name)

	// creating quota (limit ranges) for namespace
	log.Printf(`creating quota (limit range) for namespace "%s"`, *app.Name)
	if err := c.createQuota(app); err != nil {
		return err
	}
	log.Printf(`namespace quota created with success for namespace "%s"`, *app.Name)

	// creating storage secret. this will be used to store the builded App
	log.Printf(`creating storage secret for namespace "%s"`, *app.Name)
	if err := c.createStorageSecret(*app.Name, storage); err != nil {
		return err
	}
	log.Printf(`secret created with success for namespace "%s"`, *app.Name)

	log.Printf(`app "%s" created with success by user "%s" for team "%s"`, *app.Name, *tk.Email, *app.Team)
	return nil
}

func (c apps) Update(app *models.App, tk *Token) error {

	// TODO: update the deployment here if exists...
	// TODO: update namespace quota here if exists...

	if err := c.updateNamespace(app, *tk.Email); err != nil {
		return err
	}
	return nil
}

func (c apps) UpdateEnvVars(appName string, operations []*models.PatchAppRequest, tk *Token) (app *models.App, err error) {
	log.Printf(`updating env vars for App "%s"`, appName)
	// getting app
	app, err = c.Get(appName, tk)
	if err != nil {
		if IsUnauthorizedError(err) {
			log.Printf(`token "%s" is not allowed to update env vars for the App "%s"`, *tk.Email, *app.Name)
		}
		return nil, err
	}

	// applying the operations to App
	if err = updateAppEnvVars(app, operations); err != nil {
		return nil, err
	}

	if err := c.Update(app, tk); err != nil {
		return nil, err
	}

	log.Printf(`env vars updated with success for the App "%s/%s" by the user "%s"`, *app.Team, *app.Name, *tk.Email)
	return
}

// Get returns an App by the name
func (c apps) Get(appName string, tk *Token) (app *models.App, err error) {
	ns, err := c.getNamespace(appName)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, NewUnauthorizedErrorf(`app "%s" not found or user is not allowed to see the same`, appName)
		}
		return nil, err
	}
	app, err = unmarshalAppFromNamespace(ns)
	if err != nil {
		return nil, err
	}
	// check if the user is authorized to get this App
	if tk.IsAuthorized(*app.Team) == false {
		log.Printf(`user token "%s" is not allowed to see the App "%s"`, *tk.Email, appName)
		return nil, NewUnauthorizedErrorf(`app "%s" not found or user is not allowed to see the same`, appName)
	}
	return
}

// createQuota creates an k8s Limit Range for the App (namespace)
func (c apps) createQuota(app *models.App) error {
	quota, err := newQuotaYaml(app)
	if err != nil {
		return err
	}
	if _, err = c.k.k8sClient.LimitRanges(*app.Name).Create(quota); err != nil {
		log.Printf(`error when creating "quotas" for the namespace "%s". Err: %s`, *app.Name, err)
		return err
	}
	return nil
}

// createNamespace creates an k8s namespace
// Inside kubernetes, every app is a k8s namespaces (1:1) with the App information inside
func (c apps) createNamespace(app *models.App, userEmail string) error {
	nsy := newAppNamespaceYaml(app, userEmail)
	if err := addAppToNamespaceYaml(app, nsy); err != nil {
		return err
	}
	if _, err := c.k.k8sClient.Namespaces().Create(nsy); err != nil {
		if k8s_errors.IsAlreadyExists(err) {
			msg := fmt.Sprintf(`already exists a namespace with the name "%s"`, *app.Name)
			log.Print(msg)
			return NewAlreadyExistsError(msg)
		}
		log.Printf(`error found when creating the namespace "%s". Err: %s`, *app.Name, err)
		return err
	}
	return nil
}

// updateNamespace updates App information inside the namespace
func (c apps) updateNamespace(app *models.App, userEmail string) error {
	ns, err := c.getNamespace(*app.Name)
	if err != nil {
		return err
	}
	ai, err := json.Marshal(app)
	if err != nil {
		log.Printf(`error when updating the namespace "%s". Err: %s`, *app.Name, err)
		return err
	}
	ns.Annotations["teresa.io/app"] = string(ai)
	ns.Annotations["teresa.io/last-user"] = userEmail
	if _, err := c.k.k8sClient.Namespaces().Update(ns); err != nil {
		log.Printf(`error when updating the namespace "%s". Err: %s`, *app.Name, err)
		return err
	}
	return nil
}

// createStorageSecret creates a K8s Secret that will be used by the Builder and Runner processes
func (c apps) createStorageSecret(appName string, storage helpers.Storage) error {
	log.Printf(`creating secret for namespace "%s"`, appName)
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
		log.Printf(`error found when creating service for the namespace "%s". Err: %s`, appName, err)
		return err
	}
	return nil
}

// getNamespace returns a namespace or an error if any
func (c apps) getNamespace(name string) (ns *api.Namespace, err error) {
	ns, err = c.k.k8sClient.Namespaces().Get(name)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			log.Printf(`namespace "%s" not found`, name)
			return nil, NewNotFoundErrorf(`"%s" not found`, name)
		}
		log.Printf(`error when trying to get the namespace "%s", Err: %s`, name, err)
		return nil, err
	}
	return
}

// validateBeforeCreate validade all App parameters and return an InputError if any
func validateBeforeCreate(app *models.App, tk *Token) error {
	// TODO: validate...
	// - name
	// - team
	// - userEmail
	if app.Limits == nil {
		log.Printf(`quota not specified for the App %s`, *app.Name)
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
		log.Printf(`error found when marshalling the app to put inside the namespace annotation. App "%s". Err: %s`, *app.Name, err)
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
			log.Printf(`error when trying to parse limits value "%s:%s". Err: %s`, *item.Resource, *item.Quantity, err)
			return err
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

// newQuotaYaml is a helper to create an k8s LimitRange based on App Limits
func newQuotaYaml(app *models.App) (lr *api.LimitRange, err error) {
	lrItem := api.LimitRangeItem{
		Type: api.LimitTypeContainer,
	}
	// parse limits params to k8s params
	if err = parseLimitRangeParams(&lrItem, app.Limits); err != nil {
		log.Printf(`error found when parsing "limits" for the App "%s"`, *app.Name)
		return nil, NewInputError(`error found when parsing "limits"`)
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
		msg := fmt.Sprintf(`annotation "teresa.io/app" not found on this provided namespace "%s"`, ns.Name)
		log.Print(msg)
		return nil, errors.New(msg)
	}
	app = &models.App{}
	err = json.Unmarshal([]byte(s), app)
	if err != nil {
		log.Printf(`error when trying to unmarshal the app from namespace "%s". Err: %s`, ns.Name, err)
		return nil, err
	}
	return
}

// checkForProtectedEnvVars check if is there any Operation trying to modify some of the protected env vars
func checkForProtectedEnvVars(operations []*models.PatchAppRequest) error {
	protectedEnvVars := [...]string{"SLUG_URL", "PORT", "DEIS_DEBUG", "BUILDER_STORAGE"}
	for _, operation := range operations {
		for _, operationValue := range operation.Value {
			for _, pv := range protectedEnvVars {
				if *operationValue.Key == pv {
					msg := fmt.Sprintf(`manually changing the env var "%s" isn't allowed`, pv)
					log.Print(msg)
					return NewInputError(msg)
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
