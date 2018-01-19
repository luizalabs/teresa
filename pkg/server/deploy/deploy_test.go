package deploy

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	st "github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	context "golang.org/x/net/context"
)

type fakeReadSeeker struct{}

func (f *fakeReadSeeker) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (f *fakeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

type fakeK8sOperations struct {
	lastDeploySpec           *DeploySpec
	createDeployReturn       error
	hasSrvErr                error
	exposeDeployWasCalled    bool
	podRunReadCloser         io.ReadCloser
	podRunExitCodeChan       chan int
	podRunErr                error
	replicaSetListByLabelErr error
}

func (f *fakeK8sOperations) PodRun(podSpec *PodSpec) (io.ReadCloser, <-chan int, error) {
	return f.podRunReadCloser, f.podRunExitCodeChan, f.podRunErr
}

func (f *fakeK8sOperations) CreateOrUpdateDeploy(deploySpec *DeploySpec) error {
	f.lastDeploySpec = deploySpec
	return f.createDeployReturn
}

func (f *fakeK8sOperations) ExposeDeploy(namespace, name, vHost string, w io.Writer) error {
	f.exposeDeployWasCalled = true
	return nil
}

func (f *fakeK8sOperations) ReplicaSetListByLabel(namespace, label, value string) ([]*ReplicaSetListItem, error) {
	items := []*ReplicaSetListItem{
		{
			Revision:    "1",
			Description: "Test 1",
			Age:         1,
			Current:     false,
		},
		{
			Revision:    "2",
			Description: "Test 2",
			Age:         2,
			Current:     true,
		},
	}
	return items, f.replicaSetListByLabelErr
}

func (f *fakeK8sOperations) DeployRollbackToRevision(namespace, name, revision string) error {
	return nil
}

func (f *fakeK8sOperations) DeletePod(namespace, podName string) error {
	return nil
}

func TestDeployPermissionDenied(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		st.NewFake(),
	)
	u := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.Background()
	_, errChan := ops.Deploy(ctx, u, "teresa", &fakeReadSeeker{}, "test", &Options{})

	if err := <-errChan; err != auth.ErrPermissionDenied {
		t.Errorf("expecter ErrPermissionDenied, got %v", err)
	}
}

func TestDeploy(t *testing.T) {
	// this is a dummy test to prevent panic errors
	podExitCodeChan := make(chan int, 1)
	defer close(podExitCodeChan)
	fakeK8s := &fakeK8sOperations{
		podRunExitCodeChan: podExitCodeChan,
		podRunReadCloser:   ioutil.NopCloser(new(bytes.Buffer)),
	}
	podExitCodeChan <- 0

	tarBall, err := os.Open(filepath.Join("testdata", "fooTxt.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		st.NewFake(),
	)
	u := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.Background()
	r, errChan := ops.Deploy(ctx, u, "teresa", tarBall, "test", &Options{})
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

	fakeK8s := new(fakeK8sOperations)
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		st.NewFake(),
	)

	deployOperations := ops.(*DeployOperations)
	err := deployOperations.createDeploy(
		a,
		nil,
		expectedDescription,
		expectedSlugURL,
		opts,
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

func TestCreateDeployReturnError(t *testing.T) {
	expectedErr := errors.New("Some k8s error")
	fakeK8s := &fakeK8sOperations{createDeployReturn: expectedErr}

	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		st.NewFake(),
	)

	deployOperations := ops.(*DeployOperations)
	err := deployOperations.createDeploy(
		&app.App{Name: "test"},
		nil,
		"some desc",
		"some slug",
		&Options{},
	)

	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
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
			st.NewFake(),
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

func TestBuildApp(t *testing.T) {
	var testCases = []struct {
		exitCode    int
		expectedErr error
	}{
		{0, nil}, {1, ErrBuildFail},
	}

	podExitCodeChan := make(chan int, 1)
	defer close(podExitCodeChan)

	fakeK8s := &fakeK8sOperations{
		podRunExitCodeChan: podExitCodeChan,
		podRunReadCloser:   ioutil.NopCloser(new(bytes.Buffer)),
	}

	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		st.NewFake(),
	)

	for _, tc := range testCases {
		podExitCodeChan <- tc.exitCode
		deployOperations := ops.(*DeployOperations)
		err := deployOperations.buildApp(
			context.Background(),
			&fakeReadSeeker{},
			&app.App{Name: "Test"},
			"123456",
			"/slug.tgz",
			new(bytes.Buffer),
			&Options{},
		)

		if err != tc.expectedErr {
			t.Errorf("expected %v, got %v", tc.expectedErr, err)
		}
	}
}

func TestRunReleaseCmd(t *testing.T) {
	var testCases = []struct {
		exitCode    int
		expectedErr error
	}{
		{0, nil}, {1, ErrReleaseFail},
	}

	podExitCodeChan := make(chan int, 1)
	defer close(podExitCodeChan)

	fakeK8s := &fakeK8sOperations{
		podRunExitCodeChan: podExitCodeChan,
		podRunReadCloser:   ioutil.NopCloser(new(bytes.Buffer)),
	}

	ops := NewDeployOperations(
		app.NewFakeOperations(),
		fakeK8s,
		st.NewFake(),
	)

	for _, tc := range testCases {
		podExitCodeChan <- tc.exitCode
		deployOperations := ops.(*DeployOperations)
		err := deployOperations.runReleaseCmd(
			&app.App{Name: "Test"},
			"123456",
			"/slug.tgz",
			new(bytes.Buffer),
			&Options{},
		)

		if err != tc.expectedErr {
			t.Errorf("expected %v, got %v", tc.expectedErr, err)
		}
	}
}

func TestGenDeployId(t *testing.T) {
	generatedIds := make(map[string]bool)
	for i := 0; i < 10; i++ {
		gId := genDeployId()
		if _, found := generatedIds[gId]; found {
			t.Fatal("collision detected")
			generatedIds[gId] = true
		}
	}
}

func TestDeployListSuccess(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		st.NewFake(),
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

	if item.Age != 2 {
		t.Errorf("expected 2, got %d", item.Age)
	}

	if !item.Current {
		t.Error("expected true, got false")
	}
}

func TestDeployListErrPermissionDenied(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		st.NewFake(),
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
		st.NewFake(),
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
		st.NewFake(),
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
		st.NewFake(),
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
		st.NewFake(),
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
		st.NewFake(),
	)
	user := &database.User{Email: "gopher@luizalabs.com"}
	name := "bad-app"

	if err := ops.Rollback(user, name, ""); err != app.ErrNotFound {
		t.Errorf("expected app.ErrNotFound, got %s", err)
	}
}
