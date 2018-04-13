package deploy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/build"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/luizalabs/teresa/pkg/server/test"
)

type fakeK8sOperations struct {
	lastDeploySpec                *spec.Deploy
	lastCronJobSpec               *spec.CronJob
	createDeployReturn            error
	createCronJobReturn           error
	hasSrvErr                     error
	exposeDeployWasCalled         bool
	replicaSetListByLabelErr      error
	createConfigMapWasCalled      bool
	deleteConfigMapWasCalled      bool
	containerExplicitEnvVarsErr   error
	containerExplicitEnvVarsValue []*app.EnvVar
}

func (f *fakeK8sOperations) CreateOrUpdateConfigMap(namespace, name string, data map[string]string) error {
	f.createConfigMapWasCalled = true
	return nil
}

func (f *fakeK8sOperations) DeleteConfigMap(namespace, name string) error {
	f.deleteConfigMapWasCalled = true
	return nil
}

func (f *fakeK8sOperations) IsNotFound(err error) bool {
	return true
}

func (f *fakeK8sOperations) CreateOrUpdateDeploy(deploySpec *spec.Deploy) error {
	f.lastDeploySpec = deploySpec
	return f.createDeployReturn
}

func (f *fakeK8sOperations) CreateOrUpdateCronJob(cronJobSpec *spec.CronJob) error {
	f.lastCronJobSpec = cronJobSpec
	return f.createCronJobReturn
}

func (f *fakeK8sOperations) ExposeDeploy(namespace, name, vHost, svcType, portName string, w io.Writer) error {
	f.exposeDeployWasCalled = true
	return nil
}

func (f *fakeK8sOperations) ReplicaSetListByLabel(namespace, label, value string) ([]*ReplicaSetListItem, error) {
	items := []*ReplicaSetListItem{
		{
			Revision:    "1",
			Description: "Test 1",
			CreatedAt:   "1",
			Current:     false,
		},
		{
			Revision:    "2",
			Description: "Test 2",
			CreatedAt:   "2",
			Current:     true,
		},
	}
	return items, f.replicaSetListByLabelErr
}

func (f *fakeK8sOperations) DeployRollbackToRevision(namespace, name, revision string) error {
	return nil
}

func (f *fakeK8sOperations) ContainerExplicitEnvVars(namespace, deployName, containerName string) ([]*app.EnvVar, error) {
	return f.containerExplicitEnvVarsValue, f.containerExplicitEnvVarsErr
}

func (f *fakeK8sOperations) WatchDeploy(namespace, deployName string) error {
	return nil
}

func TestDeployPermissionDenied(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	u := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.Background()
	_, errChan := ops.Deploy(ctx, u, "teresa", &test.FakeReadSeeker{}, "test")

	if err := <-errChan; err != auth.ErrPermissionDenied {
		t.Errorf("expecter ErrPermissionDenied, got %v", err)
	}
}

func TestDeploy(t *testing.T) {
	// this is a dummy test to prevent panic errors
	podExitCodeChan := make(chan int, 1)
	defer close(podExitCodeChan)
	podExitCodeChan <- 0

	tarBall, err := os.Open(filepath.Join("testdata", "fooTxt.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	u := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.Background()
	r, errChan := ops.Deploy(ctx, u, "teresa", tarBall, "test")
	select {
	case err = <-errChan:
	default:
	}

	if err != nil {
		t.Fatal("error making deploy:", err)
	}
	defer r.Close()
}

func TestCreateDeploy(t *testing.T) {
	expectedName := "Test app"
	a := &app.App{Name: expectedName}
	expectedDescription := "test-description"
	expectedSlugURL := "test-slug"
	opts := &Options{RevisionHistoryLimit: 3}

	conf := &DeployConfigFiles{Procfile: map[string]string{"worker": "echo hello world"}}

	fakeK8s := new(fakeK8sOperations)
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		opts,
	)

	err := ops.(*DeployOperations).createOrUpdateDeploy(
		a,
		conf,
		new(bytes.Buffer),
		expectedSlugURL,
		expectedDescription,
		"123",
	)

	if err != nil {
		t.Fatal("error create deploy:", err)
	}

	if fakeK8s.lastDeploySpec.Name != expectedName {
		t.Errorf("expected %s, got %s", expectedName, fakeK8s.lastDeploySpec.Name)
	}
	if fakeK8s.lastDeploySpec.Description != expectedDescription {
		t.Errorf("expected %s, got %s", expectedDescription, fakeK8s.lastDeploySpec.Description)
	}
	if fakeK8s.lastDeploySpec.SlugURL != expectedSlugURL {
		t.Errorf("expected %s, got %s", expectedSlugURL, fakeK8s.lastDeploySpec.SlugURL)
	}
	if fakeK8s.lastDeploySpec.RevisionHistoryLimit != opts.RevisionHistoryLimit {
		t.Errorf("expected %d, got %d", opts.RevisionHistoryLimit, fakeK8s.lastDeploySpec.RevisionHistoryLimit)
	}
}

func TestCreateDeployCreateNginxConfigMap(t *testing.T) {
	conf := &DeployConfigFiles{NginxConf: "nginx conf"}

	fakeK8s := new(fakeK8sOperations)
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)

	err := ops.(*DeployOperations).createOrUpdateDeploy(
		&app.App{},
		conf,
		new(bytes.Buffer),
		"test-slug",
		"test-description",
		"123",
	)

	if err != nil {
		t.Fatal("error create deploy:", err)
	}

	if !fakeK8s.createConfigMapWasCalled {
		t.Error("expected create config map was called, but don't")
	}

}

