package cmd

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login in the currently selected cluster",
	Long: `Login in the selected cluster.

eg.:

	$ teresa login --user user@mydomain.com
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if userNameFlag == "" {
			Usage(cmd)
			return
		}
		fmt.Printf("Password: ")
		p, err := gopass.GetPasswdMasked()
		if err != nil {
			if err != gopass.ErrInterrupted {
				log.WithError(err).Error("Error trying to get the user password")
			}
			return
		}

		tc := NewTeresa()
		token, err := tc.Login(strfmt.Email(userNameFlag), strfmt.Password(p))
		if err != nil {
			log.Fatalf("Failed to login: %s\n", err)
		}
		log.Infof("Login OK")
		log.Debugf("Auth token: %s\n", token)
		if err := SetAuthToken(token); err != nil {
			log.Fatalf("Failed to update the auth token: %s\n", err)
		}
	},
}

func init() {
	loginCmd.Flags().StringVar(&userNameFlag, "user", "", "username to login with")
	RootCmd.AddCommand(loginCmd)
}
