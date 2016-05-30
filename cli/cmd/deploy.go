package cmd

import "github.com/spf13/cobra"

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "View and modifies deployments of the apps",
}

func init() {
	RootCmd.AddCommand(deployCmd)
}
