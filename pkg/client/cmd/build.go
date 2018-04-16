package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/luizalabs/teresa/pkg/client"
	"github.com/luizalabs/teresa/pkg/client/connection"
	"github.com/luizalabs/teresa/pkg/client/tar"
	bpb "github.com/luizalabs/teresa/pkg/protobuf/build"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Everything about build",
	Long: `To build your application you must use the "teresa build create ..."

To see the "build create" help, please enter either "teresa build create --help"
or just "teresa build create".`,
}

var buildCreateCmd = &cobra.Command{
	Use:   "create <app folder>",
	Short: "Create a build of an app",
	Long: `Create a new build of an application.

The build create command follow almost the same rules of
deploy create command, you have to pass it's name, the path,
filename or url to the source code and a build name
(following the RFC-952), this name will help you
to promote a build to a new deploy.
	`,
	Example: "  $ teresa build create . --app myapp --name v1-0-0-rc1",
	Run:     buildApp,
}

func buildApp(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		return
	}
	appURL := args[0]

	appName, err := cmd.Flags().GetString("app")
	if err != nil || appName == "" {
		client.PrintErrorAndExit("Invalid app parameter")
	}

	buildName, err := cmd.Flags().GetString("name")
	if err != nil {
		client.PrintErrorAndExit("Invalid name parameter")
	}

	runApp, err := cmd.Flags().GetBool("run")
	if err != nil {
		client.PrintErrorAndExit("Invalid run parameter")
	}

	currentClusterName := cfgCluster
	if currentClusterName == "" {
		currentClusterName, err = getCurrentClusterName()
		if err != nil {
			client.PrintErrorAndExit("error reading config file: %v", err)
		}
	}

	fmt.Printf(
		"Creating build %s of app %s to the cluster %s...\n",
		color.CyanString(`"%s"`, buildName),
		color.CyanString(`"%s"`, appName),
		color.YellowString(`"%s"`, currentClusterName),
	)

	path, cleanup := fetchApp(appURL)
	if cleanup {
		defer os.Remove(path)
	}

	dir, cleanup := extractApp(path)
	if cleanup {
		defer os.RemoveAll(dir)
	}

	ip, err := getIgnorePatterns(dir)
	if err != nil {
		client.PrintErrorAndExit("Error acessing .teresaignore file: %v", err)
	}

	fmt.Println("Generating tarball of:", appURL)
	tarPath, err := tar.CreateTemp(dir, appName, ip)
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

	cli := bpb.NewBuildClient(conn)
	stream, err := cli.Make(ctx)
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	info := &bpb.BuildRequest{Value: &bpb.BuildRequest_Info_{&bpb.BuildRequest_Info{
		App:  appName,
		Name: buildName,
		Run:  runApp,
	}}}
	if err := stream.Send(info); err != nil {
		client.PrintErrorAndExit("Error sending build information: %v", err)
	}

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error { return sendBuildTarball(tarPath, stream) })
	g.Go(func() error { return streamServerBuildMsgs(stream) })

	if err := g.Wait(); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
}

func streamServerBuildMsgs(stream bpb.Build_MakeClient) error {
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

func sendBuildTarball(tarPath string, stream bpb.Build_MakeClient) error {
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

		bufMsg := &bpb.BuildRequest{Value: &bpb.BuildRequest_File_{&bpb.BuildRequest_File{
			Chunk: buf,
		}}}
		if err := stream.Send(bufMsg); err != nil {
			fmt.Fprintln(os.Stderr, "Error sending tarball chunk:")
			return err
		}
	}
	return nil
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.AddCommand(buildCreateCmd)

	buildCreateCmd.Flags().String("app", "", "app name (required)")
	buildCreateCmd.Flags().String("name", "", "build name (required)")
	buildCreateCmd.Flags().Bool("run", false, "run build in an isolate replica with a temporary service")
}
