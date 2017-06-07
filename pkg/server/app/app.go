package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"k8s.io/client-go/pkg/api"

	log "github.com/Sirupsen/logrus"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/team"
)

type Operations interface {
	Create(user *storage.User, app *App) error
	Logs(user *storage.User, appName string, lines int64, follow bool) (io.ReadCloser, error)
}

type K8sOperations interface {
	Create(app *App, st st.Storage) error
	NamespaceAnnotation(namespace, annotation string) (string, error)
	PodList(namespace string) ([]*Pod, error)
	PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error)
}

type AppOperations struct {
	tops team.Operations
	kops K8sOperations
	st   st.Storage
}

const (
	TeresaAnnotation = "teresa.io/app"
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

func (ops *AppOperations) getAppTeam(appName string) (string, error) {
	annotation, err := ops.kops.NamespaceAnnotation(appName, TeresaAnnotation)
	if err != nil {
		return "", err
	}
	a := new(App)
	if err := json.Unmarshal([]byte(annotation), a); err != nil {
		return "", err
	}
	return a.Team, nil
}

func (ops *AppOperations) Create(user *storage.User, app *App) error {
	if !ops.hasPerm(user, app.Team) {
		return auth.ErrPermissionDenied
	}
	return ops.kops.Create(app, ops.st)
}

func (ops *AppOperations) Logs(user *storage.User, appName string, lines int64, follow bool) (io.ReadCloser, error) {
	team, err := ops.getAppTeam(appName)
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
		if pod.State != string(api.PodRunning) {
			continue
		}

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

func NewOperations(tops team.Operations, kops K8sOperations, st st.Storage) Operations {
	return &AppOperations{tops: tops, kops: kops, st: st}
}
