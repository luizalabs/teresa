package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var setClusterCmd = &cobra.Command{
	Use:   "set-cluster name",
	Short: "sets a cluster entry in the config file",
	Long: `Add or update a cluster entry.

eg.:

	$ teresa config set-cluster aws_staging --server https://staging.mydomain.com
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			Usage(cmd)
			return
		}
		name := args[0]
		if serverFlag == "" {
			Fatalf(cmd, "Server not provided")
		}
		if err := setCluster(name, serverFlag, currentFlag, cfgFile); err != nil {
			Fatalf(cmd, "%s", err)
		}
	},
}

var useClusterCmd = &cobra.Command{
	Use:   "use-cluster name",
	Short: "sets a cluster as the current in the config file",
	Long: `Set a cluster as in-use, so every action will be sent to it.

eg.:

	$ teresa config use-cluster aws_staging
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			Usage(cmd)
			return
		}
		name := args[0]
		setCurrentCluster(name, cfgFile)
	},
}

// add a new server to the config file
func setCluster(name string, server string, current bool, f string) error {
	if name == "" || server == "" || f == "" {
		return errors.New("Name, server and filename must be provided")
	}

	// try and parse the server url upfront
	if _, err := ParseServerURL(server); err != nil {
		return err
	}

	c, err := readOrCreateConfigFile(f)
	if err != nil {
		return err
	}

	c.Clusters[name] = clusterConfig{Server: server}
	// check and set this new cluster as the current one (default cluster)
	if current {
		c.CurrentCluster = name
	}
	return writeConfigFile(f, c)
}

func setCurrentCluster(name string, f string) error {
	if name == "" || f == "" {
		return errors.New("Name and filename must be provided")
	}
	c, err := readOrCreateConfigFile(f)
	if err != nil {
		return err
	}
	if len(c.Clusters) == 0 {
		return newSysError("There is no cluster configured yet.")
	}
	if _, e := c.Clusters[name]; !e {
		return newSysError(fmt.Sprintf(`Cluster "%s" not configured yet`, name))
	}
	// set the cluster as the current one
	c.CurrentCluster = name
	// write the config file
	if err := writeConfigFile(f, c); err != nil {
		return err
	}
	log.WithField("clusterName", name).Debug("New cluster set as current")
	return nil
}

func init() {
	setClusterCmd.Flags().StringVarP(&serverFlag, "server", "s", "", "URI of the server")
	setClusterCmd.Flags().BoolVar(&currentFlag, "current", false, "Set this server to future use")
	configCmd.AddCommand(setClusterCmd)
	configCmd.AddCommand(useClusterCmd)
}
