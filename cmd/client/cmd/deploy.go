package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/luizalabs/teresa-api/cmd/client/tar"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <app folder>",
	Short: "Deploy an app",
	// 	Long: `Deploy an application.
	//
	// To deploy an app you have to pass it's name, the team the app
	// belongs and the path to the source code. You might want to
	// describe your deployments through --description, as that'll
	// eventually help on rollbacks.
	//
	// eg.:
	//
	//   $ teresa deploy . --app webapi --team site --description "release 1.2 with new checkout"
	// 	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster, err := getCurrentClusterName()
		if err != nil {
			return newCmdError("You have to select a cluster first, check the config help: teresa config")
		}
		if len(args) != 1 {
			return newUsageError("You should provide the app folder in order to continue")
		}
		appFolder := args[0]
		appName, _ := cmd.Flags().GetString("app")
		deployDescription, _ := cmd.Flags().GetString("description")
		// showing warning message to the user
		fmt.Printf("Deploying app %s to the cluster %s...\n", color.CyanString(`"%s"`, appName), color.YellowString(`"%s"`, cluster))
		noinput, _ := cmd.Flags().GetBool("no-input")
		if !noinput {
			fmt.Print("Are you sure? (yes/NO)? ")
			// Waiting for the user answer...
			s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if s = strings.ToLower(strings.TrimRight(s, "\r\n")); s != "yes" {
				return nil
			}
		}

		// create and get the archive
		fmt.Println("Generating tarball of:", appFolder)
		tarPath, err := createTempArchiveToUpload(appName, appFolder)
		if err != nil {
			// TODO: check what happen when the app folder is not valid
			return err
		}
		file, err := os.Open(*tarPath)
		if err != nil {
			return err
		}
		defer file.Close()

		tc := NewTeresa()
		err = tc.CreateDeploy(appName, deployDescription, file, os.Stdout)
		if err != nil {
			return err
		}
		return nil
	},
}

// // Writer to be used on deployment, as Write() is very specific and
// // should be implemented some other way -- moving out the deployment
// // error checking from it's Write method.
// type deploymentWriter struct {
// 	w io.Writer
// }
//
// // Write the buffer out to logger, return an error when the string
// // `----------deployment-error----------` is found on the buffer.
// func (tw *deploymentWriter) Write(p []byte) (n int, err error) {
// 	s := strings.Replace(string(p), deploymentErrorMark, "", -1)
// 	s = strings.Replace(s, deploymentSuccessMark, "", -1)
// 	// log.Info(strings.Trim(fmt.Sprintf("%s", s), "\n"))
// 	if strings.Contains(string(p), deploymentErrorMark) {
// 		return len(p), errors.New("Deploy failed")
// 	}
// 	return len(p), nil
// }

// create a temporary archive file of the app to deploy and return the path of this file
func createTempArchiveToUpload(appName, source string) (path *string, err error) {
	id := uuid.NewV4()
	source, err = filepath.Abs(source)
	if err != nil {
		return nil, err
	}
	p := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s.tar.gz", appName, id))
	if err = createArchive(source, p); err != nil {
		return nil, err
	}
	return &p, nil
}

// create an archive of the source folder
func createArchive(source, target string) error {
	dir, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("Dir not found to create an archive. %s", err)
	} else if !dir.IsDir() {
		return errors.New("Path to create the app archive isn't a directory")
	}

	ignorePatterns, err := getIgnorePatterns(source)
	if err != nil {
		return errors.New("Invalid file '.teresaignore'")
	}

	t, err := tar.New(target)
	if err != nil {
		return err
	}
	defer t.Close()

	if ignorePatterns != nil {
		if err = addFiles(source, t, ignorePatterns); err != nil {
			return err
		}
	} else {
		t.AddAll(source)
	}
	return nil
}

func getIgnorePatterns(source string) ([]string, error) {
	fPath := filepath.Join(source, ".teresaignore")
	if _, err := os.Stat(fPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	file, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	patterns := make([]string, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if text := scanner.Text(); text != "" {
			patterns = append(patterns, text)
		}
	}

	if len(patterns) == 0 {
		return nil, nil
	}

	return patterns, nil
}

func addFiles(source string, tar tar.Writer, ignorePatterns []string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, ip := range ignorePatterns {
			if matched, _ := filepath.Match(ip, info.Name()); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		basePath := fmt.Sprintf("%s%c", source, filepath.Separator)
		filename := strings.Replace(path, basePath, "", 1)
		return tar.AddFile(path, filename)
	})
}

func init() {
	RootCmd.AddCommand(deployCmd)

	deployCmd.Flags().String("app", "", "app name (required)")
	deployCmd.Flags().String("description", "", "deploy description (required)")
	deployCmd.Flags().Bool("no-input", false, "deploy app without warning")

}
