package cmd

import (
	"fmt"

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
func NewTeresa(s, p, suffix, authToken string) TeresaClient {
	tc := TeresaClient{teresa: apiclient.Default}
	log.Infof(`Setting new teresa client. "TERESA_SERVER":"%s", "TERESA_SERVER_PORT":"%s", "TERESA_API_SUFFIX":"%s"`, s, p, suffix)
	c := client.New(fmt.Sprintf("%s:%s", s, p), suffix, []string{"http"})
	tc.teresa.SetTransport(c)

	if authToken != "" {
		tc.apiKeyAuthFunc = httptransport.APIKeyAuth("Authorization", "header", authToken)
	}
	return tc
}

// Login foo bar
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
