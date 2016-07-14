package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jhoonb/archivex"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var createDeployCmd = &cobra.Command{
	Use:   "deploy APP_FOLDER",
	Short: "deploy an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if appNameFlag == "" {
			log.Debug("App name not provided")
			return newInputError("App not provided")
		}
		if len(args) == 0 || (len(args) > 0 && args[0] == "") {
			log.Debug("App folder not provided")
			return newInputError("App folder not provided")
		}
		return createDeploy(appNameFlag, teamNameFlag, descriptionFlag, args[0])
	},
}

func createDeploy(appName, teamName, description, appFolder string) error {
	tc := NewTeresa()
	a := tc.GetAppInfo(teamName, appName)
	tc.GetAppInfo(teamName, appName)
	// create and get the archive
	tar, err := createTempArchiveToUpload(appName, teamName, appFolder)
	if err != nil {
		log.Fatalf("error creating the archive. %s", err)
	}
	file, err := os.Open(tar)
	if err != nil {
		log.Fatalf("error getting the archive to upload. %s", err)
	}
	defer file.Close()
	_, err = tc.CreateDeploy(a.TeamID, a.AppID, description, file)
	if err != nil {
		log.Fatalf("error creating the deploy. %s", err)
	}
	log.Infoln("Deploy created with success")
	return nil
}

// create a temporary archive file of the app to deploy and return the path of this file
func createTempArchiveToUpload(app, team, source string) (path string, err error) {
	id := uuid.NewV4()
	source, err = filepath.Abs(source)
	if err != nil {
		return "", err
	}
	path = filepath.Join(archiveTempFolder, fmt.Sprintf("%s_%s_%s.tar.gz", team, app, id))
	if err = createArchive(source, path); err != nil {
		return "", err
	}
	return
}

// create an archive of the source folder
func createArchive(source string, target string) error {
	log.WithField("dir", source).Debug("Creating archive")
	dir, err := os.Stat(source)
	if err != nil {
		log.WithError(err).WithField("dir", source).Error("Dir not found to create an archive")
		return err
	} else if !dir.IsDir() {
		log.WithField("dir", source).Error("Path to create the app archive isn't a directory")
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
	createDeployCmd.Flags().StringVarP(&teamNameFlag, "team", "t", "", "team name")
	createDeployCmd.Flags().StringVarP(&descriptionFlag, "description", "d", "", "deploy description")

	createCmd.AddCommand(createDeployCmd)
}
