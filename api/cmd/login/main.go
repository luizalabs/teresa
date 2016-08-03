package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/luizalabs/teresa/api/client"
	"github.com/luizalabs/teresa/api/client/apps"
	"github.com/luizalabs/teresa/api/client/auth"
	"github.com/luizalabs/teresa/api/client/users"
	"github.com/luizalabs/teresa/api/models"
)

// AuthToken is the jwt token
var AuthToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFybmFsZG9AbHVpemFsYWJzLmNvbSIsImV4cCI6MTQ2ODk0MzcwN30.N-KOBlwGFpGxah9Y82SgBNyH6noAXDpJP4rRB7HWBtpPFOnQrccGNf64Euk3c4bzvOjjvr5jPnoYcJIqLyoFCduXZXRayxo65z49zlQiNMX7mRNtR7eqRwr0Bv_4SVLh0t3VMPMcX9fUzgGyRJkQdUqVLlU8AYntKr2STxypkbxHfeb1IkvdDuxfoiMl4WntgbxxTckFEpA9TDAzHyvK8N4BaRKe-BArCh5Qe0a3XpZFOJb_RPKEn-_XmbXsVWPmfnWQ2RjKla04VsPoHL-cuDmz-rgmZPKlRoAS3EoAw_33uY3GDRLNyLy8AadDMPwp1HNsyeGE9EUHcX7jY1KsteWIcXEIwMBtaCnQeyhxPO-U_qPydprdFwaBLAMOp0SQ1cNcsLunMnv3L4RzX6On7ngWyvsRTz2dLswZjUma-0hRqjvx8xDZG_wul1ISCXNzZ1j-mcmQZrw1KWK5n_5tUCH8owVzo1aV3eATX2Mn6NNvNk6qvXmtjNfvLDr7xma_pcSbQXCnDY-cyaOUbbbOSCH-hQYh5bPpPOm1EVZHlKV8zuV17lfcdsA7UXNJ4g2Xhes62HB3ZZNy-yZM9T2K2IFv2IGf-2CWuWMt2uk9ATsQz2aCStaFigAD0twsw8qrI8jhxfs9gY2B96OhqoYCd80SPekCdLSIA4n9g_BNsB0"

// TeresaClient foo bar
type TeresaClient struct {
	teresa         *apiclient.Teresa
	apiKeyAuthFunc runtime.ClientAuthInfoWriter
}

// NewTeresa foo bar
func NewTeresa(s, p, suffix string) TeresaClient {
	tc := TeresaClient{teresa: apiclient.Default}
	log.Printf(`Setting new teresa client. "TERESA_SERVER":"%s", "TERESA_SERVER_PORT":"%s", "TERESA_API_SUFFIX":"%s"`, s, p, suffix)
	c := client.New(fmt.Sprintf("%s:%s", s, p), suffix, []string{"http"})
	tc.teresa.SetTransport(c)

	if AuthToken != "" {
		tc.apiKeyAuthFunc = httptransport.APIKeyAuth("Authorization", "header", AuthToken)
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
	log.Printf("r: %+v\n", r)
	if err != nil {
		return nil, err
	}
	return r.Payload, nil
}

func main() {
	teresa := NewTeresa(os.Getenv("TERESA_SERVER"), os.Getenv("TERESA_SERVER_PORT"), os.Getenv("TERESA_API_SUFFIX"))

	// login
	l, err := teresa.Login(strfmt.Email("arnaldo@luizalabs.com"), strfmt.Password("foobarfoobar"))
	log.Printf("Login: %+v err: %+v\n\n", l, err)

	// create user
	/*
		u, err := teresa.CreateUser("arnaldo+3@luizalabs.com", "arnaldo+2", "lerolero")
		log.Printf("User: %+v, err: %+v\n\n", u, err)
	*/

	// create app
	a, err := teresa.CreateApp("myapp", 1)
	log.Printf("App: %+v, err: %+v\n", a, err)
}

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
	log.Printf("token: %#v err: %#v\n", t, err)
}
*/
