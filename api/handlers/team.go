package handlers

import (
	"fmt"
	"log"

	"github.com/astaxie/beego/orm"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/teams"
)

func CreateTeamHandler(params teams.CreateTeamParams, principal interface{}) middleware.Responder {
	o := orm.NewOrm()
	o.Using("default")
	t := models.Team{
		Name:  params.Body.Name,
		Email: params.Body.Email,
		URL:   params.Body.URL,
	}
	st := storage.Team{
		Name:  *params.Body.Name,
		Email: params.Body.Email.String(),
		Url:   params.Body.URL,
	}
	id, err := o.Insert(&st)
	if err != nil {
		log.Printf("CreateTeamHandler failed: %s\n", err)
		return teams.NewCreateTeamDefault(422)
	}
	t.ID = id
	r := teams.NewCreateTeamCreated()
	r.SetPayload(&t)
	return r
}

func GetTeamDetailsHandler(params teams.GetTeamDetailParams, principal interface{}) middleware.Responder {
	o := orm.NewOrm()
	o.Using("default")
	team := storage.Team{Id: params.TeamID}
	err := o.Read(&team)
	if err == orm.ErrNoRows {
		return teams.NewGetTeamsNotFound()
	} else {
		fmt.Printf("Found team with ID [%d] name [%s] email [%s]\n", team.Id, team.Name, team.Email)
		r := teams.NewGetTeamDetailOK()
		t := models.Team{
			ID:    team.Id,
			Name:  &team.Name,
			Email: strfmt.Email(team.Email),
		}
		r.SetPayload(&t)
		return r
	}
}

func GetTeamsHandler(params teams.GetTeamsParams, principal interface{}) middleware.Responder {
	o := orm.NewOrm()
	o.Using("default")

	var sts []*storage.Team
	num, err := o.QueryTable("team").All(&sts)
	if err != nil {
		log.Printf("ERROR querying teams: %s", err)
		return teams.NewGetTeamsDefault(500)
	}
	if num == 0 {
		return teams.NewGetTeamsOK()
	}

	rts := make([]*models.Team, len(sts))
	for i := range sts {
		t := models.Team{
			Name:  &sts[i].Name,
			Email: strfmt.Email(sts[i].Email),
			URL:   sts[i].Url,
		}
		rts[i] = &t
	}

	payload := teams.GetTeamsOKBodyBody{Items: rts}
	r := teams.NewGetTeamsOK()
	r.SetPayload(payload)
	return r
}

func UpdateTeamHandler(params teams.UpdateTeamParams, principal interface{}) middleware.Responder {
	o := orm.NewOrm()
	o.Using("default")

	team := storage.Team{Id: params.TeamID}
	err := o.Read(&team)
	if err != nil {
		return teams.NewGetTeamsDefault(500)
	}
	if params.Body.Name != nil {
		team.Name = *params.Body.Name
	}
	if params.Body.URL != "" {
		team.Url = params.Body.URL
	}
	if _, err := o.Update(&team); err != nil {
		log.Printf("ERROR updating team, err: %s", err)
		return teams.NewGetTeamsDefault(500)
	} else {
		r := teams.NewGetTeamDetailOK()
		t := models.Team{
			ID:    team.Id,
			Name:  &team.Name,
			Email: strfmt.Email(team.Email),
			URL:   team.Url,
		}
		r.SetPayload(&t)
		return r
	}
}

func DeleteTeamHandler(params teams.DeleteTeamParams, principal interface{}) middleware.Responder {
	o := orm.NewOrm()
	o.Using("default")
	if _, err := o.Delete(&storage.Team{Id: params.TeamID}); err != nil {
		return teams.NewGetTeamsDefault(500)
	} else {
		return teams.NewDeleteTeamNoContent()
	}
}
