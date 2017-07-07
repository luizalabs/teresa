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

func init() {
	RootCmd.AddCommand(teamCmd)
	// Commands
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamCreateCmd)
	teamCmd.AddCommand(teamAddUserCmd)

	teamListCmd.Flags().Bool("show-users", false, "show members of team")

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

	conn, err := connection.New(cfgFile, &connOpts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server:", err)
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
	team, err := cmd.Flags().GetString("team")
	if err != nil {
		fmt.Fprint(os.Stderr, "Invalid team parameter: ", err)
	}
	user, err := cmd.Flags().GetString("user")
	if err != nil {
		fmt.Fprint(os.Stderr, "Invalid user parameter: ", err)
	}
	if team == "" || user == "" {
		cmd.Usage()
		return
	}

	conn, err := connection.New(cfgFile, &connOpts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server:", err)
		return
	}
	defer conn.Close()

	cli := teampb.NewTeamClient(conn)
	req := &teampb.AddUserRequest{Name: team, User: user}
	if _, err := cli.AddUser(context.Background(), req); err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
	}
	fmt.Printf("User %s is now member of the team %s\n", color.CyanString(user), color.CyanString(team))
}

func teamList(cmd *cobra.Command, args []string) {
	showUsers, _ := cmd.Flags().GetBool("show-users")

	conn, err := connection.New(cfgFile, &connOpts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server:", err)
		return
	}
	defer conn.Close()

	cli := teampb.NewTeamClient(conn)
	resp, err := cli.List(context.Background(), &teampb.Empty{})
	if err != nil {
		fmt.Fprintln(os.Stderr, client.GetErrorMsg(err))
		return
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
