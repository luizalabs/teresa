package k8s

import (
	"encoding/json"
	"fmt"

	"github.com/luizalabs/tapi/helpers"
	"github.com/luizalabs/tapi/models"
	"k8s.io/kubernetes/pkg/api"
	k8s_errors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"
)

// AppsInterface is used to allow mock testing
type AppsInterface interface {
	Apps() AppInterface
}

// AppInterface is used to interact with Kubernetes and also to allow mock testing
type AppInterface interface {
	Create(app *models.App, storage helpers.Storage, tk *Token) error
	Update(app *models.App, storage helpers.Storage, tk *Token) error
	UpdateEnvVars(appName string, operations []*models.PatchAppRequest, storage helpers.Storage, tk *Token) (app *models.App, err error)
	UpdateScale(appName string, scale int64, storage helpers.Storage, tk *Token) (app *models.App, err error)
	UpdateAutoScale(appName string, autoScale *models.AutoScale, storage helpers.Storage, tk *Token) (app *models.App, err error)
	Get(appName string, tk *Token) (app *models.App, err error)
	List(tk *Token) (app []*models.App, err error)
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
	// validating input params
	if err := validateBeforeCreate(app, tk); err != nil {
		return err
	}
	// check if user can create apps for the team
	if tk.IsAuthorized(*app.Team) == false {
		return NewUnauthorizedError(`token "%s" not allowed to create apps for the team "%s"`, *tk.Email, *app.Team)
	}
	app.Creator = &models.User{
		Email: tk.Email,
	}
	// creating namespace
	if err := c.createNamespace(app, *tk.Email); err != nil {
		return err
	}
	// creating quota (limit ranges) for namespace
	if err := c.createQuota(app); err != nil {
		return err
	}
	// creating storage secret. this will be used to store the built App
	if err := c.createStorageSecret(*app.Name, storage); err != nil {
		return err
	}
	// creating first deployment (welcome project)
	if err := c.k.Deployments().CreateWelcomeDeployment(app); err != nil {
		return err
	}
	// creating horizontal auto scaling
	if err := c.k.Deployments().CreateAutoScale(app); err != nil {
		return err
	}
	// creating the loadbalancer
	if err := c.k.Networks().CreateLoadBalancerService(app); err != nil {
		return err
	}
	return nil
}

func (c apps) Update(app *models.App, storage helpers.Storage, tk *Token) error {
	// TODO: validate here
	//

	// ############################################################
	// FIXME: stopped this because it's not very usefull right now
	// ############################################################

	// // getting app
	// app, err := c.Get(*app.Name, tk)
	// if err != nil {
	// 	if IsUnauthorizedError(err) {
	// 		return NewUnauthorizedErrorf(`token "%s" is not allowed to update the app "%s". %s`, *tk.Email, *app.Name, err)
	// 	}
	// 	return err
	// }
	//
	// // TODO: update quota
	// // TODO: update hpa
	// //
	// //
	//
	// // updating deployment
	// if err := c.k.Deployments().Update(app, "update:app", storage); err != nil {
	// 	return err
	// }
	// if err := c.updateNamespace(app, *tk.Email); err != nil {
	// 	return err
	// }
	return nil
}

func (c apps) UpdateEnvVars(appName string, operations []*models.PatchAppRequest, storage helpers.Storage, tk *Token) (app *models.App, err error) {
	// getting app
	app, err = c.Get(appName, tk)
	if err != nil {
		if IsUnauthorizedError(err) {
			return nil, NewUnauthorizedErrorf(`token "%s" is not allowed to update env vars for the app "%s". %s`, *tk.Email, appName, err)
		}
		return nil, err
	}
	// applying the operations to App
	if err = updateAppEnvVars(app, operations); err != nil {
		return nil, err
	}
	// updating deployment
	if err := c.k.Deployments().Update(app, "update:env vars", storage); err != nil {
		return nil, err
	}
	if err := c.updateNamespace(app, *tk.Email); err != nil {
		return nil, err
	}
	return
}

func (c apps) UpdateScale(appName string, scale int64, storage helpers.Storage, tk *Token) (app *models.App, err error) {
	// getting app
	app, err = c.Get(appName, tk)
	if err != nil {
		if IsUnauthorizedError(err) {
			return nil, NewUnauthorizedErrorf(`token "%s" is not allowed to update scale for the app "%s". %s`, *tk.Email, appName, err)
		}
		return nil, err
	}
	// updating scale inside the app
	app.Scale = scale
	// updating deployment
	if err := c.k.Deployments().Update(app, "update:scale", storage); err != nil {
		return nil, err
	}
	if err := c.updateNamespace(app, *tk.Email); err != nil {
		return nil, err
	}
	return
}