func TestCreateDeployDeleteNginxConfigMap(t *testing.T) {
	conf := &DeployConfigFiles{}

	fakeK8s := new(fakeK8sOperations)
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)

	err := ops.(*DeployOperations).createOrUpdateDeploy(
		&app.App{},
		conf,
		new(bytes.Buffer),
		"test-slug",
		"test-description",
		"123",
	)

	if err != nil {
		t.Fatal("error create deploy:", err)
	}

	if !fakeK8s.deleteConfigMapWasCalled {
		t.Error("expected delete config map was called, but don't")
	}

}

func TestCreateDeployReturnError(t *testing.T) {
	expectedErr := errors.New("Some k8s error")
	fakeK8s := &fakeK8sOperations{createDeployReturn: expectedErr}

	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)

	err := ops.(*DeployOperations).createOrUpdateDeploy(
		&app.App{Name: "test"},
		&DeployConfigFiles{Procfile: map[string]string{}},
		new(bytes.Buffer),
		"some slug",
		"some desc",
		"123",
	)

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestCreateCronJob(t *testing.T) {
	validCronPt := fmt.Sprintf("%s-hw", app.ProcessTypeCronPrefix)
	expectedName := "Test cron"
	a := &app.App{Name: expectedName, ProcessType: validCronPt}
	expectedDescription := "test-description"
	expectedSlugURL := "test-slug"
	conf := &DeployConfigFiles{
		Procfile: map[string]string{validCronPt: "echo hello world"},
		TeresaYaml: &spec.TeresaYaml{
			Cron: &spec.CronArgs{Schedule: "*/1 * * * *"},
		},
	}

	fakeK8s := new(fakeK8sOperations)
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)

	deployOperations := ops.(*DeployOperations)
	err := deployOperations.createOrUpdateCronJob(a, conf, new(bytes.Buffer), expectedSlugURL, expectedDescription)

	if err != nil {
		t.Fatal("error create cronJob:", err)
	}

	if fakeK8s.lastCronJobSpec.Name != expectedName {
		t.Errorf("expected %s, got %s", expectedName, fakeK8s.lastCronJobSpec.Name)
	}
	if fakeK8s.lastCronJobSpec.SlugURL != expectedSlugURL {
		t.Errorf("expected %s, got %s", expectedSlugURL, fakeK8s.lastCronJobSpec.SlugURL)
	}
	if fakeK8s.lastCronJobSpec.Schedule != conf.TeresaYaml.Cron.Schedule {
		t.Errorf(
			"expected %s, got %s",
			conf.TeresaYaml.Cron.Schedule,
			fakeK8s.lastCronJobSpec.Schedule,
		)
	}
}

func TestCreateCronJobReturnError(t *testing.T) {
	expectedErr := errors.New("Some k8s error")
	fakeK8s := &fakeK8sOperations{createCronJobReturn: expectedErr}

	conf := &DeployConfigFiles{
		Procfile: map[string]string{"cron": "echo hello world"},
		TeresaYaml: &spec.TeresaYaml{
			Cron: &spec.CronArgs{Schedule: "*/1 * * * *"},
		},
	}

	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)

	deployOperations := ops.(*DeployOperations)
	err := deployOperations.createOrUpdateCronJob(
		&app.App{Name: "test"},
		conf,
		new(bytes.Buffer),
		"some slug",
		"some desc",
	)

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestCreateCronJobScheduleNotFound(t *testing.T) {
	validCronPt := fmt.Sprintf("%s-hw", app.ProcessTypeCronPrefix)
	a := &app.App{Name: "test", ProcessType: validCronPt}
	conf := &DeployConfigFiles{
		Procfile:   map[string]string{validCronPt: "echo hello world"},
		TeresaYaml: &spec.TeresaYaml{},
	}
	fakeK8s := new(fakeK8sOperations)
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)

	deployOperations := ops.(*DeployOperations)
	err := deployOperations.createOrUpdateCronJob(a, conf, new(bytes.Buffer), "test", "test")

	if err != ErrCronScheduleNotFound {
		t.Errorf("expected %v, got %v", ErrCronScheduleNotFound, err)
	}
}

