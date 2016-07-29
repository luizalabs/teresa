package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// returns the config file
var viewConfigCmd = &cobra.Command{
	Use:   "view",
	Short: "view the config file",
	Long: `View the config file.

eg.:

	$ teresa config view
	`,
	Run: func(cmd *cobra.Command, args []string) {
		y, err := getConfigFileYaml(cfgFile)
		if err != nil {
			log.Fatalf("Failed to read config file: %s", err)
			log.Infof("") // to avoid GoImports importing "log"
		}
		fmt.Print(y)
	},
}

func init() {
	configCmd.AddCommand(viewConfigCmd)
}
