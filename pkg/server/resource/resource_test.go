package resource

import (
	"errors"
	"io"
	"testing"

	respb "github.com/luizalabs/teresa/pkg/protobuf/resource"
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/luizalabs/teresa/pkg/server/test"
)

type fakeReader struct{}

func (f *fakeReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

type fakeTemplater struct {
	Err        error
	WelcomeErr error
}

func (f *fakeTemplater) Template(resName string) (io.Reader, error) {
	return &fakeReader{}, f.Err
}

func (f *fakeTemplater) WelcomeTemplate(resName string) (io.Reader, error) {
	return &fakeReader{}, f.WelcomeErr
}

type fakeExecuter struct {
	Err error
}

func (f *fakeExecuter) Execute(w io.Writer, r io.Reader, settings []*Setting) error {
	w.Write([]byte("Test Representation"))
	return f.Err
}

type fakeK8sOperations struct {
	CreateNamespaceErr error
	ResourcesErr       error
	DeleteNamespaceErr error
	IsAlreadyExistsErr bool
	IsNotFoundErr      bool
}

func (f *fakeK8sOperations) CreateNamespaceFromName(nsName, teamName, userEmail string) error {
	return f.CreateNamespaceErr
}

func (f *fakeK8sOperations) CreateResources(nsName string, r io.Reader) error {
	return f.ResourcesErr
}

func (f *fakeK8sOperations) DeleteNamespace(nsName string) error {
	return f.DeleteNamespaceErr
}

func (f *fakeK8sOperations) IsAlreadyExists(err error) bool {
	return f.IsAlreadyExistsErr
}

func (f *fakeK8sOperations) IsNotFound(err error) bool {
	return f.IsNotFoundErr
}

func testResource() *Resource {
	return &Resource{
		Name:     "teresa",
		TeamName: "luizalabs",
		Settings: []*Setting{
			&Setting{Key: "key1", Value: "value1"},
			&Setting{Key: "key2", Value: "value2"},
		},
	}
}

func TestNewResource(t *testing.T) {
	s1 := &respb.CreateRequest_Setting{Key: "key1", Value: "value1"}
	s2 := &respb.CreateRequest_Setting{Key: "key2", Value: "value2"}
	req := &respb.CreateRequest{
		Name:     "test",
		TeamName: "luizalabs",
		Settings: []*respb.CreateRequest_Setting{s1, s2},
	}

	res := newResource(req)

	if !test.DeepEqual(req, res) {
		t.Errorf("expected %v, got %v", req, res)
	}
}

func TestOperationsCreateSuccess(t *testing.T) {
	tpl := &fakeTemplater{}
	exe := &fakeExecuter{}
	k8s := &fakeK8sOperations{}
	appOps := app.NewFakeOperations()
	ops := NewOperations(tpl, exe, k8s, appOps)
	res := testResource()
	user := &database.User{Email: "gopher@luizalabs.com"}

	if _, err := ops.Create(user, res); err != nil {
		t.Error("got error creating resource:", err)
	}
}

func TestOperationsCreateErrPermissionDenied(t *testing.T) {
	tpl := &fakeTemplater{}
	exe := &fakeExecuter{}
	k8s := &fakeK8sOperations{}
	appOps := app.NewFakeOperations()
	ops := NewOperations(tpl, exe, k8s, appOps)
	res := testResource()
	user := &database.User{Email: "bad-user@luizalabs.com"}

	if _, err := ops.Create(user, res); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestOperationsCreateErrAlreadyExists(t *testing.T) {
	tpl := &fakeTemplater{}
	exe := &fakeExecuter{}
	k8s := &fakeK8sOperations{
		CreateNamespaceErr: errors.New("test"),
		IsAlreadyExistsErr: true,
	}
	appOps := app.NewFakeOperations()
	ops := NewOperations(tpl, exe, k8s, appOps)
	res := testResource()
	user := &database.User{Email: "gopher@luizalabs.com"}

	if _, err := ops.Create(user, res); err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestOperationsCreateErrInternalServerError(t *testing.T) {
	var testCases = []struct {
		tpl Templater
		exe Executer
		k8s K8sOperations
	}{
		{&fakeTemplater{Err: errors.New("test")}, &fakeExecuter{}, &fakeK8sOperations{}},
		{&fakeTemplater{WelcomeErr: errors.New("test")}, &fakeExecuter{}, &fakeK8sOperations{}},
		{&fakeTemplater{}, &fakeExecuter{Err: errors.New("test")}, &fakeK8sOperations{}},
		{&fakeTemplater{}, &fakeExecuter{}, &fakeK8sOperations{ResourcesErr: errors.New("test")}},
	}
	appOps := app.NewFakeOperations()
	user := &database.User{Email: "gopher@luizalabs.com"}
	res := testResource()

	for _, tc := range testCases {
		ops := NewOperations(tc.tpl, tc.exe, tc.k8s, appOps)

		if _, err := ops.Create(user, res); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
			t.Errorf("expected ErrInternalServerError, got %v", err)
		}
	}
}

func TestOperationsDeleteSuccess(t *testing.T) {
	tpl := &fakeTemplater{}
	exe := &fakeExecuter{}
	k8s := &fakeK8sOperations{}
	appOps := app.NewFakeOperations()
	ops := NewOperations(tpl, exe, k8s, appOps)
	user := &database.User{Email: "gopher@luizalabs.com"}

	if err := ops.Delete(user, "test"); err != nil {
		t.Error("got error deleting resource:", err)
	}
}

func TestOperationsDeleteErrPermissionDenied(t *testing.T) {
	tpl := &fakeTemplater{}
	exe := &fakeExecuter{}
	k8s := &fakeK8sOperations{}
	appOps := app.NewFakeOperations()
	ops := NewOperations(tpl, exe, k8s, appOps)
	user := &database.User{Email: "bad-user@luizalabs.com"}

	if err := ops.Delete(user, "test"); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestOperationsDeleteErrNotFound(t *testing.T) {
	tpl := &fakeTemplater{}
	exe := &fakeExecuter{}
	k8s := &fakeK8sOperations{
		DeleteNamespaceErr: errors.New("test"),
		IsNotFoundErr:      true,
	}
	appOps := app.NewFakeOperations()
	ops := NewOperations(tpl, exe, k8s, appOps)
	user := &database.User{Email: "gopher@luizalabs.com"}

	if err := ops.Delete(user, "test"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestOperationsDeleteErrInternalServerError(t *testing.T) {
	tpl := &fakeTemplater{}
	exe := &fakeExecuter{}
	k8s := &fakeK8sOperations{DeleteNamespaceErr: errors.New("test")}
	appOps := app.NewFakeOperations()
	ops := NewOperations(tpl, exe, k8s, appOps)
	user := &database.User{Email: "gopher@luizalabs.com"}

	if err := ops.Delete(user, "test"); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}
