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

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Create a user",
	Long:  `Create a user`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if userNameFlag == "" {
			return newInputError("user name is required")
		}
		if userEmailFlag == "" {
			return newInputError("user email is required")
		}
		if userPasswordFlag == "" {
			return newInputError("user password is required")
		}
		tc := NewTeresa()
		user, err := tc.CreateUser(userNameFlag, userEmailFlag, userPasswordFlag)
		if err != nil {
			log.Fatalf("Failed to create user: %s", err)
		}
		log.Infof("User created. Name: %s Email: %s\n", *user.Name, *user.Email)
		return err
	},
}

// delete user
var deleteUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Delete an user",
	Long:  `Delete an user`,
	Run: func(cmd *cobra.Command, args []string) {
		if userIDFlag == 0 {
			Fatalf(cmd, "user name is required")
		}
		if err := NewTeresa().DeleteUser(userIDFlag); err != nil {
			Fatalf(cmd, "Failed to delete user, err: %s\n", err)
		}
		log.Infof("User deleted.")
	},
}

func init() {
	createCmd.AddCommand(userCmd)
	userCmd.Flags().StringVar(&userNameFlag, "name", "", "user name [required]")
	userCmd.Flags().StringVar(&userEmailFlag, "email", "", "user email [required]")
	userCmd.Flags().StringVar(&userPasswordFlag, "password", "", "user password [required]")

	deleteCmd.AddCommand(deleteUserCmd)
	deleteUserCmd.Flags().Int64Var(&userIDFlag, "id", 0, "user ID [required]")
}
