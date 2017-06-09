package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/luizalabs/teresa-api/cmd/client/connection"
	"github.com/luizalabs/teresa-api/models"
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
		fmt.Fprintln(os.Stderr, "Invalid team parameter: ", err)
		return
	}

	targetCPU, err := cmd.Flags().GetInt32("scale-cpu")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid scale-cpu parameter: ", err)
		return
	}

	scaleMax, err := cmd.Flags().GetInt32("scale-max")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid scale-max parameter: ", err)
		return
	}

	scaleMin, err := cmd.Flags().GetInt32("scale-min")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid scale-min parameter: ", err)
		return
	}

	cpu, err := cmd.Flags().GetString("cpu")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid cpu parameter: ", err)
		return
	}

	memory, err := cmd.Flags().GetString("memory")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid memory parameter: ", err)
		return
	}

	maxCPU, err := cmd.Flags().GetString("max-cpu")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid max-cpu parameter: ", err)
		return
	}

	maxMemory, err := cmd.Flags().GetString("max-memory")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid max-memory parameter: ", err)
		return
	}

	processType, err := cmd.Flags().GetString("process-type")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid process-type parameter: ", err)
		return
	}

	conn, err := connection.New(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server: ", err)
		return
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
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
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
			if isNotFound(err) {
				return newCmdError("You have no apps")
			}
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return newUsageError("You should provide the name of the app in order to continue")
		}
		appName := args[0]
		tc := NewTeresa()
		app, err := tc.GetAppInfo(appName)
		if err != nil {
			if isNotFound(err) {
				return newCmdError("App not found")
			}
			return err
		}

		color.New(color.FgCyan, color.Bold).Printf("[%s]\n", *app.Name)
		bold := color.New(color.Bold).SprintFunc()

		fmt.Println(bold("team:"), *app.Team)
		if len(app.AddressList) > 0 {
			fmt.Println(bold("addresses:"))
			for _, a := range app.AddressList {
				fmt.Printf("  %s\n", a)
			}
		}
		if len(app.EnvVars) > 0 {
			fmt.Println(bold("env vars:"))
			for _, e := range app.EnvVars {
				fmt.Printf("  %s=%s\n", *e.Key, *e.Value)
			}
		}
		if app.Status != nil && (app.Status.CPU != nil || app.Status.Pods != nil) {
			fmt.Println(bold("status:"))
			if app.Status.CPU != nil {
				fmt.Printf("  %s %d%%\n", bold("cpu:"), *app.Status.CPU)
			}
			if app.Status.Pods != nil {
				fmt.Printf("  %s %d\n", bold("pods:"), *app.Status.Pods)
			}
		}
		if app.AutoScale != nil {
			fmt.Println(bold("autoscale:"))
			if app.AutoScale.CPUTargetUtilization != nil {
				fmt.Printf("  %s %d%%\n", bold("cpu:"), *app.AutoScale.CPUTargetUtilization)
			}
			fmt.Printf("  %s %d\n", bold("max:"), app.AutoScale.Max)
			fmt.Printf("  %s %d\n", bold("min:"), app.AutoScale.Min)
		}
		fmt.Println(bold("limits:"))
		if len(app.Limits.Default) > 0 {
			fmt.Println(bold("  defaults"))
			for _, l := range app.Limits.Default {
				fmt.Printf("    %s %s\n", bold(*l.Resource), *l.Quantity)
			}
		}
		if len(app.Limits.DefaultRequest) > 0 {
			fmt.Println(bold("  request"))
			for _, l := range app.Limits.DefaultRequest {
				fmt.Printf("    %s %s\n", bold(*l.Resource), *l.Quantity)
			}
		}
		if app.HealthCheck != nil && (app.HealthCheck.Liveness != nil || app.HealthCheck.Readiness != nil) {
			fmt.Println(bold("healthcheck:"))
			if app.HealthCheck.Liveness != nil {
				fmt.Println(bold("  liveness:"))
				fmt.Printf("    %s %s\n", bold("path:"), app.HealthCheck.Liveness.Path)
				fmt.Printf("    %s %ds\n", bold("period:"), app.HealthCheck.Liveness.PeriodSeconds)
				fmt.Printf("    %s %ds\n", bold("timeout:"), app.HealthCheck.Liveness.TimeoutSeconds)
				fmt.Printf("    %s %ds\n", bold("initial delay:"), app.HealthCheck.Liveness.InitialDelaySeconds)
				fmt.Printf("    %s %d\n", bold("success threshold:"), app.HealthCheck.Liveness.SuccessThreshold)
				fmt.Printf("    %s %d\n", bold("failure threshold:"), app.HealthCheck.Liveness.FailureThreshold)
			}
			if app.HealthCheck.Readiness != nil {
				fmt.Println(bold("  readiness:"))
				fmt.Printf("    %s %s\n", bold("path:"), app.HealthCheck.Readiness.Path)
				fmt.Printf("    %s %ds\n", bold("period:"), app.HealthCheck.Readiness.PeriodSeconds)
				fmt.Printf("    %s %ds\n", bold("timeout:"), app.HealthCheck.Readiness.TimeoutSeconds)
				fmt.Printf("    %s %ds\n", bold("initial delay:"), app.HealthCheck.Readiness.InitialDelaySeconds)
				fmt.Printf("    %s %d\n", bold("success threshold:"), app.HealthCheck.Readiness.SuccessThreshold)
				fmt.Printf("    %s %d\n", bold("failure threshold:"), app.HealthCheck.Readiness.FailureThreshold)
			}
		}
		return nil
	},
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
	RunE: func(cmd *cobra.Command, args []string) error {
		appName, _ := cmd.Flags().GetString("app")
		if appName == "" {
			return newUsageError("You should provide the name of the app in order to continue")
		}
		if len(args) == 0 {
			return newUsageError("You should provide env vars following the examples...")
		}
		// parsing env vars from args...
		evars := make([]*models.PatchAppEnvVar, len(args))
		for i, s := range args {
			x := strings.SplitN(s, "=", 2)
			if len(x) != 2 {
				return newUsageError("Env vars must be in the format FOO=bar")
			}
			evars[i] = &models.PatchAppEnvVar{
				Key:   &x[0],
				Value: x[1],
			}
		}
		// creating body for request
		action := "add"
		path := "/envvars"
		op := &models.PatchAppRequest{
			Op:    &action,
			Path:  &path,
			Value: evars,
		}
		fmt.Printf("Setting env vars and %s %s...\n", color.YellowString("restarting"), color.CyanString(`"%s"`, appName))
		for _, e := range evars {
			fmt.Printf("  %s: %s\n", *e.Key, e.Value)
		}
		noinput, _ := cmd.Flags().GetBool("no-input")
		if noinput == false {
			fmt.Print("Are you sure? (yes/NO)? ")
			// Waiting for the user answer...
			s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			s = strings.ToLower(strings.TrimRight(s, "\r\n"))
			if s != "yes" {
				return nil
			}
		}
		// Updating env vars
		tc := NewTeresa()
		_, err := tc.PartialUpdateApp(appName, []*models.PatchAppRequest{op})
		if err != nil {
			if isNotFound(err) {
				return newCmdError("App not found")
			} else if isBadRequest(err) {
				return newCmdError("Manualy change system env vars are not allowed")
			}
			return err
		}
		fmt.Println("Env vars updated with success")
		return nil
	},
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
	RunE: func(cmd *cobra.Command, args []string) error {
		appName, _ := cmd.Flags().GetString("app")
		if appName == "" {
			return newUsageError("You should provide the name of the app in order to continue")
		}
		if len(args) == 0 {
			return newUsageError("You should provide env vars following the examples...")
		}
		// parsing env vars from args...
		evars := make([]*models.PatchAppEnvVar, len(args))
		for i, k := range args {
			key := k
			e := models.PatchAppEnvVar{
				Key: &key,
			}
			evars[i] = &e
		}
		// creating body for request
		action := "remove"
		path := "/envvars"
		op := &models.PatchAppRequest{
			Op:    &action,
			Path:  &path,
			Value: evars,
		}
		fmt.Printf("Unsetting env vars and %s %s...\n", color.YellowString("restarting"), color.CyanString(`"%s"`, appName))
		for _, e := range evars {
			fmt.Printf("  %s\n", *e.Key)
		}
		noinput, _ := cmd.Flags().GetBool("no-input")
		if noinput == false {
			fmt.Print("Are you sure? (yes/NO)? ")
			// Waiting for the user answer...
			s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			s = strings.ToLower(strings.TrimRight(s, "\r\n"))
			if s != "yes" {
				return nil
			}
		}
		// Updating env vars
		tc := NewTeresa()
		_, err := tc.PartialUpdateApp(appName, []*models.PatchAppRequest{op})
		if err != nil {
			if isNotFound(err) {
				return newCmdError("App not found")
			} else if isBadRequest(err) {
				return newCmdError("Manualy change system env vars are not allowed")
			}
			return err
		}
		fmt.Println("Env vars updated with success")
		return nil
	},
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
		fmt.Fprintln(os.Stderr, "Error connecting to server:", err)
		return
	}
	defer conn.Close()

	cli := appb.NewAppClient(conn)
	req := &appb.LogsRequest{Name: appName, Lines: lines, Follow: follow}
	stream, err := cli.Logs(context.Background(), req)
	if err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
			return
		}
		fmt.Println(msg.Text)
	}
}
