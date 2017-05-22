package cmd

import (
	"fmt"
	"os"

	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa-api/cmd/client/connection"
	"github.com/luizalabs/teresa-api/pkg/client"
	_ "github.com/prometheus/common/log"
	"github.com/spf13/cobra"

	userpb "github.com/luizalabs/teresa-api/pkg/protobuf/user"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Create a user",
	Long: `Create a user.

Note that the user's password must be at least 8 characters long. eg.:

	$ teresa create user --email user@mydomain.com --name john --password foobarfoo
	`,
	Run: createUser,
}

// delete user
var deleteUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Delete an user",
	Long:  `Delete an user.`,
	Run:   deleteUser,
}

// set password for an user
var setUserPasswordCmd = &cobra.Command{
	Use:   "set-password",
	Short: "Set password for current user",
	Long:  `Set password for current user.`,
	Run:   setPassword,
}

func setPassword(cmd *cobra.Command, args []string) {
	p, err := client.GetMaskedPassword("New Password: ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error trying to get the user password: ", err)
	}

	conn, err := connection.New(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server: ", err)
	}
	defer conn.Close()

	cli := userpb.NewUserClient(conn)
	if _, err := cli.SetPassword(context.Background(), &userpb.SetPasswordRequest{Password: p}); err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}
	fmt.Println("Password updated")
}

func deleteUser(cmd *cobra.Command, args []string) {
	email, _ := cmd.Flags().GetString("email")
	if email == "" {
		cmd.Usage()
		return
	}
	conn, err := connection.New(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server: ", err)
		return
	}
	defer conn.Close()

	cli := userpb.NewUserClient(conn)
	_, err = cli.Delete(
		context.Background(),
		&userpb.DeleteRequest{Email: email},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}
	fmt.Println("User deleted")
}

func createUser(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid user parameter: ", err)
		return
	}
	email, err := cmd.Flags().GetString("email")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid email parameter: ", err)
		return
	}
	pass, err := cmd.Flags().GetString("password")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid password parameter: ", err)
		return
	}
	admin, err := cmd.Flags().GetBool("admin")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid admin parameter: ", err)
		return
	}
	if email == "" || name == "" || pass == "" {
		cmd.Usage()
		return
	}
	conn, err := connection.New(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server: ", err)
		return
	}
	defer conn.Close()

	cli := userpb.NewUserClient(conn)
	_, err = cli.Create(
		context.Background(),
		&userpb.CreateRequest{
			Name:     name,
			Email:    email,
			Password: pass,
			Admin:    admin,
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}
	fmt.Println("User created")
}

func init() {
	createCmd.AddCommand(userCmd)
	userCmd.Flags().String("name", "", "user name [required]")
	userCmd.Flags().String("email", "", "user email [required]")
	userCmd.Flags().String("password", "", "user password [required]")
	userCmd.Flags().Bool("admin", false, "admin privileges")

	deleteCmd.AddCommand(deleteUserCmd)
	deleteUserCmd.Flags().String("email", "", "user email [required]")

	RootCmd.AddCommand(setUserPasswordCmd)
}
