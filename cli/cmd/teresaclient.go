package cmd

import (
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/luizalabs/paas/api/client"
	"github.com/luizalabs/paas/api/client/apps"
	"github.com/luizalabs/paas/api/client/auth"
	"github.com/luizalabs/paas/api/client/users"
	"github.com/luizalabs/paas/api/models"
)

// TeresaClient foo bar
type TeresaClient struct {
	teresa         *apiclient.Teresa
	apiKeyAuthFunc runtime.ClientAuthInfoWriter
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

	// the split is ugly.
	// using unecessary vars for clarity
	ss := strings.Split(cluster.Server, "://")
	scheme := ss[0]
	host := ss[1]

	c := client.New(host, suffix, []string{scheme})
	tc.teresa.SetTransport(c)

	if cluster.Token != "" {
		tc.apiKeyAuthFunc = httptransport.APIKeyAuth("Authorization", "header", cluster.Token)
	}
	return tc
}

// Login login the user
func (tc TeresaClient) Login(email strfmt.Email, password strfmt.Password) (token string, err error) {
	params := auth.NewUserLoginParams()
	params.WithBody(&models.Login{Email: &email, Password: &password})

	r, err := tc.teresa.Auth.UserLogin(params)
	if err != nil {
		return "", err
	}
	return r.Payload.Token, nil
}

// CreateApp creates an user
func (tc TeresaClient) CreateApp(name string, scale int64) (app *models.App, err error) {
	params := apps.NewCreateAppParams()
	params.WithBody(&models.App{Name: &name, Scale: &scale})
	r, err := tc.teresa.Apps.CreateApp(params, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

// GetApp get an app
func (tc TeresaClient) GetApp(teamId, appId int64) (app *models.App, err error) {

	// TODO: implement this
	// params := apps.NewGetAppDetailsParams().WithTeamID(teamId).WithAppID(appId)
	// r, err := tc.teresa.Apps.GetAppDetails(params, tc.apiKeyAuthFunc)
	// if err != nil {
	// 	return nil, err
	// }
	// r.
	//
	// return r.Payload, nil
	return nil, nil
}

// CreateUser ...
func (tc TeresaClient) CreateUser(email, name, password string) (user *models.User, err error) {
	params := users.NewCreateUserParams()
	params.WithBody(&models.User{Email: &email, Name: &name, Password: &password})

	r, err := tc.teresa.Users.CreateUser(params, tc.apiKeyAuthFunc)
	log.Infof("r: %+v\n", r)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

// Me get's the user infos + teams + apps
func (tc TeresaClient) Me() (user *models.User, err error) {
	r, err := tc.teresa.Users.GetCurrentUser(nil, tc.apiKeyAuthFunc)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

/*
func main() {
	teresa := NewTeresa(os.Getenv("TERESA_SERVER"), os.Getenv("TERESA_SERVER_PORT"), os.Getenv("TERESA_API_SUFFIX"))

	// login
	l, err := teresa.Login(strfmt.Email("arnaldo@luizalabs.com"), strfmt.Password("foobarfoobar"))
	log.Infof("Login: %+v err: %+v\n\n", l, err)

	// create user
		//u, err := teresa.CreateUser("arnaldo+3@luizalabs.com", "arnaldo+2", "lerolero")
		//log.Infof("User: %+v, err: %+v\n\n", u, err)

	// create app
	a, err := teresa.CreateApp("myapp", 1)
	log.Infof("App: %+v, err: %+v\n", a, err)
}
*/

/*
func login(email strfmt.Email, password strfmt.Password) (token string, err error) {
	s := apiclient.Default
	c := client.New("localhost:8080", "/v1", []string{"http"}) // FIXME
	s.SetTransport(c)

	params := auth.NewUserLoginParams()
	params.WithBody(&models.Login{Email: &email, Password: &password})

	r, err := s.Auth.UserLogin(params)
	if err != nil {
		return "", err
	}
	return r.Payload.Token, nil
}

func main() {
	t, err := login(strfmt.Email("arnaldo@luizalabs.com"), strfmt.Password("foobarfoobar"))
	log.Infof("token: %#v err: %#v\n", t, err)
}
*/
