package cmd

import "github.com/spf13/cobra"

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete team or user",
	Long:  `Delete team or user.`,
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
