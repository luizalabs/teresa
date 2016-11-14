package k8s

import (
	log "github.com/Sirupsen/logrus"

	strfmt "github.com/go-openapi/strfmt"
	"github.com/luizalabs/teresa-api/models"
	"github.com/luizalabs/teresa-api/models/storage"
)

// TeamsInterface is used to allow mock testing
type TeamsInterface interface {
	Teams() TeamInterface
}

// TeamInterface is used to interact with Kubernetes and also to allow mock testing
type TeamInterface interface {
	ListAll(l *log.Entry) (teams []*models.Team, err error)
}

type teams struct {
	k *k8sHelper
}

func newTeams(c *k8sHelper) *teams {
	return &teams{k: c}
}

// ListAll gets all teams, no mather if the user can see that team or not
func (c teams) ListAll(l *log.Entry) (teams []*models.Team, err error) {
	sts := []*storage.Team{}
	if storage.DB.Find(&sts).RecordNotFound() {
		msg := "there are no teams registered"
		log.Debug(msg)
		return nil, NewNotFoundError(msg)
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
