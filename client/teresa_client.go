package client

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/luizalabs/teresa-api/client/apps"
	"github.com/luizalabs/teresa-api/client/deployments"
	"github.com/luizalabs/teresa-api/client/teams"
	"github.com/luizalabs/teresa-api/client/users"
)

// Default teresa HTTP client.
var Default = NewHTTPClient(nil)

// NewHTTPClient creates a new teresa HTTP client.
func NewHTTPClient(formats strfmt.Registry) *Teresa {
	if formats == nil {
		formats = strfmt.Default
	}
	transport := httptransport.New("localhost:8080", "/v1", []string{"http"})
	return New(transport, formats)
}

// New creates a new teresa client
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Teresa {
	cli := new(Teresa)
	cli.Transport = transport

	cli.Apps = apps.New(transport, formats)

	cli.Deployments = deployments.New(transport, formats)

	cli.Teams = teams.New(transport, formats)

	cli.Users = users.New(transport, formats)

	return cli
}

// Teresa is a client for teresa
type Teresa struct {
	Apps *apps.Client

	Deployments *deployments.Client

	Teams *teams.Client

	Users *users.Client

	Transport runtime.ClientTransport
}

// SetTransport changes the transport on the client and all its subresources
func (c *Teresa) SetTransport(transport runtime.ClientTransport) {
	c.Transport = transport

	c.Apps.SetTransport(transport)

	c.Deployments.SetTransport(transport)

	c.Teams.SetTransport(transport)

	c.Users.SetTransport(transport)

}
