package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/luizalabs/teresa-api/cmd/client/connection"
	"github.com/luizalabs/teresa-api/pkg/client"
	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"golang.org/x/net/context"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Everything about apps",
}

var appCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Creates a new app.",
	Long: `Creates a new application.

You should provide the unique name for the App and the team name in order to create a new App.
Remember, all team members can view and modify this newly created app.

The app name must follow this rules:
  - must contain only letters and the special character "-"
  - must start and finish with a letter
  - something like: foo or foo-bar`,
	Example: `  $ teresa app create foo --team bar

  With specific process type:
  $ teresa app create foo-worker --team bar --process-type worker

  With scale rules... min 2, max 10 pods, scalling with a cpu target of 70%
  $ teresa app create foo --team bar --scale-min 2 --scale-max 10 --scale-cpu 70

  With specific cpu and memory size...
  $ teresa create foo --team bar --cpu 200m --max-cpu 1Gi --memory 512Mi --max-memory 1Gi

  With all flags...
  $ teresa app create foo --team bar --cpu 200m --max-cpu 500m --memory 512Mi --max-memory 1Gi --scale-min 2 --scale-max 10 --scale-cpu 70 --process-type web`,
	Run: createApp,
}

func createApp(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}
	name := args[0]

	team, err := cmd.Flags().GetString("team")
	if err != nil || team == "" {
		client.PrintErrorAndExit("Invalid team parameter")
	}

	targetCPU, err := cmd.Flags().GetInt32("scale-cpu")
	if err != nil {
		client.PrintErrorAndExit("Invalid scale-cpu parameter")
	}

	scaleMax, err := cmd.Flags().GetInt32("scale-max")
	if err != nil {
		client.PrintErrorAndExit("Invalid scale-max parameter")
	}

	scaleMin, err := cmd.Flags().GetInt32("scale-min")
	if err != nil {
		client.PrintErrorAndExit("Invalid scale-min parameter")
	}

	cpu, err := cmd.Flags().GetString("cpu")
	if err != nil {
		client.PrintErrorAndExit("Invalid cpu parameter")
	}

	memory, err := cmd.Flags().GetString("memory")
	if err != nil {
		client.PrintErrorAndExit("Invalid memory parameter")
	}

	maxCPU, err := cmd.Flags().GetString("max-cpu")
	if err != nil {
		client.PrintErrorAndExit("Invalid max-cpu parameter")
	}

	maxMemory, err := cmd.Flags().GetString("max-memory")
	if err != nil {
		client.PrintErrorAndExit("Invalid max-memory parameter")
	}

	processType, err := cmd.Flags().GetString("process-type")
	if err != nil {
		client.PrintErrorAndExit("Invalid process-type parameter")
	}

	conn, err := connection.New(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	lim := &appb.CreateRequest_Limits{
		Default: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			&appb.CreateRequest_Limits_LimitRangeQuantity{
				Resource: "cpu",
				Quantity: maxCPU,
			},
			&appb.CreateRequest_Limits_LimitRangeQuantity{
				Resource: "memory",
				Quantity: maxMemory,
			},
		},
		DefaultRequest: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			&appb.CreateRequest_Limits_LimitRangeQuantity{
				Resource: "cpu",
				Quantity: cpu,
			},
			&appb.CreateRequest_Limits_LimitRangeQuantity{
				Resource: "memory",
				Quantity: memory,
			},
		},
	}
	as := &appb.CreateRequest_AutoScale{
		CpuTargetUtilization: targetCPU,
		Min:                  scaleMin,
		Max:                  scaleMax,
	}
	cli := appb.NewAppClient(conn)
	_, err = cli.Create(
		context.Background(),
		&appb.CreateRequest{
			Name:        name,
			Team:        team,
			ProcessType: processType,
			Limits:      lim,
			AutoScale:   as,
		},
	)
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	fmt.Println("App created")
}

var appListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all apps",
	Long:    "Return all apps with address and team.",
	Example: "  $ teresa app list",
	RunE: func(cmd *cobra.Command, args []string) error {
		tc := NewTeresa()
		apps, err := tc.GetApps()
		if err != nil {
			return err
		}
		// rendering app info in a table view
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"TEAM", "APP", "ADDRESS"})
		table.SetRowLine(true)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetRowSeparator("-")
		table.SetAutoWrapText(false)
		for _, app := range apps {
			a := ""
			if len(app.AddressList) > 0 {
				a = app.AddressList[0]
			}
			r := []string{*app.Team, *app.Name, a}
			table.Append(r)
		}
		table.Render()
		return nil
	},
}

