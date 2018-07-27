package app

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	k8sv1 "k8s.io/api/core/v1"

	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/slug"
	st "github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/team"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

type fakeK8sOperations struct {
	CreateOrUpdateAutoscaleWasCalled      bool
	CreateOrUpdateCronJobEnvVarsWasCalled bool
	DeleteCronJobEnvVarsWasCalled         bool
	Namespaces                            map[string]struct{}
	DefaultProcessType                    string
	AppInternal                           bool
	AppVirtualHost                        string
	AppIngress                            bool
	AppProtocol                           string
	IngressEnabledValue                   bool
}

type errK8sOperations struct {
	Err                                  error
	NamespaceErr                         error
	QuotaErr                             error
	SecretErr                            error
	AutoscaleErr                         error
	DeleteNamespaceErr                   error
	SetNamespaceAnnotationsErr           error
	SetNamespaceLabelsErr                error
	DeletePodErr                         error
	CreateOrUpdateDeployEnvVarsErr       error
	CreateOrUpdateDeploySecretEnvVarsErr error
	GetSecretErr                         error
	NegateIsNotFound                     bool
	NegateIsAlreadyExists                bool
	Namespaces                           map[string]struct{}
}

func (*fakeK8sOperations) CreateNamespace(app *App, user string) error {
	return nil
}

func (*fakeK8sOperations) CreateQuota(app *App) error {
	return nil
}

func (*fakeK8sOperations) CreateOrUpdateSecret(appName, secretName string, data map[string][]byte) error {
	return nil
}

func (*fakeK8sOperations) PodList(namespace string, opts *PodListOptions) ([]*Pod, error) {
	pl := []*Pod{
		{Name: "pod 1", State: string(k8sv1.PodRunning), Age: 2, Restarts: 0},
		{Name: "pod 2", State: string(k8sv1.PodRunning), Age: 5, Restarts: 1},
	}
	return pl, nil
}

func (*fakeK8sOperations) PodLogs(namespace, podName string, opts *LogOptions) (io.ReadCloser, error) {
	r := bytes.NewBufferString("foo\nbar")
	return ioutil.NopCloser(r), nil
}

func (f *fakeK8sOperations) NamespaceAnnotation(namespace, annotation string) (string, error) {
	dpt := f.DefaultProcessType
	if dpt == "" {
		dpt = "web"
	}
	tmpl := `{
		"name": "test",
		"processType": "%s",
		"internal": %t,
		"virtualHost": "%s",
		"envVars": [{"key": "ENV-KEY", "value": "ENV-VALUE"}],
		"secrets": ["SECRET-1", "SECRET-2"],
		"secret_files": ["SECRET-3"],
		"protocol": "%s"
	}`
	return fmt.Sprintf(
		tmpl,
		dpt,
		f.AppInternal,
		f.AppVirtualHost,
		f.AppProtocol,
	), nil
}

func (*fakeK8sOperations) NamespaceLabel(namespace, label string) (string, error) {
	return "luizalabs", nil
}

