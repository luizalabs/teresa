package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTeresaYamlFromDeployTarBall(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "teresaYaml.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test", "test")
	if err != nil {
		t.Fatal("error getting deploy config file from tarball:", err)
	}

	if deployConfig.TeresaYaml == nil {
		t.Fatal("expected a valid TeresaYaml struct, got nil")
	}
	expectedText := "/healthcheck/"
	actual := deployConfig.TeresaYaml.HealthCheck.Liveness.Path
	if actual != expectedText {
		t.Errorf("expected %s, got %s", expectedText, actual)
	}
}

func TestGetTeresaYamlFromDeployTarBallConfigFiles(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "fooTxt.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test", "test")
	if err != nil {
		t.Fatal("error getting deploy config file from tarball:", err)
	}
	if deployConfig.TeresaYaml != nil {
		t.Errorf("expected nil, got %v", deployConfig.TeresaYaml)
	}
	if deployConfig.Procfile != nil {
		t.Errorf("expected nil, got %v", deployConfig.Procfile)
	}
}

func TestGetTeresaYamlFromDeployTarBallInvalidYaml(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "teresaYamlInvalid.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	if _, err := getDeployConfigFilesFromTarBall(tarBall, "test", "test"); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetProcfileFromDeployTarBall(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "procfile.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test", "test")
	if err != nil {
		t.Fatal("error getting deploy config file from tarball:", err)
	}

	if deployConfig.Procfile == nil {
		t.Fatal("expected a valid Procfile map, got nil")
	}
	expectedText := "python manage.py migrate"
	actual := deployConfig.Procfile["release"]
	if actual != expectedText {
		t.Errorf("expected %s, got %s", expectedText, actual)
	}
}

func TestGetDeployConfigFromTarBall(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "deployConfig.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test", "test")
	if err != nil {
		t.Fatal("error getting deploy config file from tarball:", err)
	}

	if deployConfig.Procfile == nil {
		t.Fatal("expected a valid Procfile map, got nil")
	}
	expectedText := "python manage.py migrate"
	actual := deployConfig.Procfile["release"]
	if actual != expectedText {
		t.Errorf("expected %s, got %s", expectedText, actual)
	}

	if deployConfig.TeresaYaml == nil {
		t.Fatal("expected a valid TeresaYaml struct, got nil")
	}
	expectedText = "/healthcheck/"
	actual = deployConfig.TeresaYaml.HealthCheck.Liveness.Path
	if actual != expectedText {
		t.Errorf("expected %s, got %s", expectedText, actual)
	}
}

func TestGetTeresaYamlForProcessTypeFromDeployTarBall(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "teresaYamlTestProcessType.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test", "test")
	if err != nil {
		t.Fatal("error getting deploy config file from tarball:", err)
	}

	if deployConfig.TeresaYaml == nil {
		t.Fatal("expected a valid TeresaYaml struct, got nil")
	}

	expectedText := "/test-healthcheck/"
	actual := deployConfig.TeresaYaml.HealthCheck.Liveness.Path
	if actual != expectedText {
		t.Errorf("expected %s, got %s", expectedText, actual)
	}
}

func TestGetTeresaYamlV2OK(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "teresaYamlV2.tgz"))
	if err != nil {
		t.Fatal("failed to open the tarball:", err)
	}
	defer tarBall.Close()
	var testCases = []struct {
		appName      string
		livenessPath string
		drainTimeout int
	}{
		{"test1", "/healthcheck1/", 1},
		{"test2", "/healthcheck2/", 2},
	}

	for _, tc := range testCases {
		tarBall.Seek(0, 0)
		deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, tc.appName, "test")
		if err != nil {
			t.Fatal("error getting deploy config file from tarball:", err)
		}
		if deployConfig.TeresaYaml == nil {
			t.Fatal("want a valid TeresaYaml struct; got nil")
		}
		path := deployConfig.TeresaYaml.HealthCheck.Liveness.Path
		if path != tc.livenessPath {
			t.Errorf("got %s; want %s", path, tc.livenessPath)
		}
		timeout := deployConfig.TeresaYaml.Lifecycle.PreStop.DrainTimeoutSeconds
		if timeout != tc.drainTimeout {
			t.Errorf("got %d; want %d", timeout, tc.drainTimeout)
		}
	}
}

func TestGetTeresaYamlV2MissingApp(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "teresaYamlV2.tgz"))
	if err != nil {
		t.Fatal("failed to open the tarball:", err)
	}
	defer tarBall.Close()
	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "missing", "test")
	if err != nil {
		t.Fatal("error getting deploy config file from tarball:", err)
	}
	if deployConfig.TeresaYaml == nil {
		t.Fatal("want a valid TeresaYaml struct; got nil")
	}

	hc := deployConfig.TeresaYaml.HealthCheck
	if hc != nil {
		t.Errorf("got %v; want nil", hc)
	}
	lc := deployConfig.TeresaYaml.Lifecycle
	if lc != nil {
		t.Errorf("got %v; want nil", lc)
	}
	ru := deployConfig.TeresaYaml.RollingUpdate
	if ru != nil {
		t.Errorf("got %v; want nil", ru)
	}
}
