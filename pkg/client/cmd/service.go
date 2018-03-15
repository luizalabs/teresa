package cmd

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/luizalabs/teresa/pkg/client"
	"github.com/luizalabs/teresa/pkg/client/connection"
	svcpb "github.com/luizalabs/teresa/pkg/protobuf/service"
	"github.com/spf13/cobra"

	"golang.org/x/net/context"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Everything about services",
}

var serviceEnableSSLCmd = &cobra.Command{
	Use:   "enable-ssl",
	Short: "Enable ssl for the app",
	Long: `Enable SSL for the app.

  To enable ssl for the app on aws:

  $ teresa service enable-ssl --app myapp --cert arn:aws:iam::xxxxx:server-certificate/cert-name

  To only support ssl:

  $ teresa service enable-ssl --app myapp --cert arn:aws:iam::xxxxx:server-certificate/cert-name --only`,
	Run: serviceEnableSSL,
}

var serviceInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show service info",
	Long:  `Show service info such as ports and ssl configuration.`,
	Run:   serviceInfo,
}

func serviceEnableSSL(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		cmd.Usage()
		return
	}

	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
	}

	cert, err := cmd.Flags().GetString("cert")
	if err != nil || cert == "" {
		client.PrintErrorAndExit("Invalid cert parameter")
	}

	only, err := cmd.Flags().GetBool("only")
	if err != nil {
		client.PrintErrorAndExit("Invalid only parameter")
	}

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %s", err)
	}
	defer conn.Close()

	cli := svcpb.NewServiceClient(conn)
	req := &svcpb.EnableSSLRequest{
		AppName: appName,
		Cert:    cert,
		Only:    only,
	}
	if _, err := cli.EnableSSL(context.Background(), req); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	fmt.Println("SSL enabled with success")
}

func serviceInfo(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}
	appName := args[0]

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := svcpb.NewServiceClient(conn)
	info, err := cli.Info(context.Background(), &svcpb.InfoRequest{AppName: appName})
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	color.New(color.FgCyan, color.Bold).Printf("[%s]\n", appName)
	bold := color.New(color.Bold).SprintFunc()

	if len(info.ServicePorts) > 0 {
		fmt.Println(bold("ports:"))
		for _, port := range info.ServicePorts {
			fmt.Printf("  %d\n", port.Port)
		}
	}
	if info.Ssl != nil {
		fmt.Println(bold("ssl:"))
		cert := info.Ssl.Cert
		if cert == "" {
			cert = "n/a"
		}
		fmt.Printf("  cert: %s\n", cert)
		if info.Ssl.ServicePort != nil {
			port := strconv.Itoa(int(info.Ssl.ServicePort.Port))
			if port == "0" {
				port = "n/a"
			}
			fmt.Printf("  port: %s\n", port)
		}
	}
}

func init() {
	RootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceEnableSSLCmd)
	serviceCmd.AddCommand(serviceInfoCmd)

	serviceEnableSSLCmd.Flags().String("app", "", "app name")
	serviceEnableSSLCmd.Flags().String("cert", "", "certificate identifier")
	serviceEnableSSLCmd.Flags().Bool("only", false, "only use SSL")
}
