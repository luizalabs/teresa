package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modifies the config file",
}

func init() {
	RootCmd.AddCommand(configCmd)
}

type clusterConfig struct {
	Server     string `yaml:"server"`
	LoginToken string `yaml:"token"`
}

type configFile struct {
	Version        string                   `yaml:"version"`
	Clusters       map[string]clusterConfig `yaml:"clusters"`
	CurrentCluster string                   `yaml:"current_cluster"`
}

func readConfigFile(fileName string) (*configFile, error) {
	yamlFileContent, errRead := ioutil.ReadFile(cfgFile)

	if errRead != nil {
		lgr.WithError(errRead).Debug("Error loading the config file")
		return nil, errRead
	}

	conf := configFile{}

	if err := yaml.Unmarshal(yamlFileContent, &conf); err != nil {
		lgr.WithError(err).WithField("cfgFile", cfgFile).Error("Error trying to unmarshal the config file")
		return nil, err
	}
	return &conf, nil
}

func readOrCreateConfigFile(fileName string) (*configFile, error) {
	conf, err := readConfigFile(cfgFile)
	if err == nil {
		return conf, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	lgr.Debug("Config file not found... creating the base one")

	// set defaults
	conf = &configFile{Version: version}
	if conf.Clusters == nil {
		conf.Clusters = make(map[string]clusterConfig)
	}

	return conf, nil
}

func marshalConfigFile(config *configFile) (*[]byte, error) {
	d, err := yaml.Marshal(&config)
	if err != nil {
		lgr.WithError(err).WithField("config", config).Error("Error marshaling the config file")
		return nil, err
	}

	return &d, nil
}

func writeConfigFile(fileName string, config *configFile) error {
	// TODO: implement validate before writing
	lgr.WithField("fileName", fileName).WithField("config", *config).Debug("Marshaling the config file to save")
	d, err := marshalConfigFile(config)
	if err != nil {
		return err
	}

	// get config basepath
	p := filepath.Dir(fileName)
	if _, err := os.Stat(p); err != nil {
		if !os.IsNotExist(err) {
			lgr.WithError(err).WithField("directory", p).Error("Failed to check if the directory exists")
			return err
		}

		lgr.WithField("directory", p).Debug("Config basepath not found... creating")
		if errMkDir := os.MkdirAll(p, 0755); errMkDir != nil {
			lgr.WithError(errMkDir).WithField("directory", p).Error("Failed to create the directory")
			return errMkDir
		}
	}

	if errFile := ioutil.WriteFile(fileName, *d, 0600); errFile != nil {
		lgr.WithError(errFile).Error("Error while writing the config file to disk")
	}

	return nil
}

func getCurrentClusterName() (string, error) {
	current := viper.GetString("current_cluster")
	if current == "" {
		return "", newSysError("Not found a cluster to use")
	}
	return current, nil
}

func getCurrentServerBasePath() (string, error) {
	current, err := getCurrentClusterName()
	if err != nil {
		return "", err
	}

	return viper.GetString(fmt.Sprintf("clusters.%s.server", current)), nil
}
