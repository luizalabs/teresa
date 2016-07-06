package main

import (
	"fmt"
	"log"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/luizalabs/paas/api/client"
	"github.com/luizalabs/paas/api/client/teams"
	"github.com/luizalabs/paas/api/models"
)

func AuthenticateRequest(req runtime.ClientRequest, reg strfmt.Registry) error {
	fmt.Printf("AuthenticateRequest()-------------------\n")
	return nil
}

type ClientAuthInfoWriter interface {
	AuthenticateRequest(runtime.ClientRequest, strfmt.Registry) error
}

func main() {
	s := apiclient.Default
	c := client.New("localhost:62616", "/v1", []string{"http"})
	s.SetTransport(c)

	var authinfo ClientAuthInfoWriter

	//p.WithBody(
	/*
		p := auth.NewUserLoginParams()
		l := models.Login{}
		email := strfmt.Email("arnaldo@luizalabs.com")
		password := strfmt.Password("foobarfoobar")
		l.Email = &email
		l.Password = &password
		p.WithBody(&l)

		r, err := s.Auth.UserLogin(p)
		if err != nil {
			log.Printf("login token: %s\n", r.Payload.Token)
		}
		log.Printf("err: %#v\n", err)
	*/

	tp := teams.NewCreateTeamParams()
	payload := models.Team{}
	name := "mobileeeeeeeeeee"
	eemail := "mobile@luizalabs.com"
	url := "mobile.luizalabs.com"
	payload.Email = strfmt.Email(eemail)
	payload.Name = &name
	payload.URL = url
	tp.WithBody(&payload)

	tr, err := s.Teams.CreateTeam(tp, authinfo)
	if err != nil {
		log.Printf("created team. response: %#v\n", tr)
	} else {
		log.Printf("failed to create team, err: %#v\n", err)
	}

	//fmt.Printf("teams: %#v\n", s.Teams)
	//teamClient := s.Teams.Client

	// make the request to get all items
	/*
		resp, err := apiclient.Default.Operations.All(operations.AllParams{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%#v\n", resp.Payload)
	*/
}
