package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/luizalabs/teresa/pkg/client"
	"github.com/luizalabs/teresa/pkg/client/connection"
	appb "github.com/luizalabs/teresa/pkg/protobuf/app"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"golang.org/x/net/context"
)

const (
	//Service name component must be a valid RFC 1035 name
	appNameLimit   = 63
	flagNotDefined = -1
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Everything about apps",
}

var appCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Creates a new app",
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
  $ teresa create foo --team bar --cpu 200m --max-cpu 500m --memory 512Mi --max-memory 1Gi

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
	if len(name) > appNameLimit {
		client.PrintErrorAndExit("Invalid app name (max %d characters)", appNameLimit)
	}

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
	Run:     appList,
}

func appList(cmd *cobra.Command, args []string) {
	conn, err := connection.New(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := appb.NewAppClient(conn)
	resp, err := cli.List(context.Background(), &appb.Empty{})
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	if len(resp.Apps) == 0 {
		fmt.Println("You don't have any app")
		return
	}
	// rendering app list in a table view
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"TEAM", "APP", "ADDRESS"})
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowSeparator("-")
	table.SetAutoWrapText(false)
	for _, a := range resp.Apps {
		urls := strings.Join(a.Urls, ",")
		if urls == "" {
			urls = "n/a"
		}
		r := []string{a.Team, a.Name, urls}
		table.Append(r)
	}
	table.Render()
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
			age := shortHumanDuration(time.Duration(pod.Age))
			fmt.Printf("    Name: %s  State: %s  Age: %s  Restarts: %d\n", pod.Name, pod.State, age, pod.Restarts)
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
	if len(args) == 0 {
		cmd.Usage()
		return
	}

	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
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
	Example: `  To unset an env var called "FOO":

  $ teresa app env-unset FOO --app myapp

You can also provide more than one env var at a time:

  $ teresa app env-unset FOO BAR --app myapp`,
	Run: appEnvUnset,
}

func appEnvUnset(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		return
	}

	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
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

var appAutoScaleSetCmd = &cobra.Command{
	Use:   "autoscale <name> [flags]",
	Short: "Set autoscale parameters for the app",
	Long: `Set application's autoscaling.

You can set the lower and upper limit for the number of pods of the application, as well as the
target CPU utilization to trigger the autoscaler.

	Example:   To set the number minimum of replicas to 2:

  $ teresa app autoscale myapp --min 2`,
	Run: appAutoScaleSet,
}

func appAutoScaleSet(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	name := args[0]

	min, err := cmd.Flags().GetInt32("min")
	if err != nil {
		client.PrintErrorAndExit("invalid min parameter")
	}

	max, err := cmd.Flags().GetInt32("max")
	if err != nil {
		client.PrintErrorAndExit("invalid max parameter")
	}

	cpu, err := cmd.Flags().GetInt32("cpu-percent")
	if err != nil {
		client.PrintErrorAndExit("invalid cpu-percent parameter")
	}

	if msg, isValid := validateFlags(min, max); !isValid {
		client.PrintErrorAndExit(msg)
	}

	conn, err := connection.New(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %s", err)
	}
	defer conn.Close()

	as := &appb.SetAutoScaleRequest_AutoScale{
		Min:                  min,
		Max:                  max,
		CpuTargetUtilization: cpu,
	}
	req := &appb.SetAutoScaleRequest{
		Name:      name,
		AutoScale: as,
	}
	cli := appb.NewAppClient(conn)
	if _, err := cli.SetAutoScale(context.Background(), req); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	fmt.Println("Autoscale updated with success")
}

func validateFlags(min, max int32) (string, bool) {
	if max == flagNotDefined || min == flagNotDefined {
		return "--min and --max are required", false
	}
	if max < 1 {
		return fmt.Sprintf("--max=MAXPODS must be at least 1, max: %d", max), false
	}
	if max < min {
		return fmt.Sprintf("--max=MAXPODS must be larger or equal to --min=MINPODS, max: %d, min: %d", max, min), false
	}
	if min < 1 {
		return fmt.Sprintf("--min=MINPODS must be at least 1, min: %d", min), false
	}

	return "", true
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
	appCmd.AddCommand(appAutoScaleSetCmd)

	appCreateCmd.Flags().String("team", "", "team owner of the app")
	appCreateCmd.Flags().Int32("scale-min", 1, "minimium number of replicas")
	appCreateCmd.Flags().Int32("scale-max", 2, "maximum number of replicas")
	appCreateCmd.Flags().Int32("scale-cpu", 70, "auto scale target cpu percentage to scale")
	appCreateCmd.Flags().String("cpu", "200m", "allocated pod cpu")
	appCreateCmd.Flags().String("memory", "512Mi", "allocated pod memory")
	appCreateCmd.Flags().String("max-cpu", "200m", "when set, allows the pod to burst cpu usage up to 'max-cpu'")
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
	// App autoscale
	appAutoScaleSetCmd.Flags().Int32("min", flagNotDefined, "Minimium number of replicas")
	appAutoScaleSetCmd.Flags().Int32("max", flagNotDefined, "Maximum number of replicas")
	appAutoScaleSetCmd.Flags().Int32("cpu-percent", flagNotDefined, "The target average CPU utilization (represented as a percent of requested CPU) over all the pods. If it's not specified or negative, the current autoscaling policy will be used.")
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

// Shamelessly copied from Kubernetes
func shortHumanDuration(d time.Duration) string {
	// Allow deviation no more than 2 seconds(excluded) to tolerate machine time
	// inconsistence, it can be considered as almost now.
	if seconds := int(d.Seconds()); seconds < -1 {
		return fmt.Sprintf("<invalid>")
	} else if seconds < 0 {
		return fmt.Sprintf("0s")
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}
