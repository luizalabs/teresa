package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
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
	createSuperUserCmd.Flags().String("name", "admin", "super user name")
	createSuperUserCmd.Flags().String("email", "", "super user email [required]")
	createSuperUserCmd.Flags().String("password", "", "super user password [required]")
}

func createSuperUser(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		log.WithError(err).Fatal("invalid name parameter")
	}

	email, err := cmd.Flags().GetString("email")
	if err != nil {
		log.WithError(err).Fatal("invalid email parameter")
	}

	pass, err := cmd.Flags().GetString("password")
	if err != nil {
		log.WithError(err).Fatal("invalid password parameter")
	}

	if pass == "" {
		client.PrintErrorAndExit("Password required")
		return
	}
	if email == "" {
		client.PrintErrorAndExit("E-mail required")
	}

	db, err := getDB()
	if err != nil {
		client.PrintErrorAndExit("Error on connect to Database: %v", err)
	}

	uOps := user.NewDatabaseOperations(db, auth.NewFake())
	if err := uOps.Create(name, email, pass, true); err != nil {
		client.PrintErrorAndExit("Error on create super user: %s", client.GetErrorMsg(err))
	}
	fmt.Println("Superuser created")
}
