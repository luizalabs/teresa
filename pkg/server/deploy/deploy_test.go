package deploy

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/luizalabs/teresa-api/pkg/server/app"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/database"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
)

type fakeReadSeeker struct{}

func (f *fakeReadSeeker) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (f *fakeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

type fakeK8sOperations struct {
	lastDeploySpec         *DeploySpec
	createDeployReturn     error
	hasSrvReturn           bool
	hasSrvErr              error
	createServiceWasCalled bool
	podRunReadCloser       io.ReadCloser
	podRunExitCodeChan     chan int
	podRunErr              error
}

func (f *fakeK8sOperations) PodRun(podSpec *PodSpec) (io.ReadCloser, <-chan int, error) {
	return f.podRunReadCloser, f.podRunExitCodeChan, f.podRunErr
}

func (f *fakeK8sOperations) CreateOrUpdateDeploy(deploySpec *DeploySpec) error {
	f.lastDeploySpec = deploySpec
	return f.createDeployReturn
}

func (f *fakeK8sOperations) HasService(namespace string, name string) (bool, error) {
	return f.hasSrvReturn, f.hasSrvErr
}

func (f *fakeK8sOperations) CreateService(namespace string, name string) error {
	f.createServiceWasCalled = true
	return nil
}

func TestDeployPermissionDenied(t *testing.T) {
	ops := NewDeployOperations(
		app.NewFakeOperations(),
		&fakeK8sOperations{},
		st.NewFake(),
	)
	u := &database.User{Email: "bad-user@luizalabs.com"}
	if _, err := ops.Deploy(u, "teresa", &fakeReadSeeker{}, "test", &Options{}); err != auth.ErrPermissionDenied {
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
	r, err := ops.Deploy(u, "teresa", tarBall, "test", &Options{})
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

func TestExposeService(t *testing.T) {
	var testCases = []struct {
		appProcessType                 string
		hasSrvReturn                   bool
		hasSrvErr                      error
		expectedCreateServiceWasCalled bool
	}{
		{app.ProcessTypeWeb, false, nil, true},
		{app.ProcessTypeWeb, true, nil, false},
		{app.ProcessTypeWeb, false, errors.New("some sad error"), false},
		{"worker", false, nil, false},
	}

	for _, tc := range testCases {
		fakeK8s := &fakeK8sOperations{
			hasSrvReturn: tc.hasSrvReturn,
			hasSrvErr:    tc.hasSrvErr,
		}
		ops := NewDeployOperations(
			app.NewFakeOperations(),
			fakeK8s,
			st.NewFake(),
		)
		deployOperations := ops.(*DeployOperations)
		err := deployOperations.exposeService(&app.App{ProcessType: tc.appProcessType}, new(bytes.Buffer))
		if err != tc.hasSrvErr {
			t.Error("error exposing service:", err)
		}

		if fakeK8s.createServiceWasCalled != tc.expectedCreateServiceWasCalled {
			t.Errorf(
				"expected %v, got %v",
				tc.expectedCreateServiceWasCalled,
				fakeK8s.createServiceWasCalled,
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