func (f *fakeK8sOperations) CreateOrUpdateAutoscale(app *App) error {
	f.CreateOrUpdateAutoscaleWasCalled = true
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
			{Name: "pod 1", State: string(k8sv1.PodRunning), Age: 1, Restarts: 1},
			{Name: "pod 2", State: string(k8sv1.PodPending), Age: 2, Restarts: 2},
			{Name: "pod 3", State: string(k8sv1.PodRunning), Age: 3, Restarts: 3},
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

func (*fakeK8sOperations) IsInvalid(err error) bool {
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

func (f *fakeK8sOperations) DeleteCronJobEnvVars(namespace, name string, evNames []string) error {
	f.DeleteCronJobEnvVarsWasCalled = true
	return nil
}

func (*fakeK8sOperations) CreateOrUpdateDeployEnvVars(namespace, name string, evs []*EnvVar) error {
	return nil
}

func (f *fakeK8sOperations) CreateOrUpdateCronJobEnvVars(namespace, name string, evs []*EnvVar) error {
	f.CreateOrUpdateCronJobEnvVarsWasCalled = true
	return nil
}

func (*fakeK8sOperations) GetSecret(namespace, secretName string) (map[string][]byte, error) {
	return make(map[string][]byte), nil
}

func (*fakeK8sOperations) CreateOrUpdateDeploySecretEnvVars(namespace, name, secretName string, secrets []string) error {
	return nil
}

func (f *fakeK8sOperations) CreateOrUpdateCronJobSecretEnvVars(namespace, name, secretName string, secrets []string) error {
	return nil
}

func (*fakeK8sOperations) DeploySetReplicas(namespace, name string, replicas int32) error {
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

func (f *fakeK8sOperations) DeletePod(namespace, podName string) error {
	return nil
}

func (f *fakeK8sOperations) HasIngress(namespace, name string) (bool, error) {
	return f.AppIngress, nil
}

func (f *fakeK8sOperations) IngressEnabled() bool {
	return f.IngressEnabledValue
}

func (f *fakeK8sOperations) CreateOrUpdateDeploySecretFile(namespace, deploy, fileName string) error {
	return nil
}

func (f *fakeK8sOperations) CreateOrUpdateCronJobSecretFile(namespace, cronjob, fileName string) error {
	return nil
}

func (f *fakeK8sOperations) DeleteDeploySecrets(namespace, deploy string, envVars, volKeys []string) error {
	return nil
}

func (f *fakeK8sOperations) DeleteCronJobSecrets(namespace, cronjob string, envVars, volKeys []string) error {
	return nil
}

func (e *errK8sOperations) CreateNamespace(app *App, user string) error {
	return e.NamespaceErr
}

func (e *errK8sOperations) CreateQuota(app *App) error {
	return e.QuotaErr
}

func (e *errK8sOperations) CreateOrUpdateSecret(appName, secretName string, data map[string][]byte) error {
	return e.SecretErr
}

func (e *errK8sOperations) CreateOrUpdateAutoscale(app *App) error {
	return e.AutoscaleErr
}

func (e *errK8sOperations) PodList(namespace string, opts *PodListOptions) ([]*Pod, error) {
	return nil, e.Err
}

func (e *errK8sOperations) PodLogs(namespace, podName string, opts *LogOptions) (io.ReadCloser, error) {
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

func (e *errK8sOperations) IsAlreadyExists(err error) bool {
	return !e.NegateIsAlreadyExists
}

func (*errK8sOperations) IsInvalid(err error) bool {
	return true
}

func (e *errK8sOperations) IsNotFound(err error) bool {
	return !e.NegateIsNotFound
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

func (e *errK8sOperations) DeleteCronJobEnvVars(namespace, name string, evNames []string) error {
	return e.Err
}

func (e *errK8sOperations) CreateOrUpdateDeployEnvVars(namespace, name string, evs []*EnvVar) error {
	return e.CreateOrUpdateDeployEnvVarsErr
}

func (e *errK8sOperations) CreateOrUpdateCronJobEnvVars(namespace, name string, evs []*EnvVar) error {
	return e.Err
}

func (e *errK8sOperations) GetSecret(namespace, secretName string) (map[string][]byte, error) {
	return nil, e.GetSecretErr
}

func (e *errK8sOperations) CreateOrUpdateDeploySecretEnvVars(namespace, name, secretName string, secrets []string) error {
	return e.CreateOrUpdateDeploySecretEnvVarsErr
}

func (e *errK8sOperations) CreateOrUpdateCronJobSecretEnvVars(namespace, name, secretName string, secrets []string) error {
	return e.Err
}

func (e *errK8sOperations) DeleteNamespace(namespace string) error {
	delete(e.Namespaces, namespace)
	return e.DeleteNamespaceErr
}

func (e *errK8sOperations) DeploySetReplicas(namespace, name string, replicas int32) error {
	return e.Err
}

func (e *errK8sOperations) NamespaceListByLabel(label, value string) ([]string, error) {
	return nil, e.Err
}

func (e *errK8sOperations) DeletePod(namespace, podName string) error {
	return e.DeletePodErr
}

func (e *errK8sOperations) HasIngress(namespace, name string) (bool, error) {
	return false, nil
}

func (e *errK8sOperations) IngressEnabled() bool {
	return false
}

func (e *errK8sOperations) CreateOrUpdateDeploySecretFile(namespace, deploy, fileName string) error {
	return nil
}

func (e *errK8sOperations) CreateOrUpdateCronJobSecretFile(namespace, cronjob, fileName string) error {
	return nil
}

func (e *errK8sOperations) DeleteDeploySecrets(namespace, deploy string, envVars, volKeys []string) error {
	return nil
}

func (e *errK8sOperations) DeleteCronJobSecrets(namespace, cronjob string, envVars, volKeys []string) error {
	return nil
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

func TestAppOperationsCreateCronDoesNotCreateHPA(t *testing.T) {
	validCronPt := fmt.Sprintf("%s-test", ProcessTypeCronPrefix)
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	fakeK8s := &fakeK8sOperations{}
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name, ProcessType: validCronPt}
	tops.(*team.FakeOperations).Storage[name] = &database.Team{
		Name:  name,
		Users: []database.User{*user},
	}

	if err := ops.Create(user, app); err != nil {
		t.Fatal("error creating app: ", err)
	}

	if fakeK8s.CreateOrUpdateAutoscaleWasCalled {
		t.Error("expected no hpa for crons, but was created")
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

func TestAppOperationsCreateErrInvalidName(t *testing.T) {
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
	ops.(*AppOperations).kops = &errK8sOperations{
		NamespaceErr:          ErrInvalidName,
		NegateIsAlreadyExists: true,
	}

	if err := ops.Create(user, app); err != ErrInvalidName {
		t.Errorf("expected %v got %v", ErrInvalidName, err)
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

func TestAppCreateErrMissingVirtualHost(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	teamName := "luizalabs"
	fakeK8s := &fakeK8sOperations{IngressEnabledValue: true}
	ops := NewOperations(tops, fakeK8s, fakeSt)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: teamName, ProcessType: ProcessTypeWeb}
	tops.(*team.FakeOperations).Storage[teamName] = &database.Team{
		Name:  teamName,
		Users: []database.User{*user},
	}

	if err := ops.Create(user, app); err != ErrMissingVirtualHost {
		t.Errorf("want %v; got %v", ErrMissingVirtualHost, err)
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
	opts := &LogOptions{Lines: 10, Follow: false}

	rc, err := ops.Logs(user, app.Name, opts)
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
	opts := &LogOptions{Lines: 10, Follow: false}

	if _, err := ops.Logs(user, "teresa", opts); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppOperationsLogsErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	opts := &LogOptions{Lines: 10, Follow: false}

	if _, err := ops.Logs(user, "teresa", opts); err != ErrNotFound {
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
	ops := NewOperations(tops, &fakeK8sOperations{AppProtocol: "test"}, nil)
	teamName := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{
		Name:        "teresa",
		Team:        teamName,
		EnvVars:     []*EnvVar{&EnvVar{Key: "ENV-KEY", Value: "ENV-VALUE"}},
		Secrets:     []string{"SECRET-1", "SECRET-2"},
		SecretFiles: []string{"SECRET-3"},
	}
	tops.(*team.FakeOperations).Storage[teamName] = &database.Team{
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

	if len(info.EnvVars) != 3 { // see fakeK8sOperations.NamespaceAnnotation
		t.Errorf("expected 3, got %d", len(info.EnvVars))
	}
	for i, ev := range app.EnvVars {
		if info.EnvVars[i].Key != ev.Key {
			t.Errorf("expected %s, got %s", ev.Key, info.EnvVars[i].Key)
		}
		if info.EnvVars[i].Value != ev.Value {
			t.Errorf("expected %s, got %s", ev.Value, info.EnvVars[i].Value)
		}
	}
	for i, s := range app.Secrets {
		idx := len(app.EnvVars) + i
		if info.EnvVars[idx].Key != s {
			t.Errorf("expected %s, got %s", s, info.EnvVars[idx].Key)
		}
		expected := "*****"
		if info.EnvVars[idx].Value != expected {
			t.Errorf("expected %s, got %s", expected, info.EnvVars[idx].Value)
		}
	}

	for i, s := range app.SecretFiles {
		expected := fmt.Sprintf("%s/%s", SecretPath, s)
		if actual := info.Volumes[i]; actual != expected {
			t.Errorf("expected %s, got %s", expected, actual)
		}
	}

	if info.Protocol != "test" {
		t.Errorf("got %s; want grpc", info.Protocol)
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

func TestAppOpsInfoInternalApp(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{AppInternal: true}, nil)
	teamName := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: teamName, Internal: true}
	tops.(*team.FakeOperations).Storage[teamName] = &database.Team{
		Name:  teamName,
		Users: []database.User{*user},
	}

	info, err := ops.Info(user, app.Name)
	if err != nil {
		t.Fatal(err)
	}

	hostname := info.Addresses[0].Hostname
	if hostname != "test.test" {
		t.Errorf("expected test.test, got %s", hostname)
	}
}

func TestAppOpsInfoIngress(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{AppVirtualHost: "test", AppIngress: true}, nil)
	teamName := "luizalabs"
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: teamName, Internal: true}
	tops.(*team.FakeOperations).Storage[teamName] = &database.Team{
		Name:  teamName,
		Users: []database.User{*user},
	}

	info, err := ops.Info(user, app.Name)
	if err != nil {
		t.Fatal(err)
	}

	hostname := info.Addresses[0].Hostname
	if hostname != "test" {
		t.Errorf("expected test, got %s", hostname)
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

func TestAppOperationsListByTeam(t *testing.T) {
	tops := team.NewFakeOperations()
	appName := "teresa"

	fk8s := &fakeK8sOperations{Namespaces: map[string]struct{}{appName: {}}}

	ops := NewOperations(tops, fk8s, nil)

	apps, err := ops.ListByTeam("gophers")
	if err != nil {
		t.Fatal("error getting app list:", err)
	}

	if len(apps) == 0 {
		t.Fatal("expected at least one app")
	}

	if apps[0] != appName {
		t.Errorf("expected %s, got %s", appName, apps[0])
	}
}

func TestAppOperationsSetEnv(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
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
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.SetEnv(user, app.Name, nil); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestAppOpsSetEnvErrInvalidEnvVarName(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, nil, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{
		CreateOrUpdateDeployEnvVarsErr: ErrInvalidEnvVarName,
		NegateIsNotFound:               true,
	}
	evs := []*EnvVar{{Key: "key", Value: "value"}}

	if err := ops.SetEnv(user, app.Name, evs); err != ErrInvalidEnvVarName {
		t.Errorf("expected %v, got %v", ErrInvalidEnvVarName, err)
	}
}

func TestAppOperationsSetEnvForACronJob(t *testing.T) {
	validCronPt := fmt.Sprintf("%s-test", ProcessTypeCronPrefix)
	tops := team.NewFakeOperations()
	fakeK8s := &fakeK8sOperations{DefaultProcessType: validCronPt}
	ops := NewOperations(tops, fakeK8s, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
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

	if !fakeK8s.CreateOrUpdateCronJobEnvVarsWasCalled {
		t.Error("expected create or update CRON JOB env vars was called, but dont")
	}
}

func TestAppOperationsUnsetEnv(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
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
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
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
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.UnsetEnv(user, app.Name, nil); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestAppOperationsUnsetEnvForACronJob(t *testing.T) {
	validCronPt := fmt.Sprintf("%s-test", ProcessTypeCronPrefix)
	tops := team.NewFakeOperations()
	fakeK8s := &fakeK8sOperations{DefaultProcessType: validCronPt}
	ops := NewOperations(tops, fakeK8s, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	evs := []string{"key1", "key2"}

	if err := ops.UnsetEnv(user, app.Name, evs); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !fakeK8s.DeleteCronJobEnvVarsWasCalled {
		t.Error("expected delete CRON JOB env vars was called, but dont")
	}
}

func TestAppOperationsSetSecret(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	secrets := []*EnvVar{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	if err := ops.SetSecret(user, app.Name, secrets); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsSetSecretErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetSecret(user, "teresa", nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsSetSecretErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetSecret(user, "teresa", nil); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsSetSecretErrInternalServerErrorOnSaveApp(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{SetNamespaceAnnotationsErr: errors.New("test")}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.SetSecret(user, app.Name, nil); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestAppOpsSetSecretErrInvalidEnvVarName(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, nil, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{
		CreateOrUpdateDeploySecretEnvVarsErr: ErrInvalidEnvVarName,
		GetSecretErr:                         ErrNotFound,
	}
	secrets := []*EnvVar{{Key: "key", Value: "value"}}

	if err := ops.SetSecret(user, app.Name, secrets); err != ErrInvalidSecretName {
		t.Errorf("expected %v, got %v", ErrInvalidEnvVarName, err)
	}
}

func TestAppOperationsSetSecretFile(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.SetSecretFile(user, app.Name, "test", nil); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsSetSecretFileErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetSecretFile(user, "teresa", "test", nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsSetSecretFileErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetSecretFile(user, "teresa", "test", nil); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsSetSecretFileErrInternalServerErrorOnSaveApp(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{SetNamespaceAnnotationsErr: errors.New("test")}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.SetSecretFile(user, app.Name, "test", nil); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestAppOperationsUnsetSecret(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{
		Name:        "teresa",
		Team:        "luizalabs",
		Secrets:     []string{"key1"},
		SecretFiles: []string{"key2"},
	}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	secrets := []string{"key1", "key2"}

	if err := ops.UnsetSecret(user, app.Name, secrets); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsUnsetSecretErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.UnsetSecret(user, "teresa", nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsUnsetSecretErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.UnsetSecret(user, "teresa", nil); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsUnsetSecretErrInternalServerErrorOnSaveApp(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{SetNamespaceAnnotationsErr: errors.New("test")}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.UnsetSecret(user, app.Name, nil); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}
}

func TestAppOperationsSetAutoscale(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	req := newAutoscaleRequest("teresa")
	as := newAutoscale(req)

	if err := ops.SetAutoscale(user, app.Name, as); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsSetAutoscaleInvalidActionForCronJob(t *testing.T) {
	validCronPt := fmt.Sprintf("%s-test", ProcessTypeCronPrefix)
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{DefaultProcessType: validCronPt}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	req := newAutoscaleRequest("teresa")
	as := newAutoscale(req)

	if err := ops.SetAutoscale(user, app.Name, as); err != ErrInvalidActionForCronJob {
		t.Errorf("expected ErrInvalidActionForCronJob, got %v", err)
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
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
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
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
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
	ops := NewOperations(team.NewFakeOperations(), &fakeK8sOperations{}, nil)
	app := &App{Name: "teresa", Team: "luizalabs"}

	if err := ops.ChangeTeam(app.Name, "gopher"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsChangeTeamErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{SetNamespaceLabelsErr: ErrNotFound}, nil)

	if err := ops.ChangeTeam("gophers", "teresa"); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", teresa_errors.Get(err))
	}
}

func TestAppOperationsSetReplicas(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.SetReplicas(user, app.Name, 1); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOperationsSetReplicasInvalidForCronJob(t *testing.T) {
	validCronPt := fmt.Sprintf("%s-test", ProcessTypeCronPrefix)
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{DefaultProcessType: validCronPt}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}

	if err := ops.SetReplicas(user, app.Name, 1); err != ErrInvalidActionForCronJob {
		t.Errorf("expected ErrInvalidActionForCronJob, got %v", err)
	}
}

func TestAppOperationsSetReplicasErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetReplicas(user, "", 1); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestAppOperationsSetReplicasErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}

	if err := ops.SetReplicas(user, "teresa", 1); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOpsDeletePodsSuccess(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	pods := []string{"pod1", "pod2"}

	if err := ops.DeletePods(user, app.Name, pods); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAppOpsDeletePodsErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	pods := []string{"pod1", "pod2"}

	if err := ops.DeletePods(user, "teresa", pods); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected %v, got %v", ErrNotFound, teresa_errors.Get(err))
	}
}

func TestAppOpsDeletePodsErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	pods := []string{"pod1", "pod2"}

	if err := ops.DeletePods(user, "teresa", pods); err != auth.ErrPermissionDenied {
		t.Errorf("expected %v, got %v", auth.ErrPermissionDenied, teresa_errors.Get(err))
	}
}

func TestAppOpsDeletePodsInternalServerError(t *testing.T) {
	tops := team.NewFakeOperations()
	kops := &errK8sOperations{DeletePodErr: errors.New("test"), NegateIsNotFound: true}
	ops := NewOperations(tops, kops, nil)
	user := &database.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: "luizalabs"}
	tops.(*team.FakeOperations).Storage[app.Team] = &database.Team{
		Name:  app.Team,
		Users: []database.User{*user},
	}
	pods := []string{"pod1", "pod2"}

	if err := ops.DeletePods(user, app.Name, pods); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected %v, got %v", teresa_errors.ErrInternalServerError, teresa_errors.Get(err))
	}
}
