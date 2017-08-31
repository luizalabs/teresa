package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/slug"
	st "github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/team"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

type Operations interface {
	Create(user *database.User, app *App) error
	Logs(user *database.User, appName string, lines int64, follow bool) (io.ReadCloser, error)
	Info(user *database.User, appName string) (*Info, error)
	TeamName(appName string) (string, error)
	Get(appName string) (*App, error)
	HasPermission(user *database.User, appName string) bool
	SetEnv(user *database.User, appName string, evs []*EnvVar) error
	UnsetEnv(user *database.User, appName string, evs []string) error
	List(user *database.User) ([]*AppListItem, error)
	SetAutoscale(user *database.User, appName string, as *Autoscale) error
	CheckPermAndGet(user *database.User, appName string) (*App, error)
}

type K8sOperations interface {
	NamespaceAnnotation(namespace, annotation string) (string, error)
	NamespaceLabel(namespace, label string) (string, error)
	PodList(namespace string) ([]*Pod, error)
	PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error)
	CreateNamespace(app *App, userEmail string) error
	CreateQuota(app *App) error
	CreateSecret(appName, secretName string, data map[string][]byte) error
	CreateOrUpdateAutoscale(app *App) error
	AddressList(namespace string) ([]*Address, error)
	Status(namespace string) (*Status, error)
	Autoscale(namespace string) (*Autoscale, error)
	Limits(namespace, name string) (*Limits, error)
	IsNotFound(err error) bool
	IsAlreadyExists(err error) bool
	SetNamespaceAnnotations(namespace string, annotations map[string]string) error
	DeleteDeployEnvVars(namespace, name string, evNames []string) error
	CreateOrUpdateDeployEnvVars(namespace, name string, evs []*EnvVar) error
	DeleteNamespace(namespace string) error
	NamespaceListByLabel(label, value string) ([]string, error)
}

type AppOperations struct {
	tops team.Operations
	kops K8sOperations
	st   st.Storage
}

const (
	limitsName       = "limits"
	TeresaAnnotation = "teresa.io/app"
	TeresaTeamLabel  = "teresa.io/team"
	TeresaLastUser   = "teresa.io/last-user"
)

func (ops *AppOperations) hasPerm(user *database.User, team string) bool {
	teams, err := ops.tops.ListByUser(user.Email)
	if err != nil {
		return false
	}
	var found bool
	for _, t := range teams {
		if t.Name == team {
			found = true
			break
		}
	}
	return found
}

func (ops *AppOperations) HasPermission(user *database.User, appName string) bool {
	teamName, err := ops.TeamName(appName)
	if err != nil {
		return false
	}
	return ops.hasPerm(user, teamName)
}

func (ops *AppOperations) Create(user *database.User, app *App) (Err error) {
	if !ops.hasPerm(user, app.Team) {
		return auth.ErrPermissionDenied
	}

	if err := ops.kops.CreateNamespace(app, user.Email); err != nil {
		if ops.kops.IsAlreadyExists(err) {
			return ErrAlreadyExists
		}
		return teresa_errors.NewInternalServerError(err)
	}

	defer func() {
		if Err != nil {
			ops.kops.DeleteNamespace(app.Name)
		}
	}()

	if err := ops.kops.CreateQuota(app); err != nil {
		return teresa_errors.New(ErrInvalidLimits, err)
	}

	secretName := ops.st.K8sSecretName()
	data := ops.st.AccessData()
	if err := ops.kops.CreateSecret(app.Name, secretName, data); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	if err := ops.kops.CreateOrUpdateAutoscale(app); err != nil {
		return teresa_errors.New(ErrInvalidAutoscale, err)
	}

	return nil
}

