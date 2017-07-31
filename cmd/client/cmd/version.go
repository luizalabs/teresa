package cmd

import (
	"fmt"

	"github.com/luizalabs/teresa-api/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the client version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
