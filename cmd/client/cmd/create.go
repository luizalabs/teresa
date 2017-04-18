package cmd

import "github.com/spf13/cobra"

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create team, user or application",
	Long:  `Create team, user or application.`,
}

func init() {
	RootCmd.AddCommand(createCmd)
}
