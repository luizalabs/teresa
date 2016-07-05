package handlers

import (
	"fmt"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/teams"
)

// CreateTeamHandler ...
func CreateTeamHandler(params teams.CreateTeamParams, principal interface{}) middleware.Responder {
	t := models.Team{
		Name:  params.Body.Name,
		Email: params.Body.Email,
		URL:   params.Body.URL,
	}
	st := storage.Team{
		Name:  *params.Body.Name,
		Email: params.Body.Email.String(),
		URL:   params.Body.URL,
	}

	if err := storage.DB.Create(&st).Error; err != nil {
		log.Printf("CreateTeamHandler failed: %s\n", err)
		return teams.NewCreateTeamDefault(422)
	}
	t.ID = int64(st.ID)
	r := teams.NewCreateTeamCreated()
	r.SetPayload(&t)
	return r
}

// GetTeamDetailsHandler ...
func GetTeamDetailsHandler(params teams.GetTeamDetailParams, principal interface{}) middleware.Responder {
	st := storage.Team{}
	st.ID = uint(params.TeamID)

	if storage.DB.First(&st).RecordNotFound() {
		return teams.NewGetTeamsNotFound()
	}
	fmt.Printf("Found team with ID [%d] name [%s] email [%s]\n", st.ID, st.Name, st.Email)
	r := teams.NewGetTeamDetailOK()
	t := models.Team{
		ID:    int64(st.ID),
		Name:  &st.Name,
		Email: strfmt.Email(st.Email),
	}
	r.SetPayload(&t)
	return r
}

// GetTeamsHandler ...
func GetTeamsHandler(params teams.GetTeamsParams, principal interface{}) middleware.Responder {
	var sts []*storage.Team

	d := storage.DB.Find(&sts)

	if err := d.Error; err != nil {
		log.Printf("ERROR querying teams: %s", err)
		return teams.NewGetTeamsDefault(500)
	}
	if d.RecordNotFound() {
		return teams.NewGetTeamsOK()
	}
	rts := make([]*models.Team, len(sts))
	for i := range sts {
		t := models.Team{
			Name:  &sts[i].Name,
			Email: strfmt.Email(sts[i].Email),
			URL:   sts[i].URL,
		}
		rts[i] = &t
	}

	payload := teams.GetTeamsOKBodyBody{Items: rts}
	r := teams.NewGetTeamsOK()
	r.SetPayload(payload)
	return r
}

// UpdateTeamHandler ...
func UpdateTeamHandler(params teams.UpdateTeamParams, principal interface{}) middleware.Responder {
	st := storage.Team{}
	st.ID = uint(params.TeamID)

	if d := storage.DB.First(&st); d.Error != nil || d.RecordNotFound() {
		return teams.NewGetTeamsDefault(500)
	}
	if params.Body.Name != nil {
		st.Name = *params.Body.Name
	}
	if params.Body.URL != "" {
		st.URL = params.Body.URL
	}

	if err := storage.DB.Save(&st).Error; err != nil {
		log.Printf("ERROR updating team, err: %s", err)
		return teams.NewGetTeamsDefault(500)
	}
	r := teams.NewGetTeamDetailOK()
	t := models.Team{
		ID:    int64(st.ID),
		Name:  &st.Name,
		Email: strfmt.Email(st.Email),
		URL:   st.URL,
	}
	r.SetPayload(&t)
	return r
}

// DeleteTeamHandler ...
func DeleteTeamHandler(params teams.DeleteTeamParams, principal interface{}) middleware.Responder {
	st := storage.Team{}
	st.ID = uint(params.TeamID)

	if storage.DB.Delete(&st).Error != nil {
		return teams.NewGetTeamsDefault(500)
	}

	return teams.NewDeleteTeamNoContent()
}
