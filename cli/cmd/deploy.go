package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jhoonb/archivex"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy APP_FOLDER",
	Short: "Deploy an app",
	Long: `Deploy an application.

To deploy an app you have to pass it's name, the team the app
belongs and the path to the source code. You might want to
describe your deployments through --description, as that'll
eventually help on rollbacks.

eg.:

  $ teresa deploy --app webapi --team site --description "release 1.2 with new checkout"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if appNameFlag == "" && len(args) == 0 {
			Usagef(cmd, "")
			return
		}
		if appNameFlag == "" {
			Fatalf(cmd, "app name required")
		}
		if len(args) == 0 || (len(args) > 0 && args[0] == "") {
			Fatalf(cmd, "app folder required")
		}
		createDeploy(appNameFlag, teamNameFlag, descriptionFlag, args[0])
	},
}

func createDeploy(appName, teamName, description, appFolder string) error {
	clusterName, err := getCurrentClusterName()
	if err != nil {
		log.Fatalf("You have to select a cluster first, check the config help: teresa config")
	}

	tc := NewTeresa()
	log.Infof("Getting app info from cluster %s", clusterName)
	a := tc.GetAppInfo(teamName, appName)
	// create and get the archive
	log.Infof("Generating tarball of %s", appFolder)
	tar, err := createTempArchiveToUpload(appName, teamName, appFolder)
	if err != nil {
		log.Fatalf("error creating the archive. %s", err)
	}
	file, err := os.Open(tar)
	if err != nil {
		log.Fatalf("error getting the archive to upload. %s", err)
	}
	defer file.Close()

	log.Infof("Deploying application to cluster %s", clusterName)
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Start()
	_, err = tc.CreateDeploy(a.TeamID, a.AppID, description, file)
	s.Stop()
	if err != nil {
		log.Fatalf("error creating the deploy. %s", err)
	}
	log.Infoln("Done")
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
	deployCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name [required]")
	deployCmd.Flags().StringVarP(&teamNameFlag, "team", "t", "", "team name")
	deployCmd.Flags().StringVarP(&descriptionFlag, "description", "d", "", "deploy description")

	RootCmd.AddCommand(deployCmd)
}
