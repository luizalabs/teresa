package cmd

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"

	"github.com/luizalabs/teresa/pkg/client"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Setup cluster servers and view config file",
	Long: `Setup clusters and view the configuration.

To perform any action, you must have at least one cluster setup.
To add one named "aws-staging", for instance:

  $ teresa config set-cluster aws-staging --server mycluster.mydomain.com

You can also pass extra flags:

  --port  TCP port to use when communicating with the server

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

  $ teresa config set-cluster aws_staging --server staging.mydomain.com
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

	setClusterCmd.Flags().String("server", "", "Server hostname or IP")
	setClusterCmd.Flags().Bool("tls", false, "Enables TLS")
	setClusterCmd.Flags().Bool("tlsinsecure", false, "Allow insecure TLS connections")
	setClusterCmd.Flags().Bool("current", false, "Set this server to future use")
	setClusterCmd.Flags().Int("port", 50051, "Server TCP port")
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
	server, _ := cmd.Flags().GetString("server")
	if server == "" {
		client.PrintErrorAndExit("Server URI not provided")
	}
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		client.PrintErrorAndExit("Invalid port parameter")
	}
	server = net.JoinHostPort(server, strconv.Itoa(port))

	useTLS, err := cmd.Flags().GetBool("tls")
	if err != nil {
		client.PrintErrorAndExit("Invalid tls parameter")
	}
	insecure, err := cmd.Flags().GetBool("tlsinsecure")
	if err != nil {
		client.PrintErrorAndExit("Invalid tlsinsecure parameter")
	}
	current, err := cmd.Flags().GetBool("current")
	if err != nil {
		client.PrintErrorAndExit("Invalid current parameter")
	}
	name := args[0]

	c, err := client.ReadConfigFile(cfgFile)
	if err != nil {
		c = &client.Config{
			Clusters:       make(map[string]client.ClusterConfig),
			CurrentCluster: name,
		}
	}

	c.Clusters[name] = client.ClusterConfig{
		Server:   server,
		UseTLS:   useTLS,
		Insecure: insecure,
	}
	if current {
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
