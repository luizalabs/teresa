package cmd

import (
	"bufio"
	"errors"
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
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Everything about deploys",
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

func createTempArchiveToUpload(appName, source string) (path string, err error) {
	id := uuid.NewV4()
	source, err = filepath.Abs(source)
	if err != nil {
		return "", err
	}
	p := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s.tar.gz", appName, id))
	if err = createArchive(source, p); err != nil {
		return "", err
	}
	return p, nil
}

func createArchive(source, target string) error {
	dir, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("Dir not found to create an archive: %s", err)
	} else if !dir.IsDir() {
		return errors.New("Path to create the app archive isn't a directory")
	}

	ignorePatterns, err := getIgnorePatterns(source)
	if err != nil {
		return errors.New("Invalid file '.teresaignore'")
	}

	t, err := tar.New(target)
	if err != nil {
		return err
	}
	defer t.Close()

	if ignorePatterns != nil {
		if err = addFiles(source, t, ignorePatterns); err != nil {
			return err
		}
	} else {
		t.AddAll(source)
	}
	return nil
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

func addFiles(source string, tar tar.Writer, ignorePatterns []string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, ip := range ignorePatterns {
			if matched, _ := filepath.Match(ip, info.Name()); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		basePath := fmt.Sprintf("%s%c", source, filepath.Separator)
		filename := strings.Replace(path, basePath, "", 1)
		return tar.AddFile(path, filename)
	})
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.AddCommand(deployCreateCmd)
	deployCmd.AddCommand(deployListCmd)

	deployCreateCmd.Flags().String("app", "", "app name (required)")
	deployCreateCmd.Flags().String("description", "", "deploy description (required)")
	deployCreateCmd.Flags().Bool("no-input", false, "deploy app without warning")

	deployListCmd.Flags().String("app", "", "app name (required)")
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

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return sendAppTarball(appName, appFolder, stream) })
	g.Go(func() error { return streamServerMsgs(stream) })

	if err := g.Wait(); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
}

func sendAppTarball(appName, appFolder string, stream dpb.Deploy_MakeClient) error {
	defer stream.CloseSend()
	fmt.Println("Generating tarball of:", appFolder)
	tarPath, err := createTempArchiveToUpload(appName, appFolder)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating tarball:")
		return err
	}

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
	table.SetHeader([]string{"REVISION", "CURRENT", "AGE", "DESCRIPTION"})
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowSeparator("-")
	table.SetAutoWrapText(false)

	sort.Sort(deploy.ByRevision(resp.Deploys))
	for _, d := range resp.Deploys {
		r := []string{
			d.Revision,
			fmt.Sprintf("%t", d.Current),
			shortHumanDuration(time.Duration(d.Age)),
			d.Description,
		}
		table.Append(r)
	}
	table.Render()
}
