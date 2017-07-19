package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTeresaYamlFromTarBall(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "teresaYaml.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	tYaml, err := getTeresaYamlFromTarBall(tarBall)
	if err != nil {
		t.Fatal("error getting teresa yaml from tarball:", err)
	}

	expectedText := "/healthcheck/"
	actual := tYaml.HealthCheck.Liveness.Path
	if actual != expectedText {
		t.Errorf("expected %s, got %s", expectedText, actual)
	}
}

func TestGetTeresaYamlFromTarBallWithoutTheTersaYamlFile(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "fooTxt.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	tYaml, err := getTeresaYamlFromTarBall(tarBall)
	if err != nil {
		t.Fatal("error getting teresa yaml from tarball:", err)
	}
	if tYaml != nil {
		t.Errorf("expected nil, got %v", tYaml)
	}
}

func TestGetTeresaYamlFromTarBallInvalidYaml(t *testing.T) {
	tarBall, err := os.Open(filepath.Join("testdata", "teresaYamlInvalid.tgz"))
	if err != nil {
		t.Fatal("error getting tarBall:", err)
	}
	defer tarBall.Close()

	tYaml, err := getTeresaYamlFromTarBall(tarBall)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if tYaml != nil {
		t.Errorf("expected nil, got %v", tYaml)
	}
}
