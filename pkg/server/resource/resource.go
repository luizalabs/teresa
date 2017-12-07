package resource

import (
	"bytes"
	"fmt"
	"io"

	respb "github.com/luizalabs/teresa/pkg/protobuf/resource"
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Resource struct {
	Name     string     `json:"name"`
	TeamName string     `json:"team_name"`
	Settings []*Setting `json:"settings"`
}

type Operations interface {
	Create(user *database.User, res *Resource) (string, error)
	Delete(user *database.User, resName string) error
}

type Templater interface {
	Template(resName string) (string, error)
	WelcomeTemplate(resName string) (string, error)
}

type Executer interface {
	Execute(wr io.Writer, tpl string, settings []*Setting) error
}

type K8sOperations interface {
	CreateNamespaceFromName(nsName, teamName, userEmail string) error
	CreateResources(nsName string, r io.Reader) error
	DeleteNamespace(nsName string) error
	IsAlreadyExists(err error) bool
	IsNotFound(err error) bool
}

type ResourceOperations struct {
	tpl    Templater
	exe    Executer
	k8s    K8sOperations
	appOps app.Operations
}

func newResource(req *respb.CreateRequest) *Resource {
	var s []*Setting
	for _, tmp := range req.Settings {
		s = append(s, &Setting{Key: tmp.Key, Value: tmp.Value})
	}
	return &Resource{Name: req.Name, TeamName: req.TeamName, Settings: s}
}

func (ops *ResourceOperations) namespace(resName string) string {
	return fmt.Sprintf("%s-resource", resName)
}

func (ops *ResourceOperations) Create(user *database.User, res *Resource) (_ string, Err error) {
	if !ops.appOps.HasPermission(user, res.Name) {
		return "", auth.ErrPermissionDenied
	}

	nsName := ops.namespace(res.Name)
	if err := ops.k8s.CreateNamespaceFromName(nsName, res.TeamName, user.Email); err != nil {
		if ops.k8s.IsAlreadyExists(err) {
			return "", ErrAlreadyExists
		}
		return "", teresa_errors.NewInternalServerError(err)
	}

	defer func() {
		if Err != nil {
			ops.k8s.DeleteNamespace(nsName)
		}
	}()

	tpl, err := ops.tpl.Template(res.Name)
	if err != nil {
		return "", teresa_errors.NewInternalServerError(err)
	}

	welcomeTpl, err := ops.tpl.WelcomeTemplate(res.Name)
	if err != nil {
		return "", teresa_errors.NewInternalServerError(err)
	}

	var buf bytes.Buffer
	if err := ops.exe.Execute(&buf, tpl, res.Settings); err != nil {
		return "", teresa_errors.NewInternalServerError(err)
	}

	var welcomeBuf bytes.Buffer
	if err := ops.exe.Execute(&welcomeBuf, welcomeTpl, res.Settings); err != nil {
		return "", teresa_errors.NewInternalServerError(err)
	}

	r := bytes.NewReader(buf.Bytes())
	if err := ops.k8s.CreateResources(nsName, r); err != nil {
		return "", teresa_errors.NewInternalServerError(err)
	}

	return string(welcomeBuf.Bytes()), nil
}

func (ops *ResourceOperations) Delete(user *database.User, resName string) error {
	if !ops.appOps.HasPermission(user, resName) {
		return auth.ErrPermissionDenied
	}

	nsName := ops.namespace(resName)
	if err := ops.k8s.DeleteNamespace(nsName); err != nil {
		if ops.k8s.IsNotFound(err) {
			return ErrNotFound
		}
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func NewOperations(tpl Templater, exe Executer, k8s K8sOperations, appOps app.Operations) Operations {
	return &ResourceOperations{tpl: tpl, exe: exe, k8s: k8s, appOps: appOps}
}