var appInfoCmd = &cobra.Command{
	Use:     "info <name>",
	Short:   "All infos about the app",
	Long:    "Return all infos about an specific app, like addresses, scale, auto scale, etc...",
	Example: "  $ teresa app info foo",
	Run:     appInfo,
}

func appInfo(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}
	name := args[0]

	conn, err := connection.New(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := appb.NewAppClient(conn)
	info, err := cli.Info(context.Background(), &appb.InfoRequest{Name: name})
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	color.New(color.FgCyan, color.Bold).Printf("[%s]\n", name)
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println(bold("team:"), info.Team)
	if len(info.Addresses) > 0 {
		fmt.Println(bold("addresses:"))
		for _, addr := range info.Addresses {
			fmt.Printf("  %s\n", addr.Hostname)
		}
	}
	if len(info.EnvVars) > 0 {
		fmt.Println(bold("env vars:"))
		for _, ev := range info.EnvVars {
			fmt.Printf("  %s=%s\n", ev.Key, ev.Value)
		}
	}
	if info.Status != nil {
		fmt.Println(bold("status:"))
		fmt.Printf("  %s %d%%\n", bold("cpu:"), info.Status.Cpu)
		fmt.Printf("  %s %d\n", bold("pods:"), len(info.Status.Pods))
		for _, pod := range info.Status.Pods {
			fmt.Printf("    %s %s  %s %s\n", "Name:", pod.Name, "State:", pod.State)
		}
	}
	if info.AutoScale != nil {
		fmt.Println(bold("autoscale:"))
		fmt.Printf("  %s %d%%\n", bold("cpu:"), info.AutoScale.CpuTargetUtilization)
		fmt.Printf("  %s %d\n", bold("max:"), info.AutoScale.Max)
		fmt.Printf("  %s %d\n", bold("min:"), info.AutoScale.Min)
	}
	fmt.Println(bold("limits:"))
	if len(info.Limits.Default) > 0 {
		fmt.Println(bold("  defaults"))
		for _, item := range info.Limits.Default {
			fmt.Printf("    %s %s\n", bold(item.Resource), item.Quantity)
		}
	}
	if len(info.Limits.DefaultRequest) > 0 {
		fmt.Println(bold("  request"))
		for _, item := range info.Limits.DefaultRequest {
			fmt.Printf("    %s %s\n", bold(item.Resource), item.Quantity)
		}
	}
}

var appEnvSetCmd = &cobra.Command{
	Use:   "env-set [KEY=value, ...]",
	Short: "Set env vars for the app",
	Long: `Create or update an environment variable for the app.

You can add a new environment variable for the app, or update if it already exists.

WARNING:
  If you need to set more than one env var to the application, provide all at once.
  Every time this command is called, the application needs to be restared.`,
	Example: `  To add an new env var called "FOO":

  $ teresa app env-set FOO=bar --app myapp

  You can also provide more than one env var at a time:

  $ teresa app env-set FOO=bar BAR=foo --app myapp`,
	Run: appEnvSet,
}

func appEnvSet(cmd *cobra.Command, args []string) {
	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
	}

	if len(args) == 0 {
		cmd.Usage()
		return
	}

	evs := make([]*appb.SetEnvRequest_EnvVar, len(args))
	for i, item := range args {
		tmp := strings.SplitN(item, "=", 2)
		if len(tmp) != 2 {
			client.PrintErrorAndExit("Env vars must be in the format FOO=bar")
		}
		evs[i] = &appb.SetEnvRequest_EnvVar{Key: tmp[0], Value: tmp[1]}
	}

	fmt.Printf("Setting env vars and %s %s...\n", color.YellowString("restarting"), color.CyanString(`"%s"`, appName))
	for _, ev := range evs {
		fmt.Printf("  %s: %s\n", ev.Key, ev.Value)
	}

	noinput, err := cmd.Flags().GetBool("no-input")
	if err != nil {
		client.PrintErrorAndExit("Invalid no-input parameter")
	}
	if !noinput {
		fmt.Print("Are you sure? (yes/NO)? ")
		s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		s = strings.ToLower(strings.TrimRight(s, "\r\n"))
		if s != "yes" {
			return
		}
	}

	conn, err := connection.New(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %s", err)
	}
	defer conn.Close()

	cli := appb.NewAppClient(conn)
	req := &appb.SetEnvRequest{Name: appName, EnvVars: evs}
	if _, err := cli.SetEnv(context.Background(), req); err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}
	fmt.Println("Env vars updated with success")
}

