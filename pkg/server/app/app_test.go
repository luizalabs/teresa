package app

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"k8s.io/client-go/pkg/api"

	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/slug"
	st "github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/team"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

type fakeK8sOperations struct {
	Namespaces map[string]struct{}
}

type errK8sOperations struct {
	Err                        error
	NamespaceErr               error
	QuotaErr                   error
	SecretErr                  error
	AutoscaleErr               error
	DeleteNamespaceErr         error
	SetNamespaceAnnotationsErr error
	SetNamespaceLabelsErr      error
	Namespaces                 map[string]struct{}
}

func (*fakeK8sOperations) CreateNamespace(app *App, user string) error {
	return nil
}

func (*fakeK8sOperations) CreateQuota(app *App) error {
	return nil
}

func (*fakeK8sOperations) CreateSecret(appName, secretName string, data map[string][]byte) error {
	return nil
}

func (*fakeK8sOperations) PodList(namespace string) ([]*Pod, error) {
	pl := []*Pod{
		{Name: "pod 1", State: string(api.PodRunning), Age: 2, Restarts: 0},
		{Name: "pod 2", State: string(api.PodRunning), Age: 5, Restarts: 1},
	}
	return pl, nil
}

func (*fakeK8sOperations) PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error) {
	r := bytes.NewBufferString("foo\nbar")
	return ioutil.NopCloser(r), nil
}

func (*fakeK8sOperations) NamespaceAnnotation(namespace, annotation string) (string, error) {
	return `{"name": "test"}`, nil
}

func (*fakeK8sOperations) NamespaceLabel(namespace, label string) (string, error) {
	return "luizalabs", nil
}

func (*fakeK8sOperations) CreateOrUpdateAutoscale(app *App) error {
	return nil
}

func (*fakeK8sOperations) AddressList(namespace string) ([]*Address, error) {
	addr := []*Address{{Hostname: "host1"}}
	return addr, nil
}

func (*fakeK8sOperations) Status(namespace string) (*Status, error) {
	stat := &Status{
		CPU: 33,
		Pods: []*Pod{
			{Name: "pod 1", State: string(api.PodRunning), Age: 1, Restarts: 1},
			{Name: "pod 2", State: string(api.PodPending), Age: 2, Restarts: 2},
			{Name: "pod 3", State: string(api.PodRunning), Age: 3, Restarts: 3},
		},
	}
	return stat, nil
}

func (*fakeK8sOperations) Autoscale(namespace string) (*Autoscale, error) {
	as := &Autoscale{CPUTargetUtilization: 42, Max: 10, Min: 1}
	return as, nil
}

func (*fakeK8sOperations) Limits(namespace, name string) (*Limits, error) {
	lrq1 := &LimitRangeQuantity{Quantity: "1", Resource: "resource1"}
	lrq2 := &LimitRangeQuantity{Quantity: "2", Resource: "resource2"}
	lrq3 := &LimitRangeQuantity{Quantity: "3", Resource: "resource3"}
	lrq4 := &LimitRangeQuantity{Quantity: "4", Resource: "resource4"}
	lim := &Limits{
		Default:        []*LimitRangeQuantity{lrq1, lrq2},
		DefaultRequest: []*LimitRangeQuantity{lrq3, lrq4},
	}
	return lim, nil
}

func (*fakeK8sOperations) IsAlreadyExists(err error) bool {
	return true
}

func (*fakeK8sOperations) IsNotFound(err error) bool {
	return true
}

func (*fakeK8sOperations) SetNamespaceAnnotations(namespace string, annotations map[string]string) error {
	return nil
}

func (*fakeK8sOperations) SetNamespaceLabels(namespace string, labels map[string]string) error {
	return nil
}

func (*fakeK8sOperations) DeleteDeployEnvVars(namespace, name string, evNames []string) error {
	return nil
}

func (*fakeK8sOperations) CreateOrUpdateDeployEnvVars(namespace, name string, evs []*EnvVar) error {
	return nil
}

func (f *fakeK8sOperations) DeleteNamespace(namespace string) error {
	delete(f.Namespaces, namespace)
	return nil
}

