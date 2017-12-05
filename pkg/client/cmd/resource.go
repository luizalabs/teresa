package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/luizalabs/teresa/pkg/client"
	"github.com/luizalabs/teresa/pkg/client/connection"
	respb "github.com/luizalabs/teresa/pkg/protobuf/resource"
	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
)

const (
	resNameLimit = 50
)

var resCmd = &cobra.Command{
	Use:   "resource",
	Short: "Everything about resources",
}

var resCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Creates a new resource",
	Long: `Creates a new resource.

  Stay tuned!`,
	Run: createRes,
}

func createRes(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}
	name := args[0]
	if len(name) > resNameLimit {
		client.PrintErrorAndExit("Invalid resource name (max %d characters)", resNameLimit)
	}

	team, err := cmd.Flags().GetString("team")
	if err != nil || team == "" {
		client.PrintErrorAndExit("Invalid team parameter")
	}

	settings, err := cmd.Flags().GetStringSlice("set")
	if err != nil {
		client.PrintErrorAndExit("Invalid set parameter(s)")
	}

	conn, err := connection.New(cfgFile, cfgCluster)
	if err != nil {
		client.PrintErrorAndExit("Error connecting to server: %v", err)
	}
	defer conn.Close()

	parsedSettings, err := parseSettings(settings)
	if err != nil {
		client.PrintErrorAndExit(err.Error())
	}

	req := &respb.CreateRequest{Name: name, TeamName: team, Settings: parsedSettings}
	cli := respb.NewResourceClient(conn)
	text, err := cli.Create(context.Background(), req)
	if err != nil {
		client.PrintErrorAndExit(client.GetErrorMsg(err))
	}

	fmt.Println("Resource created")
	fmt.Println(text)
}

func init() {
	RootCmd.AddCommand(resCmd)
	resCmd.AddCommand(resCreateCmd)

	resCreateCmd.Flags().String("team", "", "team owner of the resource")
	resCreateCmd.Flags().StringSlice("set", []string{}, "resource settings")
}

func parseSettings(args []string) ([]*respb.CreateRequest_Setting, error) {
	s := make([]*respb.CreateRequest_Setting, len(args))

	for i, item := range args {
		tmp := strings.SplitN(item, "=", 2)
		if len(tmp) != 2 {
			return nil, errors.New("Resource settings must be in the format foo=bar")
		}
		s[i] = &respb.CreateRequest_Setting{Key: tmp[0], Value: tmp[1]}
	}

	return s, nil
}
