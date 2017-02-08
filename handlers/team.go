package handlers

import (
	log "github.com/Sirupsen/logrus"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/luizalabs/teresa-api/k8s"
	"github.com/luizalabs/teresa-api/models"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/restapi/operations/teams"
)

// CreateTeamHandler ...
func CreateTeamHandler(params teams.CreateTeamParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	// need to have admin permissions to do this action
	if *tk.IsAdmin == false {
		log.Printf("User [%s] doesn't have permission to create a team", *tk.Email)
		return teams.NewCreateTeamUnauthorized()
	}

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
	// st := storage.Team{}
	// st.ID = uint(params.TeamID)
	//
	// if storage.DB.First(&st).RecordNotFound() {
	// 	return teams.NewGetTeamsNotFound()
	// }
	// fmt.Printf("Found team with ID [%d] name [%s] email [%s]\n", st.ID, st.Name, st.Email)
	// r := teams.NewGetTeamDetailOK()
	// t := models.Team{
	// 	ID:    int64(st.ID),
	// 	Name:  &st.Name,
	// 	Email: strfmt.Email(st.Email),
	// }
	// r.SetPayload(&t)
	// return r
	return middleware.NotImplemented("operation teams.GetTeamDetailsHandler has not yet been implemented")
}

// GetTeamsHandler ...
func GetTeamsHandler(params teams.GetTeamsParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	l := log.WithField("token", *tk.Email)
	if len(tk.Teams) == 0 {
		l.Debug("token can't see any team or the is no team registered")
		return NewNotFoundError("No teams found")
	}
	return teams.NewGetTeamsOK().WithPayload(tk.Teams)
}

// UpdateTeamHandler ...
func UpdateTeamHandler(params teams.UpdateTeamParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	// need to have admin permissions to do this action
	if *tk.IsAdmin == false {
		log.Printf("User [%s] doesn't have permission to update a team", *tk.Email)
		return teams.NewUpdateTeamDefault(401)
	}

	st := storage.Team{}
	st.ID = uint(params.TeamID)

	if d := storage.DB.First(&st); d.Error != nil || d.RecordNotFound() {
		return teams.NewUpdateTeamDefault(500)
	}
	if params.Body.Name != nil {
		st.Name = *params.Body.Name
	}
	if params.Body.URL != "" {
		st.URL = params.Body.URL
	}

	if err := storage.DB.Save(&st).Error; err != nil {
		log.Printf("ERROR updating team, err: %s", err)
		return teams.NewUpdateTeamDefault(500)
	}

	r := teams.NewUpdateTeamOK()
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
	// tk := k8s.IToToken(principal)
	// // need to have admin permissions to do this action
	// if *tk.IsAdmin == false {
	// 	log.Printf("User [%d: %s] doesn't have permission to delete a team", tk.UserID, tk.Email)
	// 	return teams.NewDeleteTeamDefault(401)
	// }
	//
	// st := storage.Team{}
	// st.ID = uint(params.TeamID)
	//
	// if storage.DB.Delete(&st).Error != nil {
	// 	return teams.NewDeleteTeamDefault(500)
	// }

	return teams.NewDeleteTeamNoContent()
}

// AddUserToTeam add user to a specific team
func AddUserToTeam(params teams.AddUserToTeamParams, principal interface{}) middleware.Responder {
	tk := k8s.IToToken(principal)
	// need admin permissions to do this action
	if *tk.IsAdmin == false {
		log.Printf("User [%s] doesn't have permission to include another user to a team", *tk.Email)
		return teams.NewAddUserToTeamDefault(401)
	}

	st := storage.Team{}
	if storage.DB.Where("name = ?", params.TeamName).First(&st).RecordNotFound() {
		p := models.GenericError{Message: "Team doesn't exist"}
		return teams.NewAddUserToTeamDefault(422).WithPayload(&p)
	}

	susers := []storage.User{}
	storage.DB.Model(&st).Association("Users").Find(&susers)
	for _, su := range susers {
		if su.Email == params.User.Email.String() {
			p := models.GenericError{Message: "User is already member of the team"}
			return teams.NewAddUserToTeamDefault(422).WithPayload(&p)
		}
	}
	su := storage.User{}
	if storage.DB.Where("email = ? ", params.User.Email.String()).First(&su).RecordNotFound() {
		p := models.GenericError{Message: "User must be registered before adding him/her to a team"}
		return teams.NewAddUserToTeamDefault(422).WithPayload(&p)
	}
	if err := storage.DB.Model(&st).Association("Users").Append(su).Error; err != nil {
		log.Printf("Error found when trying to add user [%s] to team [%s]: %s", params.User.Email.String(), params.TeamName, err)
		return teams.NewAddUserToTeamDefault(500)
	}
	return teams.NewAddUserToTeamOK()
}