func (f *fakeK8sOperations) NamespaceListByLabel(label, value string) ([]string, error) {
	ns := make([]string, 0)
	for s := range f.Namespaces {
		ns = append(ns, s)
	}
	return ns, nil
}

func (e *errK8sOperations) CreateNamespace(app *App, user string) error {
	return e.NamespaceErr
}

func (e *errK8sOperations) CreateQuota(app *App) error {
	return e.QuotaErr
}

func (e *errK8sOperations) CreateSecret(appName, secretName string, data map[string][]byte) error {
	return e.SecretErr
}

func (e *errK8sOperations) CreateOrUpdateAutoscale(app *App) error {
	return e.AutoscaleErr
}

func (e *errK8sOperations) PodList(namespace string) ([]*Pod, error) {
	return nil, e.Err
}

func (e *errK8sOperations) PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error) {
	return nil, e.Err
}

func (e *errK8sOperations) NamespaceAnnotation(namespace, annotation string) (string, error) {
	return `{"name": "test"}`, e.Err
}

func (e *errK8sOperations) NamespaceLabel(namespace, label string) (string, error) {
	return "luizalabs", e.Err
}

func (e *errK8sOperations) AddressList(namespace string) ([]*Address, error) {
	return nil, e.Err
}

func (e *errK8sOperations) Status(namespace string) (*Status, error) {
	return nil, e.Err
}

func (e *errK8sOperations) Autoscale(namespace string) (*Autoscale, error) {
	return nil, e.Err
}

func (e *errK8sOperations) Limits(namespace, name string) (*Limits, error) {
	return nil, e.Err
}

func (*errK8sOperations) IsAlreadyExists(err error) bool {
	return true
}

func (*errK8sOperations) IsNotFound(err error) bool {
	return true
}

func (e *errK8sOperations) SetNamespaceAnnotations(namespace string, annotations map[string]string) error {
	return e.SetNamespaceAnnotationsErr
}

func (e *errK8sOperations) SetNamespaceLabels(namespace string, annotations map[string]string) error {
	return e.SetNamespaceLabelsErr
}

func (e *errK8sOperations) DeleteDeployEnvVars(namespace, name string, evNames []string) error {
	return e.Err
}

func (e *errK8sOperations) CreateOrUpdateDeployEnvVars(namespace, name string, evs []*EnvVar) error {
	return e.Err
}

func (e *errK8sOperations) DeleteNamespace(namespace string) error {
	delete(e.Namespaces, namespace)
	return e.DeleteNamespaceErr
}

func (e *errK8sOperations) NamespaceListByLabel(label, value string) ([]string, error) {
	return nil, e.Err
}

func TestAppOperationsCreate(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &database.Team{
		Name:  name,
		Users: []database.User{*user},
	}

	if err := ops.Create(user, app); err != nil {
		t.Fatal("error creating app: ", err)
	}
}

func TestAppOperationsCreateErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}

	if err := ops.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppCreateErrPermissionDeniedShouldNotTouchNamespace(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	name := "teresa"
	fakeK8s := &fakeK8sOperations{Namespaces: map[string]struct{}{name: {}}}
	ops := NewOperations(tops, fakeK8s, fakeSt)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: name, Team: "luizalabs"}

	if err := ops.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}

	if _, ok := fakeK8s.Namespaces[name]; !ok {
		t.Errorf("expected namespace %s, got none", name)
	}
}

func TestAppOperationsCreateErrAppAlreadyExists(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &database.Team{
		Name:  name,
		Users: []database.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{NamespaceErr: ErrAlreadyExists}

	if err := ops.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected %v got %v", ErrAlreadyExists, err)
	}
}

