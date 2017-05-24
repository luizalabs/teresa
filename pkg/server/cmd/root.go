package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "teresa-server",
	Short: "teresa-server",
	Long:  "Teresa server, use to start Teresa gRPC server and create super user",
}
