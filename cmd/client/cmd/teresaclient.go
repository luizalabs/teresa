package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	httptransport "github.com/go-openapi/runtime/client"
	apiclient "github.com/luizalabs/teresa-api/client"
	"github.com/luizalabs/teresa-api/client/apps"
	"github.com/luizalabs/teresa-api/client/teams"
	"github.com/luizalabs/teresa-api/models"
	cli "github.com/luizalabs/teresa-api/pkg/client"
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
	cluster, _ := cli.GetConfig(cfgFile)
	suffix := "/v1"

	tc := TeresaClient{teresa: apiclient.Default}

	ts, _ := ParseServerURL(cluster.Server)

	client.DefaultTimeout = defaultTimeout
	c := client.New(ts.host, suffix, []string{ts.scheme})

	tc.teresa.SetTransport(c)

	if cluster.Token != "" {
		tc.apiKeyAuthFunc = httptransport.APIKeyAuth("Authorization", "header", cluster.Token)
	}
	return tc
}

// DeleteTeam Deletes a team
func (tc TeresaClient) DeleteTeam(ID int64) error {
	params := teams.NewDeleteTeamParams()
	params.TeamID = ID
	_, err := tc.teresa.Teams.DeleteTeam(params, tc.apiKeyAuthFunc)
	return err
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