func TestAppCreateErrAppAlreadyExistsShouldNotTouchNamespace(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	name := "teresa"
	teamName := "luizalabs"
	errK8s := &errK8sOperations{
		NamespaceErr: ErrAlreadyExists,
		Namespaces:   map[string]struct{}{name: {}},
	}
	ops := NewOperations(tops, errK8s, fakeSt)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: name, Team: teamName}
	tops.(*team.FakeOperations).Storage[teamName] = &database.Team{
		Name:  teamName,
		Users: []database.User{*user},
	}

	if err := ops.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected %v got %v", ErrAlreadyExists, err)
	}

	if _, ok := errK8s.Namespaces[name]; !ok {
		t.Errorf("expected namespace %s, got none", name)
	}
}

func TestAppTeamName(t *testing.T) {
	ops := NewOperations(team.NewFakeOperations(), &fakeK8sOperations{}, st.NewFake())
	teamName, err := ops.TeamName("teresa")
	if err != nil {
		t.Error("got error on get teamName", err)
	}
	if teamName != "luizalabs" { //see fakeK8sOperations
		t.Errorf("expected luizalabs, got %s", teamName)
	}
}

func TestAppMeta(t *testing.T) {
	ops := NewOperations(team.NewFakeOperations(), &fakeK8sOperations{}, st.NewFake())
	a, err := ops.Get("teresa")
	if err != nil {
		t.Errorf("got error on get app Meta: %v", err)

	}
	if a.Name != "test" {
		t.Errorf("expected luizalabs, got %s", a.Name)
	}
}

func TestAppOperationsHasPermission(t *testing.T) {
	appName := "teresa"
	goodUserEmail := "teresa@luizalabs.com"

	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	teamName := "luizalabs"
	user := &database.User{Email: goodUserEmail}
	tops.(*team.FakeOperations).Storage[teamName] = &database.Team{
		Name:  teamName,
		Users: []database.User{*user},
	}

	var testCases = []struct {
		email    string
		expected bool
	}{
		{goodUserEmail, true},
		{"bad-user@luizalabs.com", false},
	}

	for _, tc := range testCases {
		u := &database.User{Email: tc.email}
		actual := ops.HasPermission(u, appName)
		if tc.expected != actual {
			t.Errorf("expected %v, got %v", tc.expected, actual)
		}
	}
}

func TestAppOperationsLogs(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	name := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &database.Team{
		Name:  name,
		Users: []database.User{*user},
	}

	rc, err := ops.Logs(user, app.Name, 10, false)
	if err != nil {
		t.Fatal("error on get logs: ", err)
	}
	defer rc.Close()

	count := 0
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, "[pod ") {
			t.Errorf("expected log with pod name, got %s", text)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal("error on read logs:", err)
	}

	if count != 4 { // see fakeK8sOperations.PodLogs
		t.Errorf("expected 4, got %d", count)
	}
}

func TestAppOperationsLogsErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Logs(user, "teresa", 10, false); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppOperationsLogsErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Logs(user, "teresa", 10, false); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsCreateErrQuota(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &database.Team{
		Name:  name,
		Users: []database.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{QuotaErr: errors.New("Quota Error")}

	if ops.Create(user, app) == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsCreateErrSecret(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &database.Team{
		Name:  name,
		Users: []database.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{SecretErr: errors.New("Secret Error")}

	if ops.Create(user, app) == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsCreateErrAutoscale(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &database.Team{
		Name:  name,
		Users: []database.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{AutoscaleErr: errors.New("Autoscale Error")}
	if ops.Create(user, app) == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsInfo(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	teamName := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: teamName}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  teamName,
		Users: []database.User{*user},
	}

	info, err := ops.Info(user, app.Name)
	if err != nil {
		t.Fatal("error getting app info: ", err)
	}

	if info.Team != teamName {
		t.Errorf("expected %s, got %s", teamName, info.Team)
	}

	if len(info.Addresses) != 1 { // see fakeK8sOperations.AddressList
		t.Errorf("expected 2, got %d", len(info.Addresses))
	}

	if info.Status.CPU != 33 { // see fakeK8sOperations.Status
		t.Errorf("expected 33, got %d", info.Status.CPU)
	}

	if info.Autoscale.CPUTargetUtilization != 42 { // see fakeK8sOperations.Autoscale
		t.Errorf("expected 42, got %d", info.Autoscale.CPUTargetUtilization)
	}

	ndef := len(info.Limits.Default)
	if ndef != 2 { // see fakeK8sOperations.Limits
		t.Errorf("expected 2, got %d", ndef)
	}

	ndefReq := len(info.Limits.DefaultRequest)
	if ndefReq != 2 {
		t.Errorf("expected 2, got %d", ndefReq)
	}
}

func TestAppOperationsInfoErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Info(user, "teresa"); teresa_errors.Get(err) != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", teresa_errors.Get(err))
	}
}

func TestAppOperationsInfoErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Info(user, "teresa"); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", teresa_errors.Get(err))
	}
}

