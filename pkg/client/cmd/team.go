package cmd

import (
	"fmt"

	context "golang.org/x/net/context"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/luizalabs/teresa/pkg/client"
	"github.com/luizalabs/teresa/pkg/client/connection"
	teampb "github.com/luizalabs/teresa/pkg/protobuf/team"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Everything about teams",
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	Run:   teamList,
}

var teamCreateCmd = &cobra.Command{
	Use:     "create <team-name>",
	Short:   "Create a team",
	Long:    "Create a team that can have many applications",
	Example: "$ teresa team create foo --email foo@foodomain.com --url http://site.foodomain.com",
	Run:     createTeam,
}

var teamAddUserCmd = &cobra.Command{
	Use:   "add-user",
	Short: "Add a member to a team",
	Long: `Add a member to a team.

You can add a new user as a member of a team with:

  $ teresa team add-user --user john.doe@foodomain.com --team foo

You need to create a user before use this command.`,
	Run: teamAddUser,
}

var teamRemoveUserCmd = &cobra.Command{
	Use:   "remove-user",
	Short: "Remove a member of a team",
	Long: `Remove a member of a team.

You can remove an user of a team with:

  $ teresa team remove-user --user john.doe@foodomain.com --team foo
`,
	Run: teamRemoveUser,
}

func init() {
	RootCmd.AddCommand(teamCmd)
	// Commands
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamCreateCmd)
	teamCmd.AddCommand(teamAddUserCmd)
	teamCmd.AddCommand(teamRemoveUserCmd)

	teamListCmd.Flags().Bool("show-users", false, "show members of team")

	teamCreateCmd.Flags().String("email", "", "team email, if any")
	teamCreateCmd.Flags().String("url", "", "team site's URL, if any")

	teamAddUserCmd.Flags().String("user", "", "user email")
	teamAddUserCmd.Flags().String("team", "", "team name")

	teamRemoveUserCmd.Flags().String("user", "", "user email")
	teamRemoveUserCmd.Flags().String("team", "", "team name")
}

func createTeam(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		return
	}
	name := args[0]
	email, _ := cmd.Flags().GetString("email")
	url, _ := cmd.Flags().GetString("url")

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := teampb.NewTeamClient(conn)
	req := &teampb.CreateRequest{Name: name, Email: email, Url: url}
	if _, err := cli.Create(context.Background(), req); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	fmt.Println("Team created with success")
}

func teamAddUser(cmd *cobra.Command, args []string) {
	team, err := cmd.Flags().GetString("team")
	if err != nil {
		client.PrintErrorAndExit("Invalid team parameter: %v", err)
	}
	user, err := cmd.Flags().GetString("user")
	if err != nil {
		client.PrintErrorAndExit("Invalid user parameter: %v", err)
	}
	if team == "" || user == "" {
		cmd.Usage()
		return
	}

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := teampb.NewTeamClient(conn)
	req := &teampb.AddUserRequest{Name: team, User: user}
	if _, err := cli.AddUser(context.Background(), req); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}
	fmt.Printf("User %s is now member of the team %s\n", color.CyanString(user), color.CyanString(team))
}

func teamList(cmd *cobra.Command, args []string) {
	showUsers, _ := cmd.Flags().GetBool("show-users")

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := teampb.NewTeamClient(conn)
	resp, err := cli.List(context.Background(), &teampb.Empty{})
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	if len(resp.Teams) == 0 {
		fmt.Println("You do not belong to any team")
		return
	}

	fmt.Println("Teams:")
	for _, t := range resp.Teams {
		fmt.Print(color.CyanString(t.Name))
		for _, s := range []string{t.Email, t.Url} {
			if s != "" {
				fmt.Printf(" - %s", s)
			}
		}
		fmt.Print("\n")
		if !showUsers {
			continue
		}
		for _, u := range t.Users {
			fmt.Printf("- %s (%s)\n", u.Name, u.Email)
		}
	}
}

func teamRemoveUser(cmd *cobra.Command, args []string) {
	team, err := cmd.Flags().GetString("team")
	if err != nil {
		client.PrintErrorAndExit("Invalid team parameter")
	}

	user, err := cmd.Flags().GetString("user")
	if err != nil {
		client.PrintErrorAndExit("Invalid user parameter")
	}

	if team == "" || user == "" {
		cmd.Usage()
		return
	}

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	cli := teampb.NewTeamClient(conn)
	req := &teampb.RemoveUserRequest{Team: team, User: user}
	if _, err := cli.RemoveUser(context.Background(), req); err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	fmt.Printf("User %s has been removed from the team %s\n", color.CyanString(user), color.CyanString(team))
}
