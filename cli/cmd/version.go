package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the client and selected server version",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Client version: %s\n", version)

	},
}

func init() {
	RootCmd.AddCommand(versionCmd)

	versionCmd.Flags().Bool("only-client", false, "Show only the client info")
}