func TestExposeApp(t *testing.T) {
	var testCases = []struct {
		appProcessType                string
		hasSrvErr                     error
		expectedExposeDeployWasCalled bool
	}{
		{app.ProcessTypeWeb, nil, true},
		{app.ProcessTypeWeb, nil, true},
		{app.ProcessTypeWeb, errors.New("some sad error"), true},
		{"worker", nil, false},
	}

	for _, tc := range testCases {
		fakeK8s := &fakeK8sOperations{
			hasSrvErr: tc.hasSrvErr,
		}
		ops := NewDeployOperations(
			app.NewFakeOperations(),
			fakeK8s,
			storage.NewFake(),
			exec.NewFakeOperations(),
			build.NewFakeOperations(),
			&Options{},
		)
		deployOperations := ops.(*DeployOperations)
		deployOperations.exposeApp(&app.App{ProcessType: tc.appProcessType}, new(bytes.Buffer))

		if fakeK8s.exposeDeployWasCalled != tc.expectedExposeDeployWasCalled {
			t.Errorf(
				"expected %v, got %v",
				tc.expectedExposeDeployWasCalled,
				fakeK8s.exposeDeployWasCalled,
			)
		}
	}
}

func TestRunReleaseCmd(t *testing.T) {
	var testCases = []struct {
		commandErr  error
		expectedErr error
	}{
		{nil, nil}, {exec.ErrNonZeroExitCode, ErrReleaseFail},
	}

	for _, tc := range testCases {
		fakeExec := exec.NewFakeOperations()
		fakeExec.ExpectedErr = tc.commandErr

		ops := NewDeployOperations(
			app.NewFakeOperations(),
			&fakeK8sOperations{},
			storage.NewFake(),
			fakeExec,
			build.NewFakeOperations(),
			&Options{},
		)

		deployOperations := ops.(*DeployOperations)
		err := deployOperations.runReleaseCmd(
			&app.App{Name: "Test"},
			"123456",
			"/slug.tgz",
			new(bytes.Buffer),
		)

		if err != tc.expectedErr {
			t.Errorf("expected %v, got %v", tc.expectedErr, err)
		}
	}
}

func TestDeployListSuccess(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	user := &database.User{Email: "gopher@luizalabs.com"}

	items, err := ops.List(user, "teresa")
	if err != nil {
		t.Fatal("got error listing replicasets: ", err)
	}

	// see fakeK8sOperations
	count := len(items)
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}

	item := items[1]
	if item.Revision != "2" {
		t.Errorf("expected '2', got %s", item.Revision)
	}

	if item.Description != "Test 2" {
		t.Errorf("expected 'Test 2', got %s", item.Description)
	}

	if item.CreatedAt != "2" {
		t.Errorf("expected 2, got %s", item.CreatedAt)
	}

	if !item.Current {
		t.Error("expected true, got false")
	}
}

func TestDeployListErrPermissionDenied(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	user := &database.User{Email: "bad-user@luizalabs.com"}

	if _, err := ops.List(user, "teresa"); err != auth.ErrPermissionDenied {
		t.Errorf("expected auth.ErrPermissionDenied, got %s", err)
	}
}

func TestDeployListErrNotFound(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	user := &database.User{Email: "gopher@luizalabs.com"}

	if _, err := ops.List(user, "app"); err != app.ErrNotFound {
		t.Errorf("expected app.ErrNotFound, got %s", err)
	}
}

func TestDeployListInternalServerError(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{replicaSetListByLabelErr: errors.New("test")},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	user := &database.User{Email: "gopher@luizalabs.com"}

	if _, err := ops.List(user, "teresa"); teresa_errors.Get(err) != teresa_errors.ErrInternalServerError {
		t.Errorf("expected ErrInternalServerError, got %s", err)
	}
}

func TestRollbackOpsSuccess(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	user := &database.User{Email: "gopher@luizalabs.com"}
	name := "teresa"

	if err := ops.Rollback(user, name, ""); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRollbackOpsErrPermissionDenied(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	user := &database.User{Email: "bad-user@luizalabs.com"}
	name := "teresa"

	if err := ops.Rollback(user, name, ""); err != auth.ErrPermissionDenied {
		t.Errorf("expected auth.ErrPermissionDenied, got %s", err)
	}
}

func TestRollbackOpsErrNotFound(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	user := &database.User{Email: "gopher@luizalabs.com"}
	name := "bad-app"

	if err := ops.Rollback(user, name, ""); err != app.ErrNotFound {
		t.Errorf("expected app.ErrNotFound, got %s", err)
	}
}

func TestRollbackErrorFromContainerEnvVars(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{containerExplicitEnvVarsErr: errors.New("test")},
		storage.NewFake(),
		exec.NewFakeOperations(),
		build.NewFakeOperations(),
		&Options{},
	)
	user := &database.User{Email: "gopher@luizalabs.com"}

	want := teresa_errors.ErrInternalServerError
	if err := ops.Rollback(user, "teresa", ""); teresa_errors.Get(err) != want {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), want)
	}
}

func TestIsProtectedEnvVar(t *testing.T) {
	var testCases = []struct {
		name string
		want bool
	}{
		{"TEST", false},
		{"APP", true},
		{"PORT", true},
	}

	for _, tc := range testCases {
		got := isProtectedEnvVar(tc.name)
		if got != tc.want {
			t.Errorf("got %v; want %v", got, tc.want)
		}
	}
}
