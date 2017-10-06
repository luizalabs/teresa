package cmd

import (
	"time"

	context "golang.org/x/net/context"

	"github.com/fatih/color"
	"github.com/luizalabs/teresa/pkg/client"
	"github.com/luizalabs/teresa/pkg/client/connection"
	"github.com/spf13/cobra"

	userpb "github.com/luizalabs/teresa/pkg/protobuf/user"
)

var (
	userName  string
	expiresIn time.Duration
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authorize access to the selected cluster",
	Long: `Authorize access to the selected cluster.

eg.:

	$ teresa login --user user@mydomain.com [--expires-in 168h]

Where valid "expires-in" units are "ns", "us" (or "µs"), "ms", "s", "m", "h",
as accepted by Go's time.ParseDuration.
	`,
	Run: login,
}

func login(cmd *cobra.Command, args []string) {
	if userName == "" {
		cmd.Usage()
		return
	}
	p, err := client.GetMaskedPassword("Password: ")
	if err != nil {
		client.PrintErrorAndExit("Error trying to get the user password: %v", err)
	}

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	exp := float64(expiresIn)
	cli := userpb.NewUserClient(conn)
	res, err := cli.Login(context.Background(), &userpb.LoginRequest{Email: userName, Password: p, ExpiresIn: exp})
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	color.Green("Login OK")

	if err = client.SaveToken(cfgFile, cfgCluster, res.Token); err != nil {
		client.PrintErrorAndExit("Error trying to save token in configuration file: %v", err)
	}
}

func init() {
	loginCmd.Flags().StringVar(&userName, "user", "", "e-mail to login with (required)")
	loginCmd.Flags().DurationVar(&expiresIn, "expires-in", 15*24*time.Hour, "duration of login token")
	RootCmd.AddCommand(loginCmd)
}
