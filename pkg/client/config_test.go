package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigFileForInvalidFiles(t *testing.T) {
	var testCases = []struct {
		file          string
		expectedError error
	}{
		{file: "doesNotExists.yaml", expectedError: ErrInvalidConfigFile},
		{file: "invalidYaml.yaml", expectedError: ErrInvalidConfigFile},
	}

	for _, tc := range testCases {
		_, err := ReadConfigFile(filepath.Join("testdata", tc.file))
		if err != ErrInvalidConfigFile {
			t.Errorf("(%s) expected ErrInvalidConfigFile, got %v", tc.file, err)
		}
	}
}

func TestReadConfigFileForBlankFile(t *testing.T) {
	c, err := ReadConfigFile(filepath.Join("testdata", "blankConfigFile.yaml"))
	if err != nil {
		t.Fatal("error on read a blank configuration file: ", err)
	}
	if c.Clusters == nil {
		t.Error("expected a empty map, but got nil")
	}
}

func TestReadConfigFileForValidFile(t *testing.T) {
	c, err := ReadConfigFile(filepath.Join("testdata", "validConfigFile.yaml"))
	if err != nil {
		t.Fatal("error on read a valid configuration file: ", err)
	}

	expectedCurrentCluster := "cluster-a"
	if c.CurrentCluster != expectedCurrentCluster {
		t.Errorf("expected %s, got %s", expectedCurrentCluster, c.CurrentCluster)
	}

	currentCluster, ok := c.Clusters[c.CurrentCluster]
	if !ok {
		t.Fatal("doesn't read clusters config")
	}

	expectedToken := "auth-token-cluster-a"
	if currentCluster.Token != expectedToken {
		t.Errorf("expected %s, got %s", expectedToken, currentCluster.Token)
	}
}

func TestGetConfigForInvalidFile(t *testing.T) {
	_, err := GetConfig(filepath.Join("testdata", "invalidConfigFile.yaml"), "")
	if err != ErrInvalidConfigFile {
		t.Error("expected ErrInvalidConfigFile, got ", err)
	}
}

func TestGetConfigForValidFiles(t *testing.T) {
	var testCases = []struct {
		file          string
		expectedToken string
	}{
		{file: "validConfigFile.yaml", expectedToken: "auth-token-cluster-a"},
		{file: "validConfigFile2.yaml", expectedToken: "auth-token-cluster-b"},
	}

	for _, tc := range testCases {
		c, err := GetConfig(filepath.Join("testdata", tc.file), "")
		if err != nil {
			t.Fatalf("error trying to get config of file %s: %s", tc.file, err)
		}
		if c.Token != tc.expectedToken {
			t.Errorf("expected %s, got %s", tc.expectedToken, c.Token)
		}
	}
}

func TestSaveConfigFile(t *testing.T) {
	var testCases = []struct {
		path         string
		pathToRemove string
		removeFunc   func(string) error
	}{
		{
			path:         filepath.Join("testdata", "temp.yaml"),
			pathToRemove: filepath.Join("testdata", "temp.yaml"),
			removeFunc:   os.Remove,
		}, {
			path:         filepath.Join("testdata", "some_dir", "temp.yaml"),
			pathToRemove: filepath.Join("testdata", "some_dir"),
			removeFunc:   os.RemoveAll,
		},
	}

	expectedConfig := &Config{
		CurrentCluster: "cluster-a",
		Clusters: map[string]ClusterConfig{
			"cluster-a": {Token: "token-a", Server: "http://teresa.com"},
		},
	}

	for _, tc := range testCases {
		if err := SaveConfigFile(tc.path, expectedConfig); err != nil {
			t.Fatal("error on save config file: ", err)
		}
		defer tc.removeFunc(tc.pathToRemove)

		c, err := GetConfig(tc.path, "")
		if err != nil {
			t.Fatalf("error trying to get config of file %s: %s", tc.path, err)
		}

		expectedToken := expectedConfig.Clusters[expectedConfig.CurrentCluster].Token
		if c.Token != expectedToken {
			t.Errorf("expected %s, got %s", expectedToken, c.Token)
		}
	}
}

func TestSaveTokenForInvalidConfigFile(t *testing.T) {
	conf := &Config{
		CurrentCluster: "cluster-x",
		Clusters: map[string]ClusterConfig{
			"cluster-a": {Token: "token-a", Server: "http://teresa.com"},
		},
	}

	confPath := filepath.Join("testdata", "temp.yaml")
	if err := SaveConfigFile(confPath, conf); err != nil {
		t.Fatal("error trying to save config file: ", err)
	}
	defer os.Remove(confPath)

	if err := SaveToken(confPath, "gopher"); err != ErrInvalidConfigFile {
		t.Errorf("expected ErrInvalidConfigFile, got %v", err)
	}
}

func TestSaveToken(t *testing.T) {
	conf := &Config{
		CurrentCluster: "cluster-a",
		Clusters: map[string]ClusterConfig{
			"cluster-a": {Token: "token-a", Server: "http://teresa.com"},
			"cluster-b": {Token: "token-b", Server: "http://teresa-b.com"},
		},
	}

	confPath := filepath.Join("testdata", "temp.yaml")
	if err := SaveConfigFile(confPath, conf); err != nil {
		t.Fatal("error trying to save config file: ", err)
	}
	defer os.Remove(confPath)

	expectedToken := "gopher"
	if err := SaveToken(confPath, expectedToken); err != nil {
		t.Fatal("error trying to save token: ", err)
	}

	c, err := GetConfig(confPath, "")
	if err != nil {
		t.Fatal("error trying to get config: ", err)
	}

	if c.Token != expectedToken {
		t.Errorf("expected %s, got %s", expectedToken, c.Token)
	}
}
