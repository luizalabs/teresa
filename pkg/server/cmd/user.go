package cmd

import (
	"fmt"
	"os"

	"github.com/luizalabs/teresa-api/pkg/client"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/user"
	"github.com/spf13/cobra"
)

var createSuperUserCmd = &cobra.Command{
	Use:   "create-super-user",
	Short: "Create an admin user",
	Run:   createSuperUser,
}

func init() {
	RootCmd.AddCommand(createSuperUserCmd)
	createSuperUserCmd.Flags().String("name", "Admin", "super user name")
	createSuperUserCmd.Flags().String("email", "", "super user email [required]")
	createSuperUserCmd.Flags().String("password", "", "super user password [required]")
}

func createSuperUser(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	email, _ := cmd.Flags().GetString("email")
	pass, _ := cmd.Flags().GetString("password")

	if pass == "" {
		fmt.Fprintln(os.Stderr, "Password required")
		return
	}
	if email == "" {
		fmt.Fprintln(os.Stderr, "E-mail required")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error on connect to Database: ", err)
		return
	}

	uOps := user.NewDatabaseOperations(db, auth.NewFake())
	if err := uOps.Create(name, email, pass, true); err != nil {
		fmt.Fprintln(os.Stderr, "Error on create super user:", client.GetErrorMsg(err))
		return
	}
	fmt.Println("Super User created")
}
