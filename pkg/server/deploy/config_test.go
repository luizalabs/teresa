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

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test")
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

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test")
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

	if _, err := getDeployConfigFilesFromTarBall(tarBall, "test"); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetProcfileFromDeployTarBall(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "procfile.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test")
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

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test")
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

	deployConfig, err := getDeployConfigFilesFromTarBall(tarBall, "test")
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
