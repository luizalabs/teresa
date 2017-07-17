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

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/slug"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/team"
)

type fakeK8sOperations struct{}

type errK8sOperations struct {
	Err          error
	NamespaceErr error
	QuotaErr     error
	SecretErr    error
	AutoScaleErr error
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
		{Name: "pod 1", State: string(api.PodRunning)},
		{Name: "pod 2", State: string(api.PodRunning)},
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

func (*fakeK8sOperations) CreateAutoScale(app *App) error {
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
			{Name: "pod 1", State: string(api.PodRunning)},
			{Name: "pod 2", State: string(api.PodPending)},
			{Name: "pod 3", State: string(api.PodRunning)},
		},
	}
	return stat, nil
}

func (*fakeK8sOperations) AutoScale(namespace string) (*AutoScale, error) {
	as := &AutoScale{CPUTargetUtilization: 42, Max: 10, Min: 1}
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

func (*fakeK8sOperations) IsNotFound(err error) bool {
	return true
}

func (*fakeK8sOperations) SetNamespaceAnnotations(namespace string, annotations map[string]string) error {
	return nil
}

func (*fakeK8sOperations) DeleteDeployEnvVars(namespace, name string, evNames []string) error {
	return nil
}

func (*fakeK8sOperations) CreateOrUpdateDeployEnvVars(namespace, name string, evs []*EnvVar) error {
	return nil
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

func (e *errK8sOperations) CreateAutoScale(app *App) error {
	return e.AutoScaleErr
}

func (e *errK8sOperations) PodList(namespace string) ([]*Pod, error) {
	return nil, e.Err
}

func (e *errK8sOperations) PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error) {
	return nil, e.Err
}

func (e *errK8sOperations) NamespaceAnnotation(namespace, annotation string) (string, error) {
	return "", e.Err
}

func (e *errK8sOperations) NamespaceLabel(namespace, label string) (string, error) {
	return "", e.Err
}

func (e *errK8sOperations) AddressList(namespace string) ([]*Address, error) {
	return nil, e.Err
}

func (e *errK8sOperations) Status(namespace string) (*Status, error) {
	return nil, e.Err
}

func (e *errK8sOperations) AutoScale(namespace string) (*AutoScale, error) {
	return nil, e.Err
}

func (e *errK8sOperations) Limits(namespace, name string) (*Limits, error) {
	return nil, e.Err
}

func (*errK8sOperations) IsNotFound(err error) bool {
	return true
}

func (e *errK8sOperations) SetNamespaceAnnotations(namespace string, annotations map[string]string) error {
	return e.Err
}

func (e *errK8sOperations) DeleteDeployEnvVars(namespace, name string, evNames []string) error {
	return e.Err
}

func (e *errK8sOperations) CreateOrUpdateDeployEnvVars(namespace, name string, evs []*EnvVar) error {
	return e.Err
}

func TestAppOperationsCreate(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
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
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}

	if err := ops.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppOperationsCreateErrAppAlreadyExists(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{NamespaceErr: ErrAlreadyExists}

	if err := ops.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected ErrAppAlreadyExists got %s", err)
	}
}

func TestAppTeamName(t *testing.T) {
	ops := NewOperations(team.NewFakeOperations(), &fakeK8sOperations{}, st.NewFake())
	teamName, err := ops.TeamName("teresa")
	if err != nil {
		t.Error("got error on get teamName:", err)
	}
	if teamName != "luizalabs" { //see fakeK8sOperations
		t.Errorf("expected luizalabs, got %s", teamName)
	}
}

func TestAppMeta(t *testing.T) {
	ops := NewOperations(team.NewFakeOperations(), &fakeK8sOperations{}, st.NewFake())
	a, err := ops.Get("teresa")
	if err != nil {
		t.Errorf("got error on get app Meta:", err)
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
	user := &storage.User{Email: goodUserEmail}
	tops.(*team.FakeOperations).Storage[teamName] = &storage.Team{
		Name:  teamName,
		Users: []storage.User{*user},
	}

	var testCases = []struct {
		email    string
		expected bool
	}{
		{goodUserEmail, true},
		{"bad-user@luizalabs.com", false},
	}

	for _, tc := range testCases {
		u := &storage.User{Email: tc.email}
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
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
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
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Logs(user, "teresa", 10, false); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppOperationsLogsErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Logs(user, "teresa", 10, false); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsCreateErrQuota(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
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
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{SecretErr: errors.New("Secret Error")}

	if ops.Create(user, app) == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsCreateErrAutoScale(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{AutoScaleErr: errors.New("AutoScale Error")}

	if ops.Create(user, app) == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsInfo(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	teamName := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: teamName}
	tops.(*team.FakeOperations).Storage[app.Name] = &storage.Team{
		Name:  teamName,
		Users: []storage.User{*user},
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

	if info.AutoScale.CPUTargetUtilization != 42 { // see fakeK8sOperations.AutoScale
		t.Errorf("expected 42, got %d", info.AutoScale.CPUTargetUtilization)
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
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Info(user, "teresa"); grpcErr(err) != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", grpcErr(err))
	}
}

func TestAppOperationsInfoErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Info(user, "teresa"); grpcErr(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", grpcErr(err))
	}
}

func TestAppOperationsSetEnv(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &storage.Team{
		Name:  app.Team,
		Users: []storage.User{*user},
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
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetEnv(user, "teresa", nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsSetEnvErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetEnv(user, "teresa", nil); grpcErr(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsUnsetEnv(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &storage.Team{
		Name:  app.Team,
		Users: []storage.User{*user},
	}
	evs := []string{"key1", "key2"}

	if err := ops.UnsetEnv(user, app.Name, evs); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsUnsetEnvErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if err := ops.UnsetEnv(user, "teresa", nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsUnsetEnvErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if err := ops.UnsetEnv(user, "teresa", nil); grpcErr(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsSetEnvProtectedVar(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &storage.Team{
		Name:  app.Team,
		Users: []storage.User{*user},
	}
	evs := make([]*EnvVar, len(slug.ProtectedEnvVars))
	for i, _ := range evs {
		evs[i] = &EnvVar{Key: slug.ProtectedEnvVars[i], Value: "test"}
	}

	if err := ops.SetEnv(user, app.Name, evs); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsUnSetEnvProtectedVar(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Name] = &storage.Team{
		Name:  app.Team,
		Users: []storage.User{*user},
	}

	if err := ops.UnsetEnv(user, app.Name, slug.ProtectedEnvVars[:]); err == nil {
		t.Errorf("expected error, got nil")
	}
}
