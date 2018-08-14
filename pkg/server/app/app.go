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

const SecretPath = "/teresa/secrets"

type Operations interface {
	Create(user *database.User, app *App) error
	Logs(user *database.User, appName string, opts *LogOptions) (io.ReadCloser, error)
	Info(user *database.User, appName string) (*Info, error)
	TeamName(appName string) (string, error)
	Get(appName string) (*App, error)
	HasPermission(user *database.User, appName string) bool
	SetEnv(user *database.User, appName string, evs []*EnvVar) error
	UnsetEnv(user *database.User, appName string, evs []string) error
	SetSecret(user *database.User, appName string, secrets []*EnvVar) error
	UnsetSecret(user *database.User, appName string, secrets []string) error
	SetSecretFile(user *database.User, appName, name string, content []byte) error
	List(user *database.User) ([]*AppListItem, error)
	ListByTeam(teamName string) ([]string, error)
	SetAutoscale(user *database.User, appName string, as *Autoscale) error
	CheckPermAndGet(user *database.User, appName string) (*App, error)
	SaveApp(app *App, lastUser string) error
	Delete(user *database.User, appName string) error
	ChangeTeam(appName, teamName string) error
	SetReplicas(user *database.User, appName string, replicas int32) error
	DeletePods(user *database.User, appName string, podsNames []string) error
}

type K8sOperations interface {
	NamespaceAnnotation(namespace, annotation string) (string, error)
	NamespaceLabel(namespace, label string) (string, error)
	PodList(namespace string, opts *PodListOptions) ([]*Pod, error)
	PodLogs(namespace, podName string, opts *LogOptions) (io.ReadCloser, error)
	CreateNamespace(app *App, userEmail string) error
	CreateQuota(app *App) error
	GetSecret(namespace, secretName string) (map[string][]byte, error)
	CreateOrUpdateSecret(appName, secretName string, data map[string][]byte) error
	CreateOrUpdateAutoscale(app *App) error
	AddressList(namespace string) ([]*Address, error)
	Status(namespace string) (*Status, error)
	Autoscale(namespace string) (*Autoscale, error)
	Limits(namespace, name string) (*Limits, error)
	IsNotFound(err error) bool
	IsAlreadyExists(err error) bool
	IsInvalid(err error) bool
	IsUnknown(err error) bool
	SetNamespaceAnnotations(namespace string, annotations map[string]string) error
	SetNamespaceLabels(namespace string, labels map[string]string) error
	DeleteDeployEnvVars(namespace, name string, evNames []string) error
	DeleteCronJobEnvVars(namespace, name string, evNames []string) error
	CreateOrUpdateDeployEnvVars(namespace, name string, evs []*EnvVar) error
	CreateOrUpdateCronJobEnvVars(namespace, name string, evs []*EnvVar) error
	CreateOrUpdateDeploySecretEnvVars(namespace, name, secretName string, secrets []string) error
	CreateOrUpdateCronJobSecretEnvVars(namespace, name, secretName string, secrets []string) error
	DeleteNamespace(namespace string) error
	NamespaceListByLabel(label, value string) ([]string, error)
	DeploySetReplicas(namespace, name string, replicas int32) error
	DeletePod(namespace, podName string) error
	HasIngress(namespace, name string) (bool, error)
	IngressEnabled() bool
	CreateOrUpdateDeploySecretFile(namespace, deploy, fileName string) error
	CreateOrUpdateCronJobSecretFile(namespace, cronjob, filename string) error
	DeleteDeploySecrets(namespace, deploy string, envVars, volKeys []string) error
	DeleteCronJobSecrets(namespace, cronjob string, envVars, volKeys []string) error
	SuspendCronJob(namespace, name string) error
	ResumeCronJob(namespace, name string) error
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
	TeresaAppSecrets = "teresa-secrets"
)

func (ops *AppOperations) HasPermission(user *database.User, appName string) bool {
	teamName, err := ops.TeamName(appName)
	if err != nil {
		return false
	}
	hasPerm, err := ops.tops.HasUser(teamName, user.Email)
	if err != nil {
		return false
	}
	return hasPerm
}

