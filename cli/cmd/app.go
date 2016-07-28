// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"strings"

	"github.com/luizalabs/paas/api/models"
	"github.com/spf13/cobra"
)

// createAppCmd represents the app command
var createAppCmd = &cobra.Command{
	Use:   "app [app_name]",
	Short: "Create an app",
	Long: `Create a new application for the team.

The application name is always required, but team name is only required if you are part of more than one.
Example:

	$ teresa create app my_app_name

or

	$ teresa create app my_app_name --team my_team

You can also provide in how many pods you want your app running.
Like in the example bellow, let's run in 4 pods:

	$ teresa create app my_app_name --team my_team --scale 4
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return newInputError("app name is required")
		}
		if appScaleFlag == 0 {
			return newInputError("at least one replica is required")
		}

		tc := NewTeresa()
		teamID := tc.GetTeamID(teamNameFlag)
		app, err := tc.CreateApp(args[0], int64(appScaleFlag), teamID)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("App created. Name: %s Replicas: %d", *app.Name, *app.Scale)

		return nil
	},
}

var getAppCmd = &cobra.Command{
	Use:   "app",
	Short: "Get an app",
	Long:  `Get an app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appNameFlag == "" {
			log.Debug("App name not provided")
			return newInputError("App not provided")
		}
		tc := NewTeresa()
		a := tc.GetAppInfo(teamNameFlag, appNameFlag)

		app, err := tc.GetAppDetail(a.TeamID, a.AppID)
		if err != nil {
			log.Fatal(err)
		}
		o := fmt.Sprintf("App: %s\n", *app.Name)
		o = o + fmt.Sprintf("Scale: %d\n", *app.Scale)
		if len(app.AddressList) > 0 {
			o = o + "Address:\n"
			for _, x := range app.AddressList {
				o = o + fmt.Sprintf("  %s\n", x)
			}
		}
		fmt.Printf(o)
		return nil
	},
}

var setEnvVarCmd = &cobra.Command{
	Use:   "env [KEY=value, ...]",
	Short: "Set env vars for the app",
	Long: `Create or update environment variables for the app.

You can add a new environment variable for the app, or update if it already exists.

WARNING:
	If you need to set more than one env var to the application, provide all at once.
	Every time this command is called, the application needs to be updated.

To add an new env var called "FOO":

	$ teresa set env FOO=bar --app my_app --team my_team

You can also provide more than one env var at a time.

	$ teresa set env FOO=bar BAR=foo --app my_app --team my_team

The application name is always required.
The team name is only required if you are part of more than one.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appNameFlag == "" {
			log.Debug("App name not provided")
			return newInputError("App not provided")
		}
		// checking for env vars
		if len(args) == 0 {
			log.Debug("Env vars not provided")
			return newInputError("Env vars not provided")
		}
		// parse args to env vars
		evars := make([]*models.PatchAppEnvVar, len(args))
		for i, s := range args {
			x := strings.SplitN(s, "=", 2)
			if len(x) != 2 {
				return newInputError("Env vars must be in the format FOO=bar")
			}
			e := models.PatchAppEnvVar{
				Key:   &x[0],
				Value: x[1],
			}
			evars[i] = &e
		}

		action := "add"
		path := "/envvars"
		op := models.PatchAppRequest{
			Op:    &action,
			Path:  &path,
			Value: evars,
		}

		tc := NewTeresa()
		// FIXME: change this to return error if any
		a := tc.GetAppInfo(teamNameFlag, appNameFlag)

		// partial update envvars... jsonpatch
		ops := []*models.PatchAppRequest{&op}
		err := tc.PartialUpdateApp(a.TeamID, a.AppID, ops)
		if err != nil {
			log.Fatal(err)
		}
		log.Info("App env vars updated successfully")
		return nil
	},
}

var unsetEnvVarCmd = &cobra.Command{
	Use:   "env [var, ...]",
	Short: "Unset env vars from the app",
	Long: `Unset env vars from the app.

You can remove one or more environment variables from the application.

WARNING:
	If you need to unset more than one env var from the application, provide all at once.
	Every time this command is called, the application needs to be updated.

To unset an env var called "FOO":

	$ teresa unset env FOO --app my_app --team my_team

You can also provide more than one env var at a time.

	$ teresa unset env FOO BAR --app my_app --team my_team

The application name is always required.
The team name is only required if you are part of more than one.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appNameFlag == "" {
			log.Debug("App name not provided")
			return newInputError("App not provided")
		}
		// checking for env vars
		if len(args) == 0 {
			log.Debug("Env vars not provided")
			return newInputError("Env vars not provided")
		}
		// parse args to env vars
		evars := make([]*models.PatchAppEnvVar, len(args))
		for i, k := range args {
			key := k
			e := models.PatchAppEnvVar{
				Key: &key,
			}
			evars[i] = &e
		}

		action := "remove"
		path := "/envvars"
		op := models.PatchAppRequest{
			Op:    &action,
			Path:  &path,
			Value: evars,
		}
		tc := NewTeresa()
		// FIXME: change this to return error if any
		a := tc.GetAppInfo(teamNameFlag, appNameFlag)

		// partial update envvars... jsonpatch
		ops := []*models.PatchAppRequest{&op}
		err := tc.PartialUpdateApp(a.TeamID, a.AppID, ops)
		if err != nil {
			log.Fatal(err)
		}
		log.Info("App env var(s) removed successfully")
		return nil
	},
}

func init() {
	createCmd.AddCommand(createAppCmd)
	createAppCmd.Flags().StringVar(&teamNameFlag, "team", "", "team name")
	createAppCmd.Flags().IntVar(&appScaleFlag, "scale", 1, "replicas")

	getCmd.AddCommand(getAppCmd)
	getAppCmd.Flags().StringVar(&appNameFlag, "app", "", "app name [required]")
	getAppCmd.Flags().StringVar(&teamNameFlag, "team", "", "team name")

	setCmd.AddCommand(setEnvVarCmd)
	setEnvVarCmd.Flags().StringVar(&appNameFlag, "app", "", "app name [required]")
	setEnvVarCmd.Flags().StringVar(&teamNameFlag, "team", "", "team name")

	unsetCmd.AddCommand(unsetEnvVarCmd)
	unsetEnvVarCmd.Flags().StringVar(&appNameFlag, "app", "", "app name [required]")
	unsetEnvVarCmd.Flags().StringVar(&teamNameFlag, "team", "", "team name")
}
