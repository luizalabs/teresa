package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/luizalabs/teresa/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "teresa",
	Short: "Teresa",
	Long: `Teresa. You can manage teams, users and create and deploy applications with it.

Teresa works by sending HTTP requests to Kubernetes clusters that have a
Teresa server running. You can have multiple clusters configured on your
local box, one for each cloud provider or one for each environment or a mix
of those.

Teresa doesn't start using any cluster by it's own: you have to tell her which
one to use.

To set a cluster, eg.:

  $ teresa config set-cluster my_cluster_name --server mycluster.mydomain.com

You can also pass extra flags:

  --tls          Use TLS when communicating with the server
  --tlsinsecure  Don't check the server certificate with the default CA certificate bundle
  --current      Set this cluster as the current one
  --port         TCP port to use when communicating with the server

To use that cluster:

  $ teresa config use-cluster my_cluster_name

From that point on, all the operations will be directed to that cluster. You can
view the whole configuration anytime by running:

  $ teresa config view
	`,
}

func init() {
	// the config is only loaded if the command is valid,
	// that is why we use OnInitialize
	cobra.OnInitialize(initConfig)
	// using this so i will check manualy for strange behavior of the cli
	RootCmd.SilenceErrors = true
	RootCmd.SilenceUsage = true

	// change the suggestion distance of the commands
	RootCmd.SuggestionsMinimumDistance = 3
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	RootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "debug mode")
	RootCmd.PersistentFlags().MarkHidden("debug")
}

// from https://github.com/spf13/viper
func getUserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := fmt.Sprintf("%s%s", os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"))
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func initConfig() {
	// enable ability to specify config file via flag
	cfgFileProvided := false
	if cfgFile != "" {
		cfgFileProvided = true
		cfgFile = filepath.Clean(cfgFile)
	} else {
		cfgFile = filepath.Join(getUserHomeDir(), ".teresa", "config.yaml")
	}
	viper.SetConfigFile(cfgFile)
	// defaults
	viper.SetDefault("debug", debugFlag)
	// get from ENV
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		if cfgFileProvided {
			client.PrintErrorAndExit("Config file provided not found or with error")
		}
	}
}
