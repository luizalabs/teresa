package cmd

import (
	"fmt"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/howeyc/gopass"
	apiclient "github.com/luizalabs/paas/api/client"
	"github.com/luizalabs/paas/api/client/auth"
	"github.com/luizalabs/paas/api/models"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login in the current cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userNameFlag == "" {
			return newInputError("User must be provided")
		}
		fmt.Printf("Password: ")
		p, err := gopass.GetPasswdMasked()
		if err != nil {
			if err != gopass.ErrInterrupted {
				log.WithError(err).Error("Error trying to get the user password")
			}
			return nil
		}
		if err := login(userNameFlag, string(p)); err != nil {
			log.Fatalf("Failed to login: %s\n", err)
		}
		return nil
	},
}

// login action to the command
func login(u string, p string) error {
	s := apiclient.Default
	c := client.New("localhost:8080", "/v1", []string{"http"}) // FIXME
	s.SetTransport(c)
	params := auth.NewUserLoginParams()
	l := models.Login{}
	email := strfmt.Email("arnaldo@luizalabs.com")
	password := strfmt.Password("foobarfoobar")
	l.Email = &email
	l.Password = &password
	params.WithBody(&l)

	r, err := s.Auth.UserLogin(params)
	if err != nil {
		return err
	}
	log.Printf("login token: %s\n", r.Payload.Token)

	cfg, err := readConfigFile(cfgFile)
	if err != nil {
		return err
	}
	n, err := getCurrentClusterName()
	if err != nil {
		return err
	}
	// update the token...
	cluster := cfg.Clusters[n]
	cluster.Token = r.Payload.Token
	cfg.Clusters[n] = cluster
	// write the config file
	err = writeConfigFile(cfgFile, cfg)
	if err != nil {
		return err
	}
	log.WithField("cluster", cluster).Debug("Token added to the cluster")
	return nil
}

// TODO: refactory to reuse the request

func init() {
	loginCmd.Flags().StringVar(&userNameFlag, "user", "", "username to login with")
	configCmd.AddCommand(loginCmd)
}