var appEnvUnSetCmd = &cobra.Command{
	Use:   "env-unset [KEY, ...]",
	Short: "Unset env vars for the app",
	Long: `Unset an environment variable for the app.

You can remove one or more environment variables from the application.

WARNING:
  If you need to unset more than one env var from the application, provide all at once.
  Every time this command is called, the application needs to be restarted.`,
	Example: `  To add an new env var called "FOO":

To unset an env var called "FOO":

  $ teresa app env-unset FOO --app myapp

You can also provide more than one env var at a time:

  $ teresa app env-unset FOO BAR --app myapp`,
	Run: appEnvUnset,
}

func appEnvUnset(cmd *cobra.Command, args []string) {
	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
	}

	if len(args) == 0 {
		cmd.Usage()
		return
	}

	fmt.Printf("Unsetting env vars and %s %s...\n", color.YellowString("restarting"), color.CyanString(`"%s"`, appName))
	for _, ev := range args {
		fmt.Printf("  %s\n", ev)
	}

	noinput, err := cmd.Flags().GetBool("no-input")
	if err != nil {
		client.PrintErrorAndExit("Invalid no-input parameter")
	}
	if !noinput {
		fmt.Print("Are you sure? (yes/NO)? ")
		s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		s = strings.ToLower(strings.TrimRight(s, "\r\n"))
		if s != "yes" {
			return
		}
	}

	conn, err := connection.New(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %s", err)
	}
	defer conn.Close()

	cli := appb.NewAppClient(conn)
	req := &appb.UnsetEnvRequest{Name: appName, EnvVars: args}
	if _, err := cli.UnsetEnv(context.Background(), req); err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}
	fmt.Println("Env vars updated with success")
}

var appLogsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Show app logs",
	Long: `Show application logs.

WARNING:
  Lines are collected from all pods.`,
	Example: `  $ teresa app logs foo

  To change the number of lines:

  $ teresa app logs foo --lines=20

  You can also simulate tail -f:

  $ teresa app logs foo --lines=20 --follow`,
	Run: appLogs,
}

func init() {
	// add AppCmd
	RootCmd.AddCommand(appCmd)
	// App commands
	appCmd.AddCommand(appCreateCmd)
	appCmd.AddCommand(appListCmd)
	appCmd.AddCommand(appInfoCmd)
	appCmd.AddCommand(appEnvSetCmd)
	appCmd.AddCommand(appEnvUnSetCmd)
	appCmd.AddCommand(appLogsCmd)

	appCreateCmd.Flags().String("team", "", "team owner of the app")
	appCreateCmd.Flags().Int32("scale-min", 1, "auto scale min size")
	appCreateCmd.Flags().Int32("scale-max", 2, "auto scale max size")
	appCreateCmd.Flags().Int32("scale-cpu", 70, "auto scale target cpu percentage to scale")
	appCreateCmd.Flags().String("cpu", "200m", "allocated pod cpu")
	appCreateCmd.Flags().String("memory", "512Mi", "allocated pod memory")
	appCreateCmd.Flags().String("max-cpu", "500m", "when set, allows the pod to burst cpu usage up to 'max-cpu'")
	appCreateCmd.Flags().String("max-memory", "512Mi", "when set, allows the pod to burst memory usage up to 'max-memory'")
	appCreateCmd.Flags().String("process-type", "", "app process type")

	appEnvSetCmd.Flags().String("app", "", "app name")
	appEnvSetCmd.Flags().Bool("no-input", false, "set env vars without warning")
	// App unset env vars
	appEnvUnSetCmd.Flags().String("app", "", "app name")
	appEnvUnSetCmd.Flags().Bool("no-input", false, "unset env vars without warning")
	// App logs
	appLogsCmd.Flags().Int64("lines", 10, "number of lines")
	appLogsCmd.Flags().Bool("follow", false, "follow logs")
}

func appLogs(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}
	appName := args[0]
	lines, _ := cmd.Flags().GetInt64("lines")
	follow, _ := cmd.Flags().GetBool("follow")

	conn, err := connection.New(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := appb.NewAppClient(conn)
	req := &appb.LogsRequest{Name: appName, Lines: lines, Follow: follow}
	stream, err := cli.Logs(context.Background(), req)
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return
			}
			client.PrintErrorAndExit(client.GetErrorMsg(err))
		}
		fmt.Println(msg.Text)
	}
}
