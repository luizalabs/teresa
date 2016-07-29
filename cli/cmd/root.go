package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/x-cray/logrus-prefixed-formatter"
)

// log object to use over the cli
var log *logrus.Logger

// variables used to capture the cli flags
var (
	cfgFile            string
	serverFlag         string
	currentFlag        bool
	teamIDFlag         int64
	teamNameFlag       string
	teamEmailFlag      string
	teamURLFlag        string
	userIDFlag         int64
	userNameFlag       string
	userEmailFlag      string
	userPasswordFlag   string
	appNameFlag        string
	appScaleFlag       int
	descriptionFlag    string
	autocompleteTarget string
)

const (
	version           = "0.1.0"
	archiveTempFolder = "/tmp"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "teresa",
	Short: "Teresa cli",
	Long: `Teresa cli. You can manage teams, users and create and deploy applications with
it.

Teresa CLI works by sending HTTP requests to Kubernetes clusters that have a
Teresa API server running. You can have multiple clusters configured on your
local box, one for each cloud provider or one for each environment or a mix
of those.

Teresa doesn't start using any cluster by it's own: you have to tell her which
one to use.

To set a cluster, eg.:

  $ teresa config set-cluster my_cluster_name -s https://mycluster.mydomain.com

To use that cluster:

  $ teresa config use-cluster my_cluster_name

From that point on, all the operations will be directed to that cluster. You can
view the whole configuration anytime by running:

  $ teresa config view
	`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if cmd, err := RootCmd.ExecuteC(); err != nil {
		if isCmdError(err) {
			fmt.Printf("%s\n", err)
			if !isSysError(err) {
				fmt.Printf("\n%s", cmd.UsageString())
			}
			os.Exit(1)
		} else {
			// Dont log error because the logger is not ready yet
			// Print messagens like: unknown command "confi" for "cli"
			fmt.Println(err)
			os.Exit(-1)
		}
	}
}

func init() {
	cobra.OnInitialize(initLog, initConfig)
	// using this so i will check manualy for strange behavior of the cli
	RootCmd.SilenceErrors = true
	RootCmd.SilenceUsage = true
	// change the suggestion distance of the commands
	RootCmd.SuggestionsMinimumDistance = 3
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
}

func initLog() {
	// TODO: melhorar o log e enviar logs para o logentries
	log = logrus.New()
	// lgr.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(prefixed.TextFormatter)
	log.Out = os.Stdout
	log.Level = logrus.InfoLevel
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
	viper.SetDefault("debug", false)
	// get from ENV
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		if cfgFileProvided {
			fmt.Println("Config file provided not found or with error")
		}
		log.WithFields(logrus.Fields{"cfgFile": cfgFile, "cfgFileProvided": cfgFileProvided, "error": err}).Fatalf("Error with the config file.")
	}
	if viper.GetBool("debug") {
		log.Level = logrus.DebugLevel
	}
	log.Debugf("Config settings %+v", viper.AllSettings())
}

// Fatalf Prints formatted output, prepends the cli usage and exits
func Fatalf(cmd *cobra.Command, format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	s = fmt.Sprintf("%s\n%s\n\n%s", s, cmd.Long, cmd.UsageString())
	log.Fatalf(s)
}

// Usage Prints the cmd Long description and the usage string
func Usage(cmd *cobra.Command) {
	fmt.Printf("%s\n%s", cmd.Long, cmd.UsageString())
}

type cmdError struct {
	msg      string
	sysError bool
}

func (e cmdError) Error() string    { return e.msg }
func (e cmdError) isSysError() bool { return e.sysError }

func newInputError(msg string) error {
	return &cmdError{msg, false}
}
func newSysError(msg string) error {
	return &cmdError{msg, true}
}

func isCmdError(err error) bool {
	if _, ok := err.(*cmdError); ok {
		return true
	}
	return false
}

func isSysError(err error) bool {
	if cErr, ok := err.(*cmdError); ok && cErr.isSysError() {
		return true
	}
	return false
}
