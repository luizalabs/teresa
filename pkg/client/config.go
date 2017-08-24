package client

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"

	yaml "gopkg.in/yaml.v2"
)

type ClusterConfig struct {
	Server   string `yaml:"server"`
	Token    string `yaml:"token"`
	UseTLS   bool   `yaml:"tls"`
	Insecure bool   `yaml:"insecure"`
}

type Config struct {
	Clusters       map[string]ClusterConfig `yaml:"clusters"`
	CurrentCluster string                   `yaml:"current_cluster"`
}

var (
	DefaultConfigFileLocation string
	ErrInvalidConfigFile      = errors.New("Invalid config file")
	ErrInvalidDir             = errors.New("Invalid directory")
)

func init() {
	homeDir, err := homedir.Dir()
	if err != nil {
		// print a warning ?
		return
	}
	DefaultConfigFileLocation = filepath.Join(homeDir, ".teresa", "config.yaml")
}

func SaveToken(cfgFile, token string) error {
	cfg, err := ReadConfigFile(cfgFile)
	if err != nil {
		return err
	}
	cc, ok := cfg.Clusters[cfg.CurrentCluster]
	if !ok {
		return ErrInvalidConfigFile
	}

	cc.Token = token
	cfg.Clusters[cfg.CurrentCluster] = cc
	return SaveConfigFile(cfgFile, cfg)
}

func SaveConfigFile(path string, cfg *Config) error {
	b, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	stat, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	} else if !stat.IsDir() {
		return ErrInvalidDir
	}

	return ioutil.WriteFile(path, b, 0600)
}

func GetConfig(cfgFile, cfgCluster string) (*ClusterConfig, error) {
	cfg, err := ReadConfigFile(cfgFile)
	if err != nil {
		return nil, err
	}
	if cfgCluster != "" {
		cfg.CurrentCluster = cfgCluster
	}
	currentClusterConfig, ok := cfg.Clusters[cfg.CurrentCluster]
	if !ok {
		return nil, ErrInvalidConfigFile
	}
	return &currentClusterConfig, nil
}

func ReadConfigFile(cfgFile string) (*Config, error) {
	y, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, ErrInvalidConfigFile
	}

	c := new(Config)
	if err = yaml.Unmarshal(y, c); err != nil {
		return nil, ErrInvalidConfigFile
	}
	if c.Clusters == nil {
		c.Clusters = make(map[string]ClusterConfig)
	}
	return c, nil
}
