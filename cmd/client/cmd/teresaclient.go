package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/luizalabs/teresa-api/client"
	"github.com/luizalabs/teresa-api/client/apps"
	"github.com/luizalabs/teresa-api/client/deployments"
	"github.com/luizalabs/teresa-api/client/teams"
	"github.com/luizalabs/teresa-api/client/users"
	"github.com/luizalabs/teresa-api/models"
	_ "github.com/prometheus/common/log" // still needed?
)

// TeresaClient foo bar
type TeresaClient struct {
	teresa         *apiclient.Teresa
	apiKeyAuthFunc runtime.ClientAuthInfoWriter
}

// TeresaServer scheme and host where the api server is running
type TeresaServer struct {
	scheme string
	host   string
}

// AppInfo foo bar
type AppInfo struct {
	AppID  int64
	TeamID int64
}

var defaultTimeout = 30 * time.Minute

// ParseServerURL parse the server url and ensure its in the format we expect
func ParseServerURL(s string) (TeresaServer, error) {
	u, err := url.Parse(s)
	if err != nil {
		return TeresaServer{}, fmt.Errorf("Failed to parse server: %+v\n", err)
	}
	if u.Scheme == "" || u.Scheme != "http" && u.Scheme != "https" {
		return TeresaServer{}, errors.New("accepted server url format: http(s)://hostname[:port]")
	}
	return TeresaServer{scheme: u.Scheme, host: u.Host}, nil
}

// NewTeresa foo bar
func NewTeresa() TeresaClient {
	cfg, err := readConfigFile(cfgFile)
	if err != nil {
		log.Fatalf("Failed to read config file, err: %+v\n", err)
	}
	n, err := getCurrentClusterName()
	if err != nil {
		log.Fatalf("Failed to get current cluster name, err: %+v\n", err)
	}
	cluster := cfg.Clusters[n]
	suffix := "/v1"

	tc := TeresaClient{teresa: apiclient.Default}
	log.Debugf(`Setting new teresa client. server: %s, api suffix: %s`, cluster.Server, suffix)

	ts, err := ParseServerURL(cluster.Server)
	if err != nil {
		log.Fatal(err)
	}

	client.DefaultTimeout = defaultTimeout
	c := client.New(ts.host, suffix, []string{ts.scheme})

	tc.teresa.SetTransport(c)

	if cluster.Token != "" {
		tc.apiKeyAuthFunc = httptransport.APIKeyAuth("Authorization", "header", cluster.Token)
	}
	return tc
}

// CreateTeam Creates a team
func (tc TeresaClient) CreateTeam(name, email, URL string) (*models.Team, error) {
	params := teams.NewCreateTeamParams()
	e := strfmt.Email(email)
	params.WithBody(&models.Team{Name: &name, Email: e, URL: URL})
	r, err := tc.teresa.Teams.CreateTeam(params, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

// DeleteTeam Deletes a team
func (tc TeresaClient) DeleteTeam(ID int64) error {
	params := teams.NewDeleteTeamParams()
	params.TeamID = ID
	_, err := tc.teresa.Teams.DeleteTeam(params, tc.apiKeyAuthFunc)
	return err
}

// CreateApp creates an user
func (tc TeresaClient) CreateApp(app models.AppIn) (*models.App, error) {
	params := apps.NewCreateAppParams().WithBody(&app)
	r, err := tc.teresa.Apps.CreateApp(params, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

// GetApps return all apps for the token
func (tc TeresaClient) GetApps() (appList []*models.App, err error) {
	p := apps.NewGetAppsParams()
	r, err := tc.teresa.Apps.GetApps(p, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

// GetAppInfo returns all info about the app
func (tc TeresaClient) GetAppInfo(appName string) (app *models.App, err error) {
	p := apps.NewGetAppDetailsParams().WithAppName(appName)
	r, err := tc.teresa.Apps.GetAppDetails(p, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

func (tc TeresaClient) GetAppLogs(appName string, lines *int64, follow *bool, writer io.Writer) error {
	p := apps.NewGetAppLogsParams().WithAppName(appName)
	p.Follow = follow
	p.Lines = lines
	_, err := tc.teresa.Apps.GetAppLogs(p, tc.apiKeyAuthFunc, writer)
	if err != nil {
		return err
	}
	return nil
}

// GetAppDetail Create app attributes
func (tc TeresaClient) GetAppDetail(teamID, appID int64) (app *models.App, err error) {
	// params := apps.NewGetAppDetailsParams().WithTeamID(teamID).WithAppID(appID)
	// r, err := tc.teresa.Apps.GetAppDetails(params, tc.apiKeyAuthFunc)
	// if err != nil {
	// 	return nil, err
	// }
	// return r.Payload, nil
	return
}

// CreateUser Create an user
func (tc TeresaClient) CreateUser(name, email, password string, isAdmin bool) (user *models.User, err error) {
	params := users.NewCreateUserParams()
	params.WithBody(&models.User{Email: &email, Name: &name, Password: &password, IsAdmin: &isAdmin})

	r, err := tc.teresa.Users.CreateUser(params, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

// DeleteUser Delete an user
func (tc TeresaClient) DeleteUser(ID int64) error {
	params := users.NewDeleteUserParams()
	params.UserID = ID
	_, err := tc.teresa.Users.DeleteUser(params, tc.apiKeyAuthFunc)
	return err
}

// GetTeams returns a list with my teams
func (tc TeresaClient) GetTeams() (teamsList []*models.Team, err error) {
	params := teams.NewGetTeamsParams()
	r, err := tc.teresa.Teams.GetTeams(params, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

// CreateDeploy creates a new deploy
func (tc TeresaClient) CreateDeploy(appName, deployDescription string, tarBall *os.File, writer io.Writer) error {
	p := deployments.NewCreateDeploymentParams()
	p.AppName = appName
	p.AppTarball = *tarBall
	p.Description = &deployDescription

	_, err := tc.teresa.Deployments.CreateDeployment(p, tc.apiKeyAuthFunc, writer)
	return err
}

// PartialUpdateApp partial updates app... for now, updates only envvars
func (tc TeresaClient) PartialUpdateApp(appName string, operations []*models.PatchAppRequest) (*models.App, error) {
	p := apps.NewPartialUpdateAppParams()
	p.AppName = appName
	p.Body = operations
	r, err := tc.teresa.Apps.PartialUpdateApp(p, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}

	return r.Payload, nil
}

// AddUserToTeam adds a user (by email) to a team.
// if the user is already part of the team, returns error
func (tc TeresaClient) AddUserToTeam(teamName, userEmail string) (team *models.Team, err error) {
	p := teams.NewAddUserToTeamParams()
	p.TeamName = teamName
	email := strfmt.Email(userEmail)
	p.User.Email = &email
	r, err := tc.teresa.Teams.AddUserToTeam(p, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}
