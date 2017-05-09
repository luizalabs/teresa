package cmd

import (
	"fmt"
	"os"

	context "golang.org/x/net/context"

	"github.com/howeyc/gopass"
	"github.com/luizalabs/teresa-api/cmd/client/connection"
	"github.com/luizalabs/teresa-api/pkg/client"
	"github.com/spf13/cobra"

	userpb "github.com/luizalabs/teresa-api/pkg/protobuf"
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

	fmt.Print("Password: ")
	p, err := gopass.GetPasswdMasked()
	if err != nil {
		if err != gopass.ErrInterrupted {
			fmt.Fprintln(os.Stderr, "Error trying to get the user password: ", err)
		}
		return
	}
	conn, err := connection.New(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server: ", err)
		return
	}
	defer conn.Close()

	cli := userpb.NewUserClient(conn)
	res, err := cli.Login(context.Background(), &userpb.LoginRequest{Email: userName, Password: string(p)})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error trying to login in cluster: ", err)
		return
	}
	fmt.Println("Login OK")

	if err = client.SaveToken(cfgFile, res.Token); err != nil {
		fmt.Fprintln(os.Stderr, "Error trying to save token in configuration file: ", err)
	}
}

func init() {
	loginCmd.Flags().StringVar(&userName, "user", "", "username to login with")
	RootCmd.AddCommand(loginCmd)
}
