package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/team"
)

type Operations interface {
	Create(user *storage.User, app *App) error
	Logs(user *storage.User, appName string, lines int64, follow bool) (io.ReadCloser, error)
	Info(user *storage.User, appName string) (*Info, error)
	TeamName(appName string) (string, error)
	Meta(appName string) (*App, error)
}

type K8sOperations interface {
	NamespaceAnnotation(namespace, annotation string) (string, error)
	NamespaceLabel(namespace, label string) (string, error)
	PodList(namespace string) ([]*Pod, error)
	PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error)
	CreateNamespace(app *App, userEmail string) error
	CreateQuota(app *App) error
	CreateSecret(appName, secretName string, data map[string][]byte) error
	CreateAutoScale(app *App) error
	AddressList(namespace string) ([]*Address, error)
	Status(namespace string) (*Status, error)
	AutoScale(namespace string) (*AutoScale, error)
	Limits(namespace, name string) (*Limits, error)
	IsNotFound(err error) bool
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
)

func (ops *AppOperations) hasPerm(user *storage.User, team string) bool {
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

func (ops *AppOperations) Create(user *storage.User, app *App) error {
	if !ops.hasPerm(user, app.Team) {
		return auth.ErrPermissionDenied
	}

	if err := ops.kops.CreateNamespace(app, user.Email); err != nil {
		return err
	}

	if err := ops.kops.CreateQuota(app); err != nil {
		return err
	}

	secretName := ops.st.K8sSecretName()
	data := ops.st.AccessData()
	if err := ops.kops.CreateSecret(app.Name, secretName, data); err != nil {
		return err
	}

	return ops.kops.CreateAutoScale(app)
}

func (ops *AppOperations) Logs(user *storage.User, appName string, lines int64, follow bool) (io.ReadCloser, error) {
	team, err := ops.kops.NamespaceLabel(appName, TeresaTeamLabel)
	if err != nil {
		return nil, err
	}

	if !ops.hasPerm(user, team) {
		return nil, auth.ErrPermissionDenied
	}

	pods, err := ops.kops.PodList(appName)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()
	var wg sync.WaitGroup
	for _, pod := range pods {
		wg.Add(1)
		go func(namespace, podName string) {
			defer wg.Done()

			logs, err := ops.kops.PodLogs(namespace, podName, lines, follow)
			if err != nil {
				log.Errorf("streaming logs from pod %s: %v", podName, err)
				return
			}
			defer logs.Close()

			scanner := bufio.NewScanner(logs)
			for scanner.Scan() {
				fmt.Fprintf(w, "[%s] - %s\n", podName, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				log.Errorf("streaming logs from pod %s: %v", podName, err)
			}
		}(appName, pod.Name)
	}
	go func() {
		wg.Wait()
		w.Close()
	}()

	return r, nil
}

func (ops *AppOperations) Info(user *storage.User, appName string) (*Info, error) {
	teamName, err := ops.TeamName(appName)
	if err != nil {
		return nil, err
	}

	if !ops.hasPerm(user, teamName) {
		err := fmt.Errorf("permission denied user %s on team %s", user.Name, teamName)
		return nil, newAppErr(auth.ErrPermissionDenied, err)
	}

	appMeta, err := ops.Meta(appName)
	if err != nil {
		return nil, err
	}

	addr, err := ops.kops.AddressList(appName)
	if err != nil {
		return nil, newAppErr(ErrUnknown, err)
	}

	stat, err := ops.kops.Status(appName)
	if err != nil {
		return nil, newAppErr(ErrUnknown, err)
	}

	as, err := ops.kops.AutoScale(appName)
	if err != nil {
		return nil, newAppErr(ErrUnknown, err)
	}

	lim, err := ops.kops.Limits(appName, limitsName)
	if err != nil {
		return nil, newAppErr(ErrUnknown, err)
	}

	info := &Info{
		Team:      teamName,
		Addresses: addr,
		Status:    stat,
		AutoScale: as,
		Limits:    lim,
		EnvVars:   appMeta.EnvVars,
	}
	return info, nil
}

func (ops *AppOperations) TeamName(appName string) (string, error) {
	teamName, err := ops.kops.NamespaceLabel(appName, TeresaTeamLabel)
	if err != nil {
		if ops.kops.IsNotFound(err) {
			return "", newAppErr(ErrNotFound, err)
		}
		return "", newAppErr(ErrUnknown, err)
	}
	return teamName, nil
}

func (ops *AppOperations) Meta(appName string) (*App, error) {
	an, err := ops.kops.NamespaceAnnotation(appName, TeresaAnnotation)
	if err != nil {
		if ops.kops.IsNotFound(err) {
			return nil, newAppErr(ErrNotFound, err)
		}
		return nil, newAppErr(ErrUnknown, err)
	}
	a := new(App)
	if err := json.Unmarshal([]byte(an), a); err != nil {
		err = fmt.Errorf("unmarshal app failed: %v", err)
		return nil, newAppErr(ErrUnknown, err)
	}

	return a, nil
}

func NewOperations(tops team.Operations, kops K8sOperations, st st.Storage) Operations {
	return &AppOperations{tops: tops, kops: kops, st: st}
}
