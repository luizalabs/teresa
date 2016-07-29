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

import "github.com/spf13/cobra"

// createCmd represents the create command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get app details",
	Long: `Get application details or get teams.

To get details about an application:

	$ teresa get app --app my_app_name --team my_team

To get the teams you belong to:

	$ teresa get teams
	`,
}

func init() {
	RootCmd.AddCommand(getCmd)
}
