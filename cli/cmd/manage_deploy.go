package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/jhoonb/archivex"
	"github.com/mozillazg/request"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

// TODO: create a file like gitignore to upload a package without unecessary files, or get from app config?!?!?

var createDeployCmd = &cobra.Command{
	Use:   "create APP_FOLDER",
	Short: "deploy an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := ""
		if len(args) > 0 {
			p = args[0]
		}
		return createDeploy(appNameFlag, p)
	},
}

func createDeploy(appName, appFolder string) error {
	if appName == "" {
		log.Debug("App name not provided")
		return newInputError("App name not provided")
	}
	if appFolder == "" {
		log.Error("App folder not provided")
		return newSysError("App folder not provided")
	}

	// requesting `me` to get team and app id to proceed
	var teamID, appID int64
	tc := NewTeresa()
	me, _ := tc.Me()
	// FIXME: check if user is in more than 1 team
	// if len(me.Teams) == 1 {
	// }
	teamID = me.Teams[0].ID
	for _, a := range me.Teams[0].Apps {
		if *a.Name == appName {
			appID = a.ID
			break
		}
	}
	if teamID == 0 || appID == 0 {
		log.Debug("teamID or appID not found")
		return newInputError("Invalid team or app.")
	}

	tar, err := createTempArchiveToUpload(appFolder)
	if err != nil {
		return err
	}

	h := &http.Client{}
	req := request.NewRequest(h)
	file, _ := os.Open(tar)
	req.Files = []request.FileField{
		request.FileField{
			FieldName: "appTarball",
			FileName:  filepath.Base(tar),
			File:      file},
	}
	cluster, err := getCurrentCluster()
	if err != nil {
		return err
	}
	req.Headers = map[string]string{
		"Accept":        "application/json",
		"Authorization": cluster.Token,
	}

	// FIXME: we need to receive this from the cli
	req.Params = map[string]string{
		"description": "put something here",
	}

	resp, err := req.Post(fmt.Sprintf("%s/v1/teams/%d/apps/%d/deployments", cluster.Server, teamID, appID))
	if err != nil {
		log.WithError(err).Error("Error when uploading an app archive to start a deploy")
		return newSysError("Error when trying to do this action")
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		log.Debug("User not logged... informing to retry after login")
		return newSysError("You need to login before do this action")
	}
	if resp.StatusCode > 401 && resp.StatusCode <= 500 {
		fields := logrus.Fields{
			"statusCode": resp.StatusCode,
		}
		if body, err := resp.Text(); err == nil {
			fields["contentBody"] = body
		}
		log.WithFields(fields).Error("Http status diff from 200 when requesting a login")
		return newSysError("Error when trying to do this action")
	}
	fmt.Println("Deploy created with success")
	return nil
}

// create a temporary archive file of the app to deploy and return the path of this file
func createTempArchiveToUpload(source string) (path string, err error) {
	id := uuid.NewV1()
	base := filepath.Base(source)
	path = filepath.Join(archiveTempFolder, fmt.Sprintf("%s_%s.tar.gz", base, id))
	if err = createArchive(source, path); err != nil {
		return "", err
	}
	return
}

// create an archive of the source folder
func createArchive(source string, target string) error {
	log.WithField("dir", source).Debug("Creating archive")
	base := filepath.Dir(source)
	dir, err := os.Stat(base)
	if err != nil {
		log.WithError(err).WithField("baseDir", base).Error("Dir not found to create an archive")
		return err
	} else if !dir.IsDir() {
		log.WithField("baseDir", base).Error("Path to create the app archive isn't a directory")
		return errors.New("Path to create the app archive isn't a directory")
	}
	tar := new(archivex.TarFile)
	tar.Create(target)
	tar.AddAll(source, false)
	tar.Close()
	return nil
}

func init() {
	createDeployCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name [required]")
	deployCmd.AddCommand(createDeployCmd)
}
