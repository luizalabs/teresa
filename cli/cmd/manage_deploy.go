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

// TODO: create a file like gitignore to upload a package without unecessary files

var createDeployCmd = &cobra.Command{
	Use:   "create APP_FOLDER",
	Short: "deploy an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			lgr.Debug("App folder not provided")
			return newInputError("App Folder not provided")
		}
		return createDeploy(filepath.Clean(args[0]))
	},
}

func createDeploy(appFolder string) error {
	if appFolder == "" {
		lgr.Error("App folder not provided")
		return newSysError("App folder not provided")
	}

	archivePath, err := createTempArchiveToUpload(appFolder)
	if err != nil {
		return err
	}

	c := &http.Client{}
	req := request.NewRequest(c)

	f, _ := os.Open(archivePath)
	req.Files = []request.FileField{
		request.FileField{FieldName: "apparchive", FileName: filepath.Base(archivePath), File: f},
	}

	currentCluster, errGetCurrent := getCurrentCluster()
	if errGetCurrent != nil {
		return errGetCurrent
	}

	req.Headers = map[string]string{
		"Accept":        "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", currentCluster.Token),
	}

	resp, errPost := req.Post(fmt.Sprintf("%s/deploy", currentCluster.Server))
	if errPost != nil {
		lgr.WithError(errPost).Error("Error when uploading an app archive to start a deploy")
		return newSysError("Error when trying to do this action")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		lgr.Debug("User not logged... informing to retry after login")
		return newSysError("You need to login before do this action")
	}

	if resp.StatusCode > 401 && resp.StatusCode <= 500 {
		logFields := logrus.Fields{
			"statusCode": resp.StatusCode,
		}

		if contentBody, errGetContent := resp.Text(); errGetContent == nil {
			logFields["contentBody"] = contentBody
		}

		lgr.WithFields(logFields).Error("Http status diff from 200 when requesting a login")
		return newSysError("Error when trying to do this action")
	}

	return nil
}

// Create a temporary archive file of the app to deploy and return the path of this file
func createTempArchiveToUpload(source string) (archivePath string, err error) {
	id := uuid.NewV1()
	base := filepath.Base(source)
	target := filepath.Join(archiveTempFolder, fmt.Sprintf("%s_%s.tar.gz", base, id))

	if errC := createArchive(source, target); errC != nil {
		return "", errC
	}
	return target, nil
}

func createArchive(source string, target string) error {
	lgr.WithField("dir", source).Debug("Creating archive")

	baseDir := filepath.Dir(source)
	dirStat, errStats := os.Stat(baseDir)
	if errStats != nil {
		lgr.WithError(errStats).WithField("baseDir", baseDir).Error("Dir not found to create an archive")
		return errStats
	} else if !dirStat.IsDir() {
		lgr.WithField("baseDir", baseDir).Error("Path to create the app archive isn't a directory")
		return errors.New("Path to create the app archive isn't a directory")
	}

	tar := new(archivex.TarFile)
	tar.Create(target)
	tar.AddAll(source, true)
	tar.Close()

	return nil
}

func init() {
	// setClusterCmd.Flags().StringVarP(&serverFlag, "server", "s", "", "URI of the server")
	// setClusterCmd.Flags().BoolVarP(&currentFlag, "default", "d", false, "Is the default server")

	deployCmd.AddCommand(createDeployCmd)
}
