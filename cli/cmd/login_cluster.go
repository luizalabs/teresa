package cmd

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/howeyc/gopass"
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

		tc := NewTeresa()
		token, err := tc.Login(strfmt.Email(userEmailFlag), strfmt.Password(p))
		if err != nil {
			log.Fatalf("Failed to login: %s\n", err)
		}
		log.Infof("Login OK")
		log.Debugf("Auth token: %s\n", token)

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
		cluster.Token = token
		cfg.Clusters[n] = cluster
		// write the config file
		err = writeConfigFile(cfgFile, cfg)
		if err != nil {
			return err
		}
		log.Debugf("Token saved to config")

		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&userNameFlag, "user", "", "username to login with")
	configCmd.AddCommand(loginCmd)
}
