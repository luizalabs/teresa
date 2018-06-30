package cmd

import (
	"fmt"
	"io"

	"github.com/luizalabs/teresa/pkg/client"
	"github.com/luizalabs/teresa/pkg/client/connection"
	execpb "github.com/luizalabs/teresa/pkg/protobuf/exec"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Exec a command on an app replica",
	Long: `Exec a command on an app replica.

You can execute a non-interactive command on an app replica (same of current deploy),
Teresa will collect and stream the stdout of replica until the command ends`,
	Example: "  $ teresa exec <app-name> -- python manage.py start_job_x -a arg1 -s arg2",
	Run:     execCommand,
}

func execCommand(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Usage()
		return
	}
	appName := args[0]
	command := args[1:]

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintConnectionErrorAndExit(err)
	}
	defer conn.Close()

	req := &execpb.CommandRequest{
		AppName: appName,
		Command: command,
	}

	cli := execpb.NewExecClient(conn)
	stream, err := cli.Command(context.Background(), req)
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
		fmt.Print(msg.Text)
	}

}

func init() {
	RootCmd.AddCommand(execCmd)
}
