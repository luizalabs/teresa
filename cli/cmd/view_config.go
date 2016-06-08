package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// returns the config file
var viewConfigCmd = &cobra.Command{
	Use:   "view",
	Short: "displays the client config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		y, err := getConfigFileYaml(cfgFile)
		if err != nil {
			log.WithError(err).Error("Error reading the config file yaml")
			return newSysError("Error reading the config file")
		}
		fmt.Print(y)
		return nil
	},
}

func init() {
	configCmd.AddCommand(viewConfigCmd)
}
