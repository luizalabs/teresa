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

var lgr *logrus.Logger

// variables used to capture use flags
var (
	cfgFile      string
	serverFlag   string
	currentFlag  bool
	userNameFlag string
)

const (
	version = "0.1.0"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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
	lgr = logrus.New()

	// lgr.Formatter = new(logrus.JSONFormatter)
	lgr.Formatter = new(prefixed.TextFormatter)

	lgr.Out = os.Stdout
	lgr.Level = logrus.WarnLevel
}

// from https://github.com/spf13/viper
func getUserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
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
		cfgFile = filepath.Join(getUserHomeDir(), ".paas_labs", "config.yaml")
	}
	viper.SetConfigFile(cfgFile)

	// defaults
	viper.SetDefault("debug", false)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		if cfgFileProvided {
			fmt.Println("Config file provided not found or with error")
		}
		lgr.WithFields(logrus.Fields{"cfgFile": cfgFile, "cfgFileProvided": cfgFileProvided, "error": err}).Fatalf("Error with the config file.")
	}

	if viper.GetBool("debug") {
		lgr.Level = logrus.DebugLevel
	}

	lgr.Debugf("Config settings %+v", viper.AllSettings())
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
