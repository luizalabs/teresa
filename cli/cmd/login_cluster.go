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
		pass, err := gopass.GetPasswdMasked()
		if err != nil {
			if err != gopass.ErrInterrupted {
				lgr.WithError(err).Error("Error trying to get the user password")
			}
			return nil
		}

		return login(userNameFlag, string(pass))
	},
}

func login(user string, password string) error {
	token, errGetToken := getLoginToken(user, password)
	if errGetToken != nil {
		return errGetToken
	}

	conf, errRead := readConfigFile(cfgFile)
	if errRead != nil {
		return errRead
	}

	clusterName, errClusterName := getCurrentCluster()
	if errClusterName != nil {
		return errClusterName
	}

	// Update the token...
	currentCluster := conf.Clusters[clusterName]
	currentCluster.LoginToken = token
	conf.Clusters[clusterName] = currentCluster

	err := writeConfigFile(cfgFile, conf)
	if err != nil {
		return err
	}

	lgr.WithField("currentCluster", currentCluster).Debug("Token added to the cluster")

	return nil
}

// TODO: refactory to reuse the request
func getLoginToken(user string, password string) (token string, err error) {
	if user == "" || password == "" {
		return "", newSysError("Name and password must be provided")
	}

	c := &http.Client{}

	req := request.NewRequest(c)
	req.Data = map[string]string{
		"email":    user,
		"password": password,
	}

	// Check if one of the registered servers are set to use
	b, errB := getCurrentServerBasePath()
	if errB != nil {
		return "", errB
	}

	url := fmt.Sprintf("%s/login", b)

	resp, errPost := req.Post(url)
	if errPost != nil {
		lgr.WithError(errPost).WithField("clusterUrl", url).WithField("user", user).Error("Error when trying to do a login request")
		return "", errPost
	}
	defer resp.Body.Close()

	// TODO: check this with the real api codes
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		lgr.WithField("statusCode", resp.StatusCode).Debug("Http status diff from 200 when requesting a login")
		return "", newSysError("Invalid user or password")
	}

	j, errToJSON := resp.Json()
	if errToJSON != nil {
		lgr.WithError(errToJSON).Error("Error trying to parte the content to json")
		return "", errToJSON
	}

	return j.Get("token").MustString(), nil
}

func init() {
	loginCmd.Flags().StringVar(&userNameFlag, "user", "", "username to login with")

	configCmd.AddCommand(loginCmd)
}
