package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// returns the version of the cli and server
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the client and selected server version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Client version: %s\n", version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	// TODO: will be necessary in the future when we will return more info
	// versionCmd.Flags().Bool("only-client", false, "Show only the client info")
}
