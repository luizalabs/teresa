package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var viewConfigCmd = &cobra.Command{
	Use:   "view",
	Short: "displays the client config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		configYaml, err := viewConfigFileYaml(cfgFile)
		if err != nil {
			lgr.WithError(err).Error("Error reading the config file yaml")
			return newSysError("Error reading the config file")
		}

		fmt.Print(configYaml)
		return nil
	},
}

func viewConfigFileYaml(fileName string) (string, error) {
	conf, err := readOrCreateConfigFile(fileName)
	if err != nil {
		return "", err
	}

	contentYaml, errMarshal := marshalConfigFile(conf)
	if err != nil {
		return "", errMarshal
	}

	return string((*contentYaml)[:]), nil
}

func init() {
	configCmd.AddCommand(viewConfigCmd)

}