func (ops *AppOperations) Create(user *database.User, app *App) (Err error) {
	hasPerm, err := ops.tops.HasUser(app.Team, user.Email)
	if err != nil || !hasPerm {
		return auth.ErrPermissionDenied
	}

	if ops.kops.IngressEnabled() && app.VirtualHost == "" && app.ProcessType == ProcessTypeWeb {
		return ErrMissingVirtualHost
	}

	if err := ops.kops.CreateNamespace(app, user.Email); err != nil {
		return ops.translateError(err)
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
	if err := ops.kops.CreateOrUpdateSecret(app.Name, secretName, data); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	if IsCronJob(app.ProcessType) {
		return nil
	}

	if err := ops.kops.CreateOrUpdateAutoscale(app); err != nil {
		return teresa_errors.New(ErrInvalidAutoscale, err)
	}

	return nil
}

func (ops *AppOperations) Logs(user *database.User, appName string, opts *LogOptions) (io.ReadCloser, error) {
	teamName, err := ops.kops.NamespaceLabel(appName, TeresaTeamLabel)
	if err != nil {
		return nil, ops.translateError(err)
	}

	hasPerm, err := ops.tops.HasUser(teamName, user.Email)
	if err != nil || !hasPerm {
		return nil, auth.ErrPermissionDenied
	}

	pods, err := ops.kops.PodList(appName, &PodListOptions{PodName: opts.PodName})
	if err != nil {
		return nil, teresa_errors.NewInternalServerError(err)
	}
	if opts.Container == "" {
		opts.Container = appName
	}

	r, w := io.Pipe()
	var wg sync.WaitGroup
	for _, pod := range pods {
		wg.Add(1)
		go func(namespace, podName string) {
			defer wg.Done()

			logs, err := ops.kops.PodLogs(namespace, podName, opts)
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

	hasPerm, err := ops.tops.HasUser(teamName, user.Email)
	if err != nil || !hasPerm {
		return nil, auth.ErrPermissionDenied
	}

	appMeta, err := ops.Get(appName)
	if err != nil {
		return nil, err
	}

	addrs, err := ops.addresses(appMeta)
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

	envVars := make([]*EnvVar, len(appMeta.EnvVars)+len(appMeta.Secrets))
	for i, ev := range appMeta.EnvVars {
		envVars[i] = &EnvVar{Key: ev.Key, Value: ev.Value}
	}
	for i, s := range appMeta.Secrets {
		envVars[len(appMeta.EnvVars)+i] = &EnvVar{
			Key: s, Value: "*****",
		}
	}
	vols := make([]string, len(appMeta.SecretFiles))
	for i := range appMeta.SecretFiles {
		vols[i] = fmt.Sprintf("%s/%s", SecretPath, appMeta.SecretFiles[i])
	}

	info := &Info{
		Team:      teamName,
		Addresses: addrs,
		Status:    stat,
		Autoscale: as,
		Limits:    lim,
		EnvVars:   envVars,
		Protocol:  appMeta.Protocol,
		Volumes:   vols,
	}
	return info, nil
}

func (ops *AppOperations) TeamName(appName string) (string, error) {
	teamName, err := ops.kops.NamespaceLabel(appName, TeresaTeamLabel)
	if err != nil {
		return "", ops.translateError(err)
	}
	return teamName, nil
}

func (ops *AppOperations) Get(appName string) (*App, error) {
	an, err := ops.kops.NamespaceAnnotation(appName, TeresaAnnotation)
	if err != nil {
		return nil, ops.translateError(err)
	}
	a := new(App)
	if err := json.Unmarshal([]byte(an), a); err != nil {
		err = fmt.Errorf("unmarshal app failed: %v", err)
		return nil, teresa_errors.NewInternalServerError(err)
	}

	return a, nil
}

func (ops *AppOperations) CheckPermAndGet(user *database.User, appName string) (*App, error) {
	teamName, err := ops.TeamName(appName)
	if err != nil {
		return nil, err
	}

	hasPerm, err := ops.tops.HasUser(teamName, user.Email)
	if err != nil || !hasPerm {
		return nil, auth.ErrPermissionDenied
	}

	return ops.Get(appName)
}

func (ops *AppOperations) SaveApp(app *App, lastUser string) error {
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
	for i := range evs {
		evNames[i] = evs[i].Key
	}
	if err := checkForProtectedEnvVars(evNames); err != nil {
		return err
	}

	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	if IsCronJob(app.ProcessType) {
		err = ops.kops.CreateOrUpdateCronJobEnvVars(appName, appName, evs)
	} else {
		err = ops.kops.CreateOrUpdateDeployEnvVars(appName, appName, evs)
	}

	if err != nil {
		if ops.kops.IsInvalid(err) {
			return ErrInvalidEnvVarName
		} else if !ops.kops.IsNotFound(err) {
			return teresa_errors.NewInternalServerError(err)
		}
	}

	setEnvVars(app, evs)

	if err := ops.SaveApp(app, user.Email); err != nil {
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

	if IsCronJob(app.ProcessType) {
		err = ops.kops.DeleteCronJobEnvVars(appName, appName, evNames)
	} else {
		err = ops.kops.DeleteDeployEnvVars(appName, appName, evNames)
	}

	if err != nil {
		if !ops.kops.IsNotFound(err) {
			return teresa_errors.NewInternalServerError(err)
		}
	}

	unsetEnvVars(app, evNames)

	if err := ops.SaveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func (ops *AppOperations) addresses(app *App) ([]*Address, error) {
	if app.Internal {
		return []*Address{{fmt.Sprintf("%s.%s", app.Name, app.Name)}}, nil
	}
	hasIngress, err := ops.kops.HasIngress(app.Name, app.Name)
	if err != nil {
		return nil, err
	}
	if hasIngress {
		return []*Address{{app.VirtualHost}}, nil
	}
	return ops.kops.AddressList(app.Name)
}

func (ops *AppOperations) SetSecretFile(user *database.User, appName, name string, content []byte) error {
	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	s, err := ops.kops.GetSecret(appName, TeresaAppSecrets)
	if err != nil {
		if !ops.kops.IsNotFound(err) {
			return teresa_errors.NewInternalServerError(err)
		}
	}
	if s == nil {
		s = make(map[string][]byte)
	}
	s[name] = content

	if err := ops.kops.CreateOrUpdateSecret(appName, TeresaAppSecrets, s); err != nil {
		if ops.kops.IsInvalid(err) {
			return ErrInvalidSecretName
		}
		return teresa_errors.NewInternalServerError(err)
	}

	if IsCronJob(app.ProcessType) {
		err = ops.kops.CreateOrUpdateCronJobSecretFile(appName, appName, name)
	} else {
		err = ops.kops.CreateOrUpdateDeploySecretFile(appName, appName, name)
	}

	if err != nil && !ops.kops.IsNotFound(err) {
		return teresa_errors.NewInternalServerError(err)
	}

	setSecretFileOnApp(app, name)

	if err := ops.SaveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func (ops *AppOperations) SetSecret(user *database.User, appName string, secrets []*EnvVar) error {
	names := make([]string, len(secrets))
	for i := range secrets {
		names[i] = secrets[i].Key
	}
	if err := checkForProtectedEnvVars(names); err != nil {
		return err
	}

	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	s, err := ops.kops.GetSecret(appName, TeresaAppSecrets)
	if err != nil {
		if !ops.kops.IsNotFound(err) {
			return teresa_errors.NewInternalServerError(err)
		}
	}
	if s == nil {
		s = make(map[string][]byte)
	}

	for _, secret := range secrets {
		s[secret.Key] = []byte(secret.Value)
	}

	if err := ops.kops.CreateOrUpdateSecret(appName, TeresaAppSecrets, s); err != nil {
		if ops.kops.IsInvalid(err) {
			return ErrInvalidSecretName
		}
		return teresa_errors.NewInternalServerError(err)
	}

	if IsCronJob(app.ProcessType) {
		err = ops.kops.CreateOrUpdateCronJobSecretEnvVars(appName, appName, TeresaAppSecrets, names)
	} else {
		err = ops.kops.CreateOrUpdateDeploySecretEnvVars(appName, appName, TeresaAppSecrets, names)
	}

	if err != nil {
		if ops.kops.IsInvalid(err) {
			return ErrInvalidSecretName
		} else if !ops.kops.IsNotFound(err) {
			return teresa_errors.NewInternalServerError(err)
		}
	}

	setSecretsOnApp(app, names)

	if err := ops.SaveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func (ops *AppOperations) UnsetSecret(user *database.User, appName string, secrets []string) error {
	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}
	envSecrets, fileSecrets := make([]string, 0), make([]string, 0)

Loop:
	for _, s := range secrets {
		for _, ev := range app.Secrets {
			if ev == s {
				envSecrets = append(envSecrets, ev)
				continue Loop
			}
		}
		for _, sf := range app.SecretFiles {
			if sf == s {
				fileSecrets = append(fileSecrets, sf)
				break
			}
		}
	}

	if len(envSecrets) > 0 {
		if err := checkForProtectedEnvVars(envSecrets); err != nil {
			return err
		}
	}

	s, err := ops.kops.GetSecret(appName, TeresaAppSecrets)
	if err != nil {
		if !ops.kops.IsNotFound(err) {
			return teresa_errors.NewInternalServerError(err)
		}
	}
	if s == nil {
		s = make(map[string][]byte)
	}

	for _, secret := range secrets {
		delete(s, secret)
	}

	if IsCronJob(app.ProcessType) {
		err = ops.kops.DeleteCronJobSecrets(appName, appName, envSecrets, fileSecrets)
	} else {
		err = ops.kops.DeleteDeploySecrets(appName, appName, envSecrets, fileSecrets)
	}

	if err != nil {
		if !ops.kops.IsNotFound(err) {
			return teresa_errors.NewInternalServerError(err)
		}
	}

	if len(envSecrets) > 0 {
		unsetSecretsOnApp(app, envSecrets)
	}
	if len(fileSecrets) > 0 {
		unsetSecretFilesOnApp(app, fileSecrets)
	}

	if err := ops.SaveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	// We remove secrets as last step to prevent errors on deploy/cron update
	if err := ops.kops.CreateOrUpdateSecret(appName, TeresaAppSecrets, s); err != nil {
		if ops.kops.IsInvalid(err) {
			return ErrInvalidSecretName
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
		apps, err := ops.ListByTeam(team.Name)
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

func (ops *AppOperations) ListByTeam(teamName string) ([]string, error) {
	return ops.kops.NamespaceListByLabel(TeresaTeamLabel, teamName)
}

func (ops *AppOperations) SetAutoscale(user *database.User, appName string, as *Autoscale) error {
	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	if IsCronJob(app.ProcessType) {
		return ErrInvalidActionForCronJob
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

	if err := ops.SaveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func (ops *AppOperations) Delete(user *database.User, appName string) error {
	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	if err := ops.kops.DeleteNamespace(app.Name); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func (ops *AppOperations) SetReplicas(user *database.User, appName string, replicas int32) error {
	app, err := ops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	if IsCronJob(app.ProcessType) {
		if replicas == 0 {
			err = ops.kops.SuspendCronJob(appName, appName)
		} else {
			err = ops.kops.ResumeCronJob(appName, appName)
		}
		if err != nil {
			return teresa_errors.NewInternalServerError(err)
		}
	} else if err := ops.kops.DeploySetReplicas(app.Name, app.Name, replicas); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

// ChangeTeam changes current team name of an App (be sure the new team exists)
func (ops *AppOperations) ChangeTeam(appName, teamName string) error {
	label := map[string]string{TeresaTeamLabel: teamName}
	if err := ops.kops.SetNamespaceLabels(appName, label); err != nil {
		return ops.translateError(err)
	}
	return nil
}

func (ops *AppOperations) DeletePods(user *database.User, appName string, podsNames []string) error {
	if _, err := ops.CheckPermAndGet(user, appName); err != nil {
		return err
	}

	for _, pod := range podsNames {
		if err := ops.kops.DeletePod(appName, pod); err != nil {
			if ops.kops.IsNotFound(err) {
				continue
			}
			return teresa_errors.NewInternalServerError(err)
		}
	}

	return nil
}

func (ops *AppOperations) translateError(err error) error {
	switch {
	case ops.kops.IsUnknown(err) || ops.kops.IsInvalid(err):
		return ErrInvalidName
	case ops.kops.IsNotFound(err):
		return ErrNotFound
	case ops.kops.IsAlreadyExists(err):
		return ErrAlreadyExists
	default:
		return teresa_errors.NewInternalServerError(err)
	}
}

func NewOperations(tops team.Operations, kops K8sOperations, st st.Storage) Operations {
	return &AppOperations{tops: tops, kops: kops, st: st}
}
