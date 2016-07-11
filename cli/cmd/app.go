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

	"github.com/spf13/cobra"
)

// appCmd represents the app command
var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Create an app",
	Long:  `Create an app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appNameFlag == "" {
			return newInputError("app name is required")
		}
		if appScaleFlag == 0 {
			return newInputError("at least one replica is required")
		}

		tc := NewTeresa()
		app, err := tc.CreateApp(appNameFlag, int64(appScaleFlag))
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("App created. Name: %s Replicas: %s", *app.Name, *app.Scale)

		return nil
	},
}

var getAppCmd = &cobra.Command{
	Use:   "app",
	Short: "Get an app",
	Long:  `Get an app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appNameFlag == "" {
			return newInputError("app name is required")
		}
		tc := NewTeresa()
		me, _ := tc.Me()
		// FIXME: check if user is in more than 1 team
		// FIXME: put this call in only one place (see manage_deploy.go)
		// if len(me.Teams) == 1 {
		// }
		var teamID, appID int64
		teamID = me.Teams[0].ID
		for _, a := range me.Teams[0].Apps {
			if *a.Name == appNameFlag {
				appID = a.ID
				break
			}
		}
		app, err := tc.GetAppDetail(teamID, appID)
		if err != nil {
			log.Fatal(err)
		}

		o := fmt.Sprintf("App: %s\n", *app.Name)
		o = o + fmt.Sprintf("Scale: %d\n", *app.Scale)
		o = o + "Address:\n"
		for _, x := range app.AddressList {
			o = o + fmt.Sprintf("  %s\n", x)
		}
		fmt.Printf(o)
		return nil
	},
}

func init() {
	createCmd.AddCommand(appCmd)

	appCmd.Flags().StringVar(&appNameFlag, "name", "", "app name [required]")
	appCmd.Flags().IntVar(&appScaleFlag, "scale", 1, "replicas [required]")

	getCmd.AddCommand(getAppCmd)
	getAppCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name [required]")

}
