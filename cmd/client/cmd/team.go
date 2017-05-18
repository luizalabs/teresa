package cmd

import (
	"fmt"
	"os"

	context "golang.org/x/net/context"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/luizalabs/teresa-api/cmd/client/connection"
	"github.com/luizalabs/teresa-api/pkg/client"
	teampb "github.com/luizalabs/teresa-api/pkg/protobuf/team"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Everything about teams",
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	RunE: func(cmd *cobra.Command, args []string) error {
		teams, err := NewTeresa().GetTeams()
		if err != nil {
			return nil
		}
		fmt.Println("Teams:")
		for _, t := range teams {
			if t.IAmMember {
				fmt.Printf("  - %s (member)\n", *t.Name)
			} else {
				fmt.Printf("  - %s\n", *t.Name)
			}
			if t.Email != "" {
				fmt.Printf("    contact: %s\n", t.Email)
			}
			if t.URL != "" {
				fmt.Printf("    url: %s\n", t.URL)
			}
		}
		return nil
	},
}

var teamCreateCmd = &cobra.Command{
	Use:     "create <team-name>",
	Short:   "Create a team",
	Long:    "Create a team that can have many applications",
	Example: "$ teresa team create foo --email foo@foodomain.com --url http://site.foodomain.com",
	Run:     createTeam,
}

//
//
// // delete team
// var deleteTeamCmd = &cobra.Command{
// 	Use:   "team",
// 	Short: "Delete a team",
// 	Long:  `Delete a team`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if teamIDFlag == 0 {
// 			Fatalf(cmd, "team ID is required")
// 		}
// 		if err := NewTeresa().DeleteTeam(teamIDFlag); err != nil {
// 			log.Fatalf("Failed to delete team: %s", err)
// 		}
// 		log.Infof("Team deleted.")
// 	},
// }
//
//
var teamAddUserCmd = &cobra.Command{
	Use:   "add-user",
	Short: "Add a member to a team",
	Long: `Add a member to a team.

You can add a new user as a member of a team with:

  $ teresa team add-user --user john.doe@foodomain.com --team foo

You need to create a user before use this command.`,
	Run: teamAddUser,
}

func init() {
	RootCmd.AddCommand(teamCmd)
	// Commands
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamCreateCmd)
	teamCmd.AddCommand(teamAddUserCmd)

	teamCreateCmd.Flags().String("email", "", "team email, if any")
	teamCreateCmd.Flags().String("url", "", "team site's URL, if any")

	teamAddUserCmd.Flags().String("user", "", "user email")
	teamAddUserCmd.Flags().String("team", "", "team name")

}

func createTeam(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		return
	}
	name := args[0]
	email, _ := cmd.Flags().GetString("email")
	url, _ := cmd.Flags().GetString("url")

	conn, err := connection.New(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro connecting to server:", err)
		return
	}
	defer conn.Close()

	cli := teampb.NewTeamClient(conn)
	req := &teampb.CreateRequest{Name: name, Email: email, Url: url}
	if _, err := cli.Create(context.Background(), req); err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}
	fmt.Println("Team created with success")
}

func teamAddUser(cmd *cobra.Command, args []string) {
	team, _ := cmd.Flags().GetString("team")
	user, _ := cmd.Flags().GetString("user")
	if team == "" || user == "" {
		cmd.Usage()
		return
	}

	conn, err := connection.New(cfgFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro connecting to server:", err)
		return
	}
	defer conn.Close()

	cli := teampb.NewTeamClient(conn)
	req := &teampb.AddUserRequest{Name: team, Email: user}
	if _, err := cli.AddUser(context.Background(), req); err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}
	fmt.Printf("User %s is now member of the team %s\n", color.CyanString(user), color.CyanString(team))
}
