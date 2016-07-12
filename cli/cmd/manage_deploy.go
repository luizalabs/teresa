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

// TODO: create a file like gitignore to upload a package without unecessary files, or get from app config?!?!?

var createDeployCmd = &cobra.Command{
	Use:   "create APP_FOLDER",
	Short: "deploy an app",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := ""
		if len(args) > 0 {
			p = args[0]
		}
		return createDeploy(appNameFlag, teamNameFlag, p)
	},
}

func createDeploy(appName, teamName, appFolder string) error {
	if appName == "" {
		log.Debug("App name not provided")
		return newInputError("App not provided")
	}
	if appFolder == "" {
		log.Debug("App folder not provided")
		return newSysError("App folder not provided")
	}

	// requesting `me` to get team and app id to proceed
	var teamID, appID int64
	tc := NewTeresa()
	me, err := tc.Me()
	if err != nil {
		log.Fatalf("unable to get user information: %s", err)
	}
	// FIXME: check if this user have access to the specific team (we should centralize this check)
	if len(me.Teams) > 1 && teamName == "" {
		log.Debug("user in more than one team and dont provided a team to do the action")
		return newSysError("Team not provided")
	}
	for _, t := range me.Teams {
		if teamName == "" || *t.Name == teamName {
			teamID = t.ID
			for _, a := range t.Apps {
				appID = a.ID
				break
			}
			break
		}
	}
	if teamID == 0 || appID == 0 {
		log.Debug("teamID or appID not found")
		return newInputError("Invalid team or app to continue")
	}
	// create and get the archive
	tar, err := createTempArchiveToUpload(appFolder)
	if err != nil {
		log.Fatalf("error creating the archive. %s", err)
	}
	file, err := os.Open(tar)
	if err != nil {
		log.Fatalf("error getting the archive to upload. %s", err)
	}

	// FIXME: change this null text for the cli real description
	_, err = tc.CreateDeploy(teamID, appID, "null", file)
	if err != nil {
		log.Fatalf("error creating the deploy. %s", err)
	}

	log.Infoln("Deploy created with success")
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
	// FIXME: add only necessary files to deploy, removing .git and .gitignore files if they exist.
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
	createDeployCmd.Flags().StringVarP(&teamNameFlag, "team", "t", "", "team name")
	deployCmd.AddCommand(createDeployCmd)
}
