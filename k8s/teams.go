package k8s

import (
	"log"

	strfmt "github.com/go-openapi/strfmt"
	"github.com/luizalabs/tapi/models"
	"github.com/luizalabs/tapi/models/storage"
)

// TeamsInterface is used to allow mock testing
type TeamsInterface interface {
	Teams() TeamInterface
}

// TeamInterface is used to interact with Kubernetes and also to allow mock testing
type TeamInterface interface {
	List() (teams []*models.Team, err error)
}

type teams struct {
	k *k8sHelper
}

func newTeams(c *k8sHelper) *teams {
	return &teams{k: c}
}

func (c teams) List() (teams []*models.Team, err error) {
	sts := []*storage.Team{}
	if storage.DB.Find(&sts).RecordNotFound() {
		log.Print("there are no teams registered")
		return nil, NewNotFoundError("there are no teams registered")
	}
	for _, st := range sts {
		t := &models.Team{}
		t.Name = &st.Name
		t.URL = st.URL
		t.Email = strfmt.Email(st.Email)
		teams = append(teams, t)
	}
	return
}
