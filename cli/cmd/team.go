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
	"github.com/go-openapi/runtime/client"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/luizalabs/paas/api/client"
	"github.com/luizalabs/paas/api/client/teams"
	"github.com/luizalabs/paas/api/models"
	"github.com/spf13/cobra"
)

func createTeam(name, email, URL string) error {
	s := apiclient.Default
	c := client.New("localhost:8080", "/v1", []string{"http"}) // FIXME
	apiKeyHeaderAuth := httptransport.APIKeyAuth("Authorization", "header", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFybmFsZG9AbHVpemFsYWJzLmNvbSIsImV4cCI6MTQ2ODk0MzcwN30.N-KOBlwGFpGxah9Y82SgBNyH6noAXDpJP4rRB7HWBtpPFOnQrccGNf64Euk3c4bzvOjjvr5jPnoYcJIqLyoFCduXZXRayxo65z49zlQiNMX7mRNtR7eqRwr0Bv_4SVLh0t3VMPMcX9fUzgGyRJkQdUqVLlU8AYntKr2STxypkbxHfeb1IkvdDuxfoiMl4WntgbxxTckFEpA9TDAzHyvK8N4BaRKe-BArCh5Qe0a3XpZFOJb_RPKEn-_XmbXsVWPmfnWQ2RjKla04VsPoHL-cuDmz-rgmZPKlRoAS3EoAw_33uY3GDRLNyLy8AadDMPwp1HNsyeGE9EUHcX7jY1KsteWIcXEIwMBtaCnQeyhxPO-U_qPydprdFwaBLAMOp0SQ1cNcsLunMnv3L4RzX6On7ngWyvsRTz2dLswZjUma-0hRqjvx8xDZG_wul1ISCXNzZ1j-mcmQZrw1KWK5n_5tUCH8owVzo1aV3eATX2Mn6NNvNk6qvXmtjNfvLDr7xma_pcSbQXCnDY-cyaOUbbbOSCH-hQYh5bPpPOm1EVZHlKV8zuV17lfcdsA7UXNJ4g2Xhes62HB3ZZNy-yZM9T2K2IFv2IGf-2CWuWMt2uk9ATsQz2aCStaFigAD0twsw8qrI8jhxfs9gY2B96OhqoYCd80SPekCdLSIA4n9g_BNsB0")
	s.SetTransport(c)

	tp := teams.NewCreateTeamParams()
	payload := models.Team{}
	payload.Email = strfmt.Email(email)
	payload.Name = &name
	payload.URL = URL
	tp.WithBody(&payload)

	tr, err := s.Teams.CreateTeam(tp, apiKeyHeaderAuth)
	if err != nil {
		log.Errorf("team not created. response: %#v\n", tr)
		return err
	}
	team := tr.Payload
	log.Infof("Team created. Name: %s Email: %s URL: %s\n", (*team.Name), team.Email, team.URL)
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