func (ops *AppOperations) Logs(user *database.User, appName string, lines int64, follow bool) (io.ReadCloser, error) {
	team, err := ops.kops.NamespaceLabel(appName, TeresaTeamLabel)
	if err != nil {
		if ops.kops.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, teresa_errors.NewInternalServerError(err)
	}

	if !ops.hasPerm(user, team) {
		return nil, auth.ErrPermissionDenied
	}

	pods, err := ops.kops.PodList(appName)
	if err != nil {
		return nil, teresa_errors.NewInternalServerError(err)
	}

	r, w := io.Pipe()
	var wg sync.WaitGroup
	for _, pod := range pods {
		wg.Add(1)
		go func(namespace, podName string) {
			defer wg.Done()

			logs, err := ops.kops.PodLogs(namespace, podName, lines, follow)
			if err != nil {
				log.WithError(err).Errorf("streaming logs from pod %s", podName)
				return
			}
			defer logs.Close()

			scanner := bufio.NewScanner(logs)
			for scanner.Scan() {
				fmt.Fprintf(w, "[%s] - %s\n", podName, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				log.WithError(err).Errorf("streaming logs from pod %s", podName)
			}
		}(appName, pod.Name)
	}
	go func() {
		wg.Wait()
		w.Close()
	}()

	return r, nil
}

func (ops *AppOperations) Info(user *database.User, appName string) (*Info, error) {
	teamName, err := ops.TeamName(appName)
	if err != nil {
		return nil, err
	}

	if !ops.hasPerm(user, teamName) {
		return nil, auth.ErrPermissionDenied
	}

	appMeta, err := ops.Get(appName)
	if err != nil {
		return nil, err
	}

	addr, err := ops.kops.AddressList(appName)
	if err != nil {
		return nil, teresa_errors.NewInternalServerError(err)
	}

	stat, err := ops.kops.Status(appName)
	if err != nil {
		return nil, teresa_errors.NewInternalServerError(err)
	}

	as, err := ops.kops.Autoscale(appName)
	if err != nil {
		return nil, teresa_errors.NewInternalServerError(err)
	}

	lim, err := ops.kops.Limits(appName, limitsName)
	if err != nil {
		return nil, teresa_errors.NewInternalServerError(err)
	}

	info := &Info{
		Team:      teamName,
		Addresses: addr,
		Status:    stat,
		Autoscale: as,
		Limits:    lim,
		EnvVars:   appMeta.EnvVars,
	}
	return info, nil
}

func (ops *AppOperations) TeamName(appName string) (string, error) {
	teamName, err := ops.kops.NamespaceLabel(appName, TeresaTeamLabel)
	if err != nil {
		if ops.kops.IsNotFound(err) {
			return "", ErrNotFound
		}
		return "", teresa_errors.NewInternalServerError(err)
	}
	return teamName, nil
}

func (ops *AppOperations) Get(appName string) (*App, error) {
	an, err := ops.kops.NamespaceAnnotation(appName, TeresaAnnotation)
	if err != nil {
		if ops.kops.IsNotFound(err) {
			return nil, teresa_errors.New(ErrNotFound, err)
		}
		return nil, teresa_errors.NewInternalServerError(err)
	}
	a := new(App)
	if err := json.Unmarshal([]byte(an), a); err != nil {
		err = fmt.Errorf("unmarshal app failed: %v", err)
		return nil, teresa_errors.NewInternalServerError(err)
	}

	return a, nil
}

func (ops *AppOperations) CheckPermAndGet(user *database.User, appName string) (*App, error) {
	team, err := ops.TeamName(appName)
	if err != nil {
		return nil, err
	}

	if !ops.hasPerm(user, team) {
		return nil, auth.ErrPermissionDenied
	}

	return ops.Get(appName)
}

func (ops *AppOperations) saveApp(app *App, lastUser string) error {
	b, err := json.Marshal(app)
	if err != nil {
		return fmt.Errorf("marshal app failed: %v", err)
	}

	anMap := map[string]string{
		TeresaAnnotation: string(b),
		TeresaLastUser:   lastUser,
	}

	return ops.kops.SetNamespaceAnnotations(app.Name, anMap)
}

func (ops *AppOperations) SetEnv(user *database.User, appName string, evs []*EnvVar) error {
	evNames := make([]string, len(evs))
	for i, _ := range evs {
		evNames[i] = evs[i].Key
	}
	if err := checkForProtectedEnvVars(evNames); err != nil {
		return err
	}

	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	setEnvVars(app, evs)

	if err := ops.saveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	if err = ops.kops.CreateOrUpdateDeployEnvVars(appName, appName, evs); err != nil {
		if ops.kops.IsNotFound(err) {
			return nil
		}
		return teresa_errors.NewInternalServerError(err)
	}
	return nil
}

func (ops *AppOperations) UnsetEnv(user *database.User, appName string, evNames []string) error {
	if err := checkForProtectedEnvVars(evNames); err != nil {
		return err
	}

	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	unsetEnvVars(app, evNames)

	if err := ops.saveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	if err = ops.kops.DeleteDeployEnvVars(appName, appName, evNames); err != nil {
		if ops.kops.IsNotFound(err) {
			return nil
		}
		return teresa_errors.NewInternalServerError(err)
	}
	return nil
}

func checkForProtectedEnvVars(evsNames []string) error {
	for _, name := range slug.ProtectedEnvVars {
		for _, item := range evsNames {
			if name == item {
				return ErrProtectedEnvVar
			}
		}
	}
	return nil
}

func (ops *AppOperations) List(user *database.User) ([]*AppListItem, error) {
	teams, err := ops.tops.ListByUser(user.Email)
	if err != nil {
		return nil, err
	}
	items := make([]*AppListItem, 0)
	for _, team := range teams {
		apps, err := ops.kops.NamespaceListByLabel(TeresaTeamLabel, team.Name)
		if err != nil {
			return nil, err
		}
		for _, a := range apps {
			addrs, err := ops.kops.AddressList(a)
			if err != nil {
				return nil, err
			}
			items = append(items, &AppListItem{
				Team:      team.Name,
				Name:      a,
				Addresses: addrs,
			})
		}
	}
	return items, nil
}

func (ops *AppOperations) SetAutoscale(user *database.User, appName string, as *Autoscale) error {
	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	old, err := ops.kops.Autoscale(appName)
	if err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	if c := as.CPUTargetUtilization; c < 0 || c > 100 {
		as.CPUTargetUtilization = old.CPUTargetUtilization
	}
	app.Autoscale = as

	if err := ops.kops.CreateOrUpdateAutoscale(app); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	if err := ops.saveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func NewOperations(tops team.Operations, kops K8sOperations, st st.Storage) Operations {
	return &AppOperations{tops: tops, kops: kops, st: st}
}
