package cmd

import (
	context "golang.org/x/net/context"

	"github.com/fatih/color"
	"github.com/luizalabs/teresa-api/cmd/client/connection"
	"github.com/luizalabs/teresa-api/pkg/client"
	"github.com/spf13/cobra"

	userpb "github.com/luizalabs/teresa-api/pkg/protobuf/user"
)

var userName string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login in the currently selected cluster",
	Long: `Login in the selected cluster.

eg.:

	$ teresa login --user user@mydomain.com
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

	conn, err := connection.New(cfgFile)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := userpb.NewUserClient(conn)
	res, err := cli.Login(context.Background(), &userpb.LoginRequest{Email: userName, Password: p})
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	color.Green("Login OK")

	if err = client.SaveToken(cfgFile, res.Token); err != nil {
		client.PrintErrorAndExit("Error trying to save token in configuration file: %v", err)
	}
}

func init() {
	loginCmd.Flags().StringVar(&userName, "user", "", "e-mail to login with")
	RootCmd.AddCommand(loginCmd)
}