func (c apps) UpdateAutoScale(appName string, autoScale *models.AutoScale, storage helpers.Storage, tk *Token) (app *models.App, err error) {
	// getting app
	app, err = c.Get(appName, tk)
	if err != nil {
		if IsUnauthorizedError(err) {
			return nil, NewUnauthorizedErrorf(`token "%s" is not allowed to update app "%s" auto scale info. %s`, *tk.Email, appName, err)
		}
		return nil, err
	}
	// updating app autoScale
	app.AutoScale = autoScale
	if err := c.k.Deployments().UpdateAutoScale(app); err != nil {
		return nil, err
	}
	if err := c.updateNamespace(app, *tk.Email); err != nil {
		return nil, err
	}
	return
}

// Get returns an App by the name
func (c apps) Get(appName string, tk *Token) (app *models.App, err error) {
	ns, err := c.getNamespace(appName)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, NewUnauthorizedErrorf(`app "%s" not found or user not allowed to see it`, appName)
		}
		return nil, err
	}
	app, err = unmarshalAppFromNamespace(ns)
	if err != nil {
		return nil, err
	}
	// check if the user is authorized to get this App
	if tk.IsAuthorized(*app.Team) == false {
		return nil, NewUnauthorizedErrorf(`app "%s" not found or user not allowed to see it`, appName)
	}
	lb, err := c.getLoadBalancer(appName)
	if err != nil {
		return nil, err
	}
	app.AddressList = []string{*lb}

	// TODO: get deployments here??

	return
}

func (c apps) getLoadBalancer(appName string) (lb *string, err error) {
	srv, err := c.k.Networks().GetService(appName)
	if err != nil {
		return nil, err
	}
	return &srv.Status.LoadBalancer.Ingress[0].Hostname, nil
}

// createQuota creates an k8s Limit Range for the App (namespace)
func (c apps) createQuota(app *models.App) error {
	quota, err := newQuotaYaml(app)
	if err != nil {
		return err
	}
	if _, err = c.k.k8sClient.LimitRanges(*app.Name).Create(quota); err != nil {
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
			return NewAlreadyExistsErrorf(`a namespace with the name "%s" already exists`, *app.Name)
		}
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
		return fmt.Errorf(`error when marshalling the app "%s". %s`, *app.Name, err)
	}
	ns.Annotations["teresa.io/app"] = string(ai)
	ns.Annotations["teresa.io/last-user"] = userEmail
	if _, err := c.k.k8sClient.Namespaces().Update(ns); err != nil {
		return err
	}
	return nil
}

// createStorageSecret creates a K8s Secret that will be used by the Builder and Runner processes
func (c apps) createStorageSecret(appName string, storage helpers.Storage) error {
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
		return err
	}
	return nil
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
func validateBeforeCreate(app *models.App, tk *Token) error {
	// TODO: validate...
	// - name
	// - team
	// - userEmail
	if app.Limits == nil {
		return NewInputErrorf(`limits where not provided for the app "%s"`, *app.Name)
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
			return fmt.Errorf(`error when trying to parse limits value "%s:%s". %s`, *item.Resource, *item.Quantity, err)
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
		return nil, NewInputErrorf(`found error when parsing "limits" for app "%s". %s`, *app.Name, err)
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
		return nil, fmt.Errorf(`annotation "teresa.io/app" not found inside the namespace "%s"`, ns.Name)
	}
	app = &models.App{}
	err = json.Unmarshal([]byte(s), app)
	if err != nil {
		return nil, fmt.Errorf(`error when trying to unmarshal the app from namespace "%s". %s`, ns.Name, err)
	}
	return
}

// checkForProtectedEnvVars check if is there any Operation trying to modify some of the protected env vars
func checkForProtectedEnvVars(operations []*models.PatchAppRequest) error {
	protectedEnvVars := [...]string{"SLUG_URL", "PORT", "DEIS_DEBUG", "BUILDER_STORAGE", "APP"}
	for _, operation := range operations {
		for _, operationValue := range operation.Value {
			for _, pv := range protectedEnvVars {
				if *operationValue.Key == pv {
					return NewInputError(`manually changing the env var "%s" isn't allowed`, pv)
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

func (c apps) List(tk *Token) (apps []*models.App, err error) {
	// list of teams from token
	s := sets.String{}
	for _, t := range tk.Teams {
		s.Insert(*t.Name)
	}
	// create label filter
	r, err := labels.NewRequirement("teresa.io/team", labels.InOperator, s)
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*r)
	list, err := c.k.k8sClient.Namespaces().List(api.ListOptions{LabelSelector: selector})
	if err != nil {
		// return err directly... there is no NotFoundError here
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, NewNotFoundErrorf(`no apps found for the token "%s"`, *tk.Email)
	}
	apps = []*models.App{}
	for _, ns := range list.Items {
		app, err := unmarshalAppFromNamespace(&ns)
		if err != nil {
			return nil, err
		}
		lb, err := c.getLoadBalancer(*app.Name)
		if err != nil {
			return nil, err
		}
		app.AddressList = []string{*lb}
		apps = append(apps, app)
	}
	return
}
