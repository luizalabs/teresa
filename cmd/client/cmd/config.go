package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/luizalabs/teresa-api/pkg/client"
	"github.com/spf13/cobra"
)

var (
	serverFlag  string
	currentFlag bool
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Setup cluster servers and view config file",
	Long: `Setup clusters and view the configuration.

To perform any action, you must have at least one cluster setup.
To add one named "aws-staging", for instance:

	$ teresa config set-cluster aws-staging -s http://mycluster.mydomain.com

That will add the "aws-staging" cluster to the configuration file,
but won't set it as the default. To do that, you must run:

	$ teresa config use-cluster aws-staging

From that point on, teresa will use this cluster until you select
another via: teresa config use-cluster another-cluster.
	`,
}

// returns the config file
var viewConfigCmd = &cobra.Command{
	Use:   "view",
	Short: "view the config file",
	Long: `View the config file.

eg.:

	$ teresa config view
	`,
	Run: viewConfigFile,
}

var setClusterCmd = &cobra.Command{
	Use:   "set-cluster name",
	Short: "sets a cluster entry in the config file",
	Long: `Add or update a cluster entry.

eg.:

	$ teresa config set-cluster aws_staging --server https://staging.mydomain.com
	`,
	Run: setCluster,
}

var useClusterCmd = &cobra.Command{
	Use:   "use-cluster name",
	Short: "sets a cluster as the current in the config file",
	Long: `Set a cluster as in-use, so every action will be sent to it.

eg.:

	$ teresa config use-cluster aws_staging
	`,
	Run: useCluster,
}

func init() {
	RootCmd.AddCommand(configCmd)

	configCmd.AddCommand(viewConfigCmd)

	setClusterCmd.Flags().StringVarP(&serverFlag, "server", "s", "", "URI of the server")
	setClusterCmd.Flags().BoolVar(&currentFlag, "current", false, "Set this server to future use")
	configCmd.AddCommand(setClusterCmd)

	configCmd.AddCommand(useClusterCmd)
}

func useCluster(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		return
	}
	name := args[0]
	c, err := client.ReadConfigFile(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Cannot read the config file, have you created it with `teresa set-cluster` command?")
	}
	if _, ok := c.Clusters[name]; !ok {
		client.PrintErrorAndExit("Cluster `%s` not configured yet", name)
	}
	c.CurrentCluster = name
	if err = client.SaveConfigFile(cfgFile, c); err != nil {
		client.PrintErrorAndExit("Erro trying to save config file: %v", err)
	}
}

// setCluster add a new server to the config file
func setCluster(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		return
	}
	if serverFlag == "" {
		client.PrintErrorAndExit("Server URI not provided")
	}
	name := args[0]

	c, err := client.ReadConfigFile(cfgFile)
	if err != nil {
		c = &client.Config{
			Clusters:       make(map[string]client.ClusterConfig),
			CurrentCluster: name,
		}
	}

	c.Clusters[name] = client.ClusterConfig{Server: serverFlag}
	if currentFlag {
		c.CurrentCluster = name
	}

	if err = client.SaveConfigFile(cfgFile, c); err != nil {
		client.PrintErrorAndExit("Erro trying to save config file: %v", err)
	}
}

func viewConfigFile(cmd *cobra.Command, args []string) {
	y, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			client.PrintErrorAndExit(
				"Config file not found on `%s` use command `teresa config set-cluster` to create the config file",
				cfgFile,
			)
		} else {
			client.PrintErrorAndExit("Error trying to read config file: %v", err)
		}
	}
	fmt.Println(string(y))
}
