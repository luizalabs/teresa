package k8s

import (
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-openapi/strfmt"
	"github.com/luizalabs/tapi/models"
	"github.com/luizalabs/tapi/models/storage"
)

// UsersInterface is used to allow mock testing
type UsersInterface interface {
	Users() UserInterface
}

// UserInterface is used to interact with Kubernetes and also to allow mock testing
type UserInterface interface {
	LoadUserToToken(tk *Token) error
}

type users struct {
	k *k8sHelper
}

type Token struct {
	jwt.StandardClaims
	*models.User
}

func newUsers(c *k8sHelper) *users {
	return &users{k: c}
}

// GetWithTeams gets the user with respective teams that he/she is member
// If the user is admin, return all teams with the flag "IAmMember", that represents if the users is member of the team
func (c users) getWithTeams(userEmail string) (user *models.User, err error) {
	su := storage.User{}
	if storage.DB.Where(&storage.User{Email: userEmail}).Preload("Teams").First(&su).RecordNotFound() {
		log.Printf(`user "%s" not found`, userEmail)
		return nil, NewNotFoundError("user not found")
	}
	user = &models.User{}
	user.Email = &su.Email
	user.IsAdmin = &su.IsAdmin
	user.Name = &su.Name
	for _, st := range su.Teams {
		t := &models.Team{}
		t.Name = &st.Name
		t.URL = st.URL
		t.Email = strfmt.Email(st.Email)
		t.IAmMember = true
		user.Teams = append(user.Teams, t)
	}
	// admin users can see all teams
	if *user.IsAdmin {
		allTeams, _ := c.k.Teams().List()
		for _, t := range allTeams {
			found := false
			for _, td := range user.Teams {
				if *td.Name == *t.Name {
					found = true
					break
				}
			}
			if found == false {
				user.Teams = append(user.Teams, t)
			}
		}
	}
	return
}

// LoadUserToToken load the user inside the token
func (c users) LoadUserToToken(tk *Token) error {
	u, err := c.getWithTeams(*tk.Email)
	if err != nil {
		return err
	}
	tk.User = u
	return nil
}

func (t *Token) IsAuthorized(teamName string) bool {
	for _, t := range t.Teams {
		if *t.Name == teamName {
			return true
		}
	}
	log.Printf(`token "%s" is not allowed to do actions with the team "%s"`, *t.Email, teamName)
	return false
}

// IToToken executes a type assertion from the interface to a Token
func IToToken(i interface{}) *Token {
	return i.(*Token)
}
