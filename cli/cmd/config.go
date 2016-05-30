package cmd

import (
	"errors"
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
	Server string `yaml:"server"`
	Token  string `yaml:"token"`
}

type configFile struct {
	Version        string                   `yaml:"version"`
	Clusters       map[string]clusterConfig `yaml:"clusters"`
	CurrentCluster string                   `yaml:"current_cluster"`
}

func readConfigFile(fileName string) (config *configFile, err error) {
	yamlFileContent, errRead := ioutil.ReadFile(cfgFile)
	if errRead != nil {
		if os.IsNotExist(errRead) {
			lgr.WithError(errRead).WithField("fileName", fileName).Debug("File not found when trying to read the config file")
		} else {
			lgr.WithError(errRead).Error("Error loading the config file")
		}
		return nil, errRead
	}

	conf := configFile{}

	if err := yaml.Unmarshal(yamlFileContent, &conf); err != nil {
		lgr.WithError(err).WithField("cfgFile", cfgFile).Error("Error trying to unmarshal the config file")
		return nil, err
	}
	return &conf, nil
}

func readOrCreateConfigFile(fileName string) (config *configFile, err error) {
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

func marshalConfigFile(config *configFile) (content *[]byte, err error) {
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
	d, errMarshal := marshalConfigFile(config)
	if errMarshal != nil {
		return errMarshal
	}

	// get config basepath
	p := filepath.Dir(fileName)
	dirStat, errStats := os.Stat(p)
	if errStats != nil {
		if !os.IsNotExist(errStats) {
			lgr.WithError(errStats).WithField("directory", p).Error("Failed to check if the directory exists")
			return errStats
		}

		lgr.WithField("directory", p).Debug("Config basepath not found... creating")
		if errMkDir := os.MkdirAll(p, 0755); errMkDir != nil {
			lgr.WithError(errMkDir).WithField("directory", p).Error("Failed to create the directory")
			return errMkDir
		}
	} else if !dirStat.IsDir() {
		lgr.WithField("directory", p).Error("Path exists, but isn't a directory")
		return errors.New("Path exists, but isn't a directory")
	}

	if errFile := ioutil.WriteFile(fileName, *d, 0600); errFile != nil {
		lgr.WithError(errFile).Error("Error while writing the config file to disk")
	}

	return nil
}

// TODO: ???? getCurrentServer ???

func getCurrentClusterName() (currentClusterName string, err error) {
	current := viper.GetString("current_cluster")
	if current == "" {
		lgr.Debug("Cluster not set yet")
		return "", newSysError("Set a cluster to use before continue")
	}
	return current, nil
}

func getCurrentCluster() (currentServer *clusterConfig, err error) {
	currentClusterName, errGetCurrent := getCurrentClusterName()
	if errGetCurrent != nil {
		return nil, errGetCurrent
	}

	var cluster clusterConfig

	if errUnmarshal := viper.UnmarshalKey(fmt.Sprintf("clusters.%s", currentClusterName), &cluster); errUnmarshal != nil {
		lgr.WithError(errUnmarshal).Error("Erro trying to unmarshal the current cluster")
		return nil, errors.New("Erro trying to unmarshal the current cluster")
	}

	return &cluster, nil
}
