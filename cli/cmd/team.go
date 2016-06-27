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
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/luizalabs/paas/api/client"
	"github.com/luizalabs/paas/api/client/teams"
	"github.com/luizalabs/paas/api/models"
	"github.com/spf13/cobra"
)

// AuthenticateRequest authenticates the request -- FIXME: actually authenticate the request
func AuthenticateRequest(req runtime.ClientRequest, reg strfmt.Registry) error {
	log.Debugf("AuthenticateRequest()-------------------\n")
	return nil
}

// ClientAuthInfoWriter authenticates the client
type ClientAuthInfoWriter interface {
	AuthenticateRequest(runtime.ClientRequest, strfmt.Registry) error
}

func createTeam(name, email, URL string) error {
	s := apiclient.Default
	c := client.New("localhost:8080", "/v1", []string{"http"}) // FIXME
	s.SetTransport(c)

	var authinfo ClientAuthInfoWriter

	tp := teams.NewCreateTeamParams()
	payload := models.Team{}
	payload.Email = strfmt.Email(email)
	payload.Name = &name
	payload.URL = URL
	tp.WithBody(&payload)

	tr, err := s.Teams.CreateTeam(tp, authinfo)
	if err != nil {
		log.Debugf("team created. response: %#v\n", tr)
		return err
	}
	return nil
}

// create a team
var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Create a team",
	Long:  `Create a team that is able to have many applications.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if teamNameFlag == "" {
			return newInputError("team name is required")
		}
		log.Debugf("create a team called. name: %s email: %s URL: %s\n", teamNameFlag, teamEmailFlag, teamURLFlag)
		if err := createTeam(teamNameFlag, teamEmailFlag, teamURLFlag); err != nil {
			log.Fatalf("Failed to create team: %s\n", err)
		}
		return nil
	},
}

func init() {
	createCmd.AddCommand(teamCmd)
	teamCmd.Flags().StringVarP(&teamNameFlag, "name", "n", "", "team name [required]")
	teamCmd.Flags().StringVarP(&teamEmailFlag, "email", "e", "", "team email, if any")
	teamCmd.Flags().StringVarP(&teamURLFlag, "url", "u", "", "team site's URL, if any")
}
