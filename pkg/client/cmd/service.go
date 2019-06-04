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

  To enable ssl for the app on gcp (ingress only):

  $ teresa service enable-ssl --app myapp --cert cert-name

  To only support ssl:

  $ teresa service enable-ssl --app myapp --cert arn:aws:iam::xxxxx:server-certificate/cert-name --only`,
	Run: serviceEnableSSL,
}

var serviceSetStaticIpCmd = &cobra.Command{
	Use:   "set-static-ip",
	Short: "Set static IP for the app (GCP and ingress only)",
	Long: `Set static IP for the app (GCP and ingress only)

  To set static IP for the app on aws:

  $ teresa service set-static-ip --app myapp --address-name myapp-address`,
	Run: serviceSetStaticIp,
}

var serviceInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show service info",
	Long:  `Show service info such as ports and ssl configuration.`,
	Run:   serviceInfo,
}

var serviceWhitelistSourceRangesCmd = &cobra.Command{
	Use:   "whitelist-source-ranges <name> [source-range, ...]",
	Short: "Configure the cloud provider firewall whitelist",
	Long: `Configure the cloud provider firewall whitelist.

  To only let the ranges 200.234.1.0/24 and 200.234.2.0/24 access the app:

  $ teresa service whitelist-source-ranges myapp 200.234.1.0/24 200.234.2.0/24

  To remove the whitelist (no firewall):

  $ teresa service whitelist-source-ranges myapp`,
	Run: serviceWhitelistSourceRanges,
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
		client.PrintConnectionErrorAndExit(err)
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

func serviceSetStaticIp(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		cmd.Usage()
		return
	}

	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
	}

	addressName, err := cmd.Flags().GetString("address-name")
	if err != nil || addressName == "" {
		client.PrintErrorAndExit("Invalid address-name parameter")
	}

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintConnectionErrorAndExit(err)
	}
	defer conn.Close()

	cli := svcpb.NewServiceClient(conn)
	req := &svcpb.SetStaticIpRequest{
		AppName:     appName,
		AddressName: addressName,
	}
	if _, err := cli.SetStaticIp(context.Background(), req); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	fmt.Println("Static IP added with success")
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
	if len(info.SourceRanges) > 0 {
		fmt.Println(bold("whitelist:"))
		for _, item := range info.SourceRanges {
			fmt.Printf("  %s\n", item)
		}
	}
}

func serviceWhitelistSourceRanges(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		return
	}
	appName, ranges := args[0], args[1:]
	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintConnectionErrorAndExit(err)
	}
	defer conn.Close()
	cli := svcpb.NewServiceClient(conn)
	req := &svcpb.WhitelistSourceRangesRequest{
		AppName:      appName,
		SourceRanges: ranges,
	}
	if _, err := cli.WhitelistSourceRanges(context.Background(), req); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	fmt.Println("Firewall whitelist configured with success")
}

func init() {
	RootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceEnableSSLCmd)
	serviceCmd.AddCommand(serviceSetStaticIpCmd)
	serviceCmd.AddCommand(serviceInfoCmd)
	serviceCmd.AddCommand(serviceWhitelistSourceRangesCmd)

	serviceEnableSSLCmd.Flags().String("app", "", "app name")
	serviceEnableSSLCmd.Flags().String("cert", "", "certificate identifier")
	serviceEnableSSLCmd.Flags().Bool("only", false, "only use SSL")

	serviceSetStaticIpCmd.Flags().String("app", "", "app name")
	serviceSetStaticIpCmd.Flags().String("address-name", "", "static IP address name")
}
