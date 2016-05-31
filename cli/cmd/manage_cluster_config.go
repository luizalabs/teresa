package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var setClusterCmd = &cobra.Command{
	Use:   "set-cluster NAME",
	Short: "sets a cluster entry in the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			lgr.Debug("Cluster name not provided")
			return newInputError("Cluster name must be provided")
		}

		if serverFlag == "" {
			lgr.Debug("Server not provided")
			return newInputError("Server not provided")
		}

		return setCluster(args[0], serverFlag, cfgFile)
	},
}

var useClusterCmd = &cobra.Command{
	Use:   "use-cluster NAME",
	Short: "sets a cluster as the current in the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			lgr.Debug("Cluster name not provided")
			return newInputError("Cluster name must be provided")
		}

		return setCurrentCluster(args[0], cfgFile)
	},
}

// add a new server to the config file
func setCluster(name string, server string, fileName string) error {
	if name == "" || server == "" || fileName == "" {
		return errors.New("Name, server and filename must be provided")
	}

	conf, err := readOrCreateConfigFile(fileName)
	if err != nil {
		return err
	}

	conf.Clusters[name] = clusterConfig{Server: server}

	if currentFlag {
		conf.CurrentCluster = name
	}

	if err := writeConfigFile(fileName, conf); err != nil {
		return err
	}

	return nil
}

func setCurrentCluster(name string, fileName string) error {
	if name == "" || fileName == "" {
		return errors.New("Name and filename must be provided")
	}

	conf, err := readOrCreateConfigFile(fileName)
	if err != nil {
		return err
	}

	if len(conf.Clusters) == 0 {
		return newSysError("The is no cluster configured yet.")
	}
	if _, exists := conf.Clusters[name]; !exists {
		return newSysError(fmt.Sprintf(`Cluster "%s" not configured yet`, name))
	}

	conf.CurrentCluster = name

	if err := writeConfigFile(fileName, conf); err != nil {
		return err
	}

	lgr.WithField("clusterName", name).Debug("New cluster set as current")

	return nil
}

func init() {
	setClusterCmd.Flags().StringVarP(&serverFlag, "server", "s", "", "URI of the server")
	setClusterCmd.Flags().BoolVarP(&currentFlag, "default", "d", false, "Is the default server")

	configCmd.AddCommand(setClusterCmd)
	configCmd.AddCommand(useClusterCmd)
}
