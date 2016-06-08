package cmd

import (
	"fmt"
	"net/http"

	"github.com/howeyc/gopass"
	"github.com/mozillazg/request"
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
		return login(userNameFlag, string(p))
	},
}

// login action to the command
func login(u string, p string) error {
	t, err := getLoginToken(u, p)
	if err != nil {
		return err
	}
	c, err := readConfigFile(cfgFile)
	if err != nil {
		return err
	}
	n, err := getCurrentClusterName()
	if err != nil {
		return err
	}
	// update the token...
	cluster := c.Clusters[n]
	cluster.Token = t
	c.Clusters[n] = cluster
	// write the config file
	err = writeConfigFile(cfgFile, c)
	if err != nil {
		return err
	}
	log.WithField("cluster", cluster).Debug("Token added to the cluster")
	return nil
}

// TODO: refactory to reuse the request
func getLoginToken(u string, p string) (t string, err error) {
	if u == "" || p == "" {
		return "", newSysError("Name and password must be provided")
	}
	h := &http.Client{}
	req := request.NewRequest(h)
	req.Data = map[string]string{
		"email":    u,
		"password": p,
	}
	// check if one of the registered servers are set to use
	c, err := getCurrentCluster()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/login", c.Server)
	resp, err := req.Post(url)
	if err != nil {
		log.WithError(err).WithField("clusterUrl", url).WithField("user", u).Error("Error when trying to do a login request")
		return "", err
	}
	defer resp.Body.Close()
	// TODO: check this with the real api codes
	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		log.WithField("statusCode", resp.StatusCode).Debug("Http status diff from 200 when requesting a login")
		return "", newSysError("Invalid user or password")
	}
	j, err := resp.Json()
	if err != nil {
		log.WithError(err).Error("Error trying to parte the content to json")
		return
	}
	t = j.Get("token").MustString()
	return
}

func init() {
	loginCmd.Flags().StringVar(&userNameFlag, "user", "", "username to login with")
	configCmd.AddCommand(loginCmd)
}
