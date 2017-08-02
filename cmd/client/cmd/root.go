package cmd

import (
	"fmt"
	"net"
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
		if isUsageError(err) {
			fmt.Printf("%s\n\n", err.Error())
			cmd.Usage()
			os.Exit(1)
		}

		if isCmdError(err) {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Hack to print a invalid command for root
		// Ex.: teresa notvalidcommand
		if !cmd.HasParent() {
			fmt.Println(err)
			os.Exit(1)
		}

		// writting errors to log for future troubleshooting
		log.WithField("command", cmd.CommandPath()).Error(err)

		// from here below, try to print some usefull information for the user...
		// check if the error is a net error
		if _, ok := err.(net.Error); ok {
			fmt.Println("Failed to connect to server, or server is down!!!")
			os.Exit(1)
		}
		os.Exit(1)
	}
}

func init() {
	// init log here
	initLog()
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

func initLog() {
	log = logrus.New()
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
	viper.SetDefault("debug", debugFlag)
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

// FIXME: from here below, delete all?!?!?!

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

type defaultClientError interface {
	Code() int
	Error() string
}

func isErrorCode(err error, code int) bool {
	if tErr, ok := err.(defaultClientError); ok && tErr.Code() == code {
		return true
	}
	return false
}
func isBadRequest(err error) bool {
	return isErrorCode(err, 400)
}
func isUnauthorized(err error) bool {
	return isErrorCode(err, 401)
}
func isNotFound(err error) bool {
	return isErrorCode(err, 404)
}
func isConflicted(err error) bool {
	return isErrorCode(err, 409)
}
func isUnprocessableEntity(err error) bool {
	return isErrorCode(err, 422)
}

type usageError struct {
	msg string
}

func (e usageError) Error() string { return e.msg }

func newUsageError(msg string) error {
	return &usageError{msg}
}

func isUsageError(err error) bool {
	if _, ok := err.(*usageError); ok {
		return true
	}
	return false
}

type cmdError struct {
	msg      string
	sysError bool
}

func newCmdError(msg string) error {
	return &cmdError{msg, false}
}
func newCmdErrorf(format string, a ...interface{}) error {
	return &cmdError{fmt.Sprintf(format, a...), false}
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
