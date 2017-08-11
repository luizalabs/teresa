package cmd

import (
	"fmt"

	"github.com/luizalabs/teresa/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the server version information",
	Run:   showVersion,
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Version: %s\n", version.Version)
}