func TestAppOperationsList(t *testing.T) {
	tops := team.NewFakeOperations()
	appName := "teresa"
	teamName := "luizalabs"

	user := &database.User{Email: "teresa@luizalabs.com"}
	fk8s := &fakeK8sOperations{Namespaces: map[string]struct{}{appName: {}}}

	ops := NewOperations(tops, fk8s, nil)
	tops.(*team.FakeOperations).Storage[appName] = &database.Team{
		Name:  teamName,
		Users: []database.User{*user},
	}

	apps, err := ops.List(user)
	if err != nil {
		t.Fatal("error getting app list:", err)
	}

	if len(apps) == 0 {
		t.Fatal("expected at least one app")
	}
	for _, a := range apps {
		if a.Name != appName {
			t.Errorf("expected %s, got %s", appName, a.Name)
		}
		if a.Team != teamName {
			t.Errorf("expected %s, got %s", teamName, a.Team)
		}
		if len(a.Addresses) != 1 { // see fakeK8sOperations.AddressList
			t.Errorf("expected 1 address, got %d", len(a.Addresses))
		}
	}
}

func TestAppOperationsSetEnv(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	evs := []*EnvVar{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	if err := ops.SetEnv(user, app.Name, evs); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsSetEnvErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetEnv(user, "teresa", nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsSetEnvErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetEnv(user, "teresa", nil); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsSetEnvErrInternalServerErrorOnSaveApp(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{SetNamespaceAnnotationsErr: errors.New("test")}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.SetEnv(user, app.Name, nil); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestAppOperationsUnsetEnv(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	evs := []string{"key1", "key2"}

	if err := ops.UnsetEnv(user, app.Name, evs); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsUnsetEnvErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.UnsetEnv(user, "teresa", nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsUnsetEnvErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.UnsetEnv(user, "teresa", nil); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsSetEnvProtectedVar(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	evs := make([]*EnvVar, len(slug.ProtectedEnvVars))
	for i := range evs {
		evs[i] = &EnvVar{Key: slug.ProtectedEnvVars[i], Value: "test"}
	}

	if err := ops.SetEnv(user, app.Name, evs); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsUnSetEnvProtectedVar(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.UnsetEnv(user, app.Name, slug.ProtectedEnvVars[:]); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsUnsetEnvErrInternalServerErrorOnSaveApp(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{SetNamespaceAnnotationsErr: errors.New("test")}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.UnsetEnv(user, app.Name, nil); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestAppOperationsSetAutoscale(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	req := newAutoscaleRequest("teresa")
	as := newAutoscale(req)

	if err := ops.SetAutoscale(user, app.Name, as); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsSetAutoscaleErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetAutoscale(user, "teresa", nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsSetAutoscaleErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetAutoscale(user, "teresa", nil); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsSetAutoscaleErrInternalServerErrorOnSaveApp(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{SetNamespaceAnnotationsErr: errors.New("test")}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	req := newAutoscaleRequest("teresa")
	as := newAutoscale(req)

	if err := ops.SetAutoscale(user, app.Name, as); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestAppOperationsDelete(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.Delete(user, app.Name); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsDeleteErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.Delete(user, ""); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsDeleteErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.Delete(user, "teresa"); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsChangeTeam(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.ChangeTeam(user, app.Name, "gopher"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsChangeTeamErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.ChangeTeam(user, "teresa", "gopher"); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsChangeTeamErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.ChangeTeam(user, "gophers", "teresa"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
