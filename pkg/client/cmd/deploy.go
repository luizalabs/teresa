package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/luizalabs/teresa/pkg/client"
	"github.com/luizalabs/teresa/pkg/client/connection"
	"github.com/luizalabs/teresa/pkg/client/tar"
	dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"
	"github.com/luizalabs/teresa/pkg/server/deploy"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Everything about deploys",
	Long: `To deploy your application you must use the new "teresa deploy create ..."

To see the "deploy create" help, please enter either "teresa deploy create --help"
or just "teresa deploy create".`,
}

var deployCreateCmd = &cobra.Command{
	Use:   "create <app folder>",
	Short: "Deploy an app",
	Long: `Deploy an application.
	
	To deploy an app you have to pass it's name, the team the app
	belongs and the path to the source code. You might want to
	describe your deployments through --description, as that'll
	eventually help on rollbacks.
	
	eg.:
	
	  $ teresa deploy create . --app webapi --description "release 1.2 with new checkout"
	`,
	Run: deployApp,
}

var deployListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List app deploys",
	Long:    "Return all deploys from a given app.",
	Example: "  $ teresa deploy list --app myapp",
	Run:     deployList,
}

var deployRollbackCmd = &cobra.Command{
	Use:     "rollback",
	Short:   "rollback app to a given revision",
	Long:    "Rollback an application to a given revision.",
	Example: "  $ teresa deploy rollback myapp --revision 1",
	Run:     deployRollback,
}

func getCurrentClusterName() (string, error) {
	cfg, err := client.ReadConfigFile(cfgFile)
	if err != nil {
		return "", err
	}
	if cfg.CurrentCluster == "" {
		return "", client.ErrInvalidConfigFile
	}
	return cfg.CurrentCluster, nil
}

func getIgnorePatterns(source string) ([]string, error) {
	fPath := filepath.Join(source, ".teresaignore")
	if _, err := os.Stat(fPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	file, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	patterns := make([]string, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if text := scanner.Text(); text != "" {
			patterns = append(patterns, text)
		}
	}

	if len(patterns) == 0 {
		return nil, nil
	}

	return patterns, nil
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.AddCommand(deployCreateCmd)
	deployCmd.AddCommand(deployListCmd)
	deployCmd.AddCommand(deployRollbackCmd)

	deployCreateCmd.Flags().String("app", "", "app name (required)")
	deployCreateCmd.Flags().String("description", "", "deploy description (required)")
	deployCreateCmd.Flags().Bool("no-input", false, "deploy app without warning")

	deployListCmd.Flags().String("app", "", "app name (required)")

	deployRollbackCmd.Flags().String("revision", "", "app revision (required)")
}

func deployApp(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		return
	}
	appFolder := args[0]

	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
	}

	deployDescription, err := cmd.Flags().GetString("description")
	if err != nil {
		client.PrintErrorAndExit("Invalid description parameter")
	}

	noInput, err := cmd.Flags().GetBool("no-input")
	if err != nil {
		client.PrintErrorAndExit("Invalid no-input parameter")
	}

	currentClusterName := cfgCluster
	if currentClusterName == "" {
		currentClusterName, err = getCurrentClusterName()
		if err != nil {
			client.PrintErrorAndExit("error reading config file: %v", err)
		}
	}

	fmt.Printf("Deploying app %s to the cluster %s...\n", color.CyanString(`"%s"`, appName), color.YellowString(`"%s"`, currentClusterName))

	if !noInput {
		fmt.Print("Are you sure? (yes/NO)? ")
		s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(s), "yes") {
			return
		}
	}

	ip, err := getIgnorePatterns(appFolder)
	if err != nil {
		client.PrintErrorAndExit("Error acessing .teresaignore file: %v", err)
	}

	fmt.Println("Generating tarball of:", appFolder)
	tarPath, err := tar.CreateTemp(appFolder, appName, ip)
	if err != nil {
		client.PrintErrorAndExit("Error generating tarball: %v", err)
	}
	defer os.Remove(tarPath)

	conn, err := connection.New(cfgFile, currentClusterName)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	cli := dpb.NewDeployClient(conn)
	stream, err := cli.Make(ctx)
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	info := &dpb.DeployRequest{Value: &dpb.DeployRequest_Info_{&dpb.DeployRequest_Info{
		App:         appName,
		Description: deployDescription,
	}}}
	if err := stream.Send(info); err != nil {
		client.PrintErrorAndExit("Error sending deploy information: %v", err)
	}

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error { return sendAppTarball(tarPath, stream) })
	g.Go(func() error { return streamServerMsgs(stream) })

	if err := g.Wait(); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
}

func sendAppTarball(tarPath string, stream dpb.Deploy_MakeClient) error {
	fmt.Println("Sending app tarbal...")
	defer stream.CloseSend()

	f, err := os.Open(tarPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading temp file:")
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		buf := make([]byte, 1024)
		_, err := r.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "Error reading bytes of temp file:")
				return err
			}
			break
		}

		bufMsg := &dpb.DeployRequest{Value: &dpb.DeployRequest_File_{&dpb.DeployRequest_File{
			Chunk: buf,
		}}}
		if err := stream.Send(bufMsg); err != nil {
			fmt.Fprintln(os.Stderr, "Error sending tarball chunk:")
			return err
		}
	}
	return nil
}

func streamServerMsgs(stream dpb.Deploy_MakeClient) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Print(msg.Text)
	}
	return nil
}

func deployList(cmd *cobra.Command, args []string) {
	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
	}

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := dpb.NewDeployClient(conn)
	resp, err := cli.List(context.Background(), &dpb.ListRequest{AppName: appName})
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	if len(resp.Deploys) == 0 {
		fmt.Println("App doesn't have any deploys")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"REVISION", "AGE", "DESCRIPTION"})
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowSeparator("-")
	table.SetAutoWrapText(false)

	sort.Sort(sort.Reverse(deploy.ByRevision(resp.Deploys)))
	for _, d := range resp.Deploys {
		if d.Current {
			d.Revision = fmt.Sprintf("%s (current)", d.Revision)
		}
		r := []string{
			d.Revision,
			shortHumanDuration(time.Duration(d.Age)),
			d.Description,
		}
		table.Append(r)
	}
	table.Render()
}

func deployRollback(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}
	appName := args[0]

	revision, err := cmd.Flags().GetString("revision")
	if err != nil || revision == "" {
		client.PrintErrorAndExit("invalid revision parameter")
	}

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	req := &dpb.RollbackRequest{
		AppName:  appName,
		Revision: revision,
	}
	cli := dpb.NewDeployClient(conn)
	if _, err = cli.Rollback(context.Background(), req); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	fmt.Println("rollback done")
}
