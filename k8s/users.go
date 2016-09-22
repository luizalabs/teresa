package k8s

import (
	log "github.com/Sirupsen/logrus"

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
	GetWithTeams(userEmail string, tk *Token, l *log.Entry) (user *models.User, err error)
	LoadUserToToken(tk *Token, l *log.Entry) error
}

type users struct {
	k *k8sHelper
}

// Token is a web token with all the user info
type Token struct {
	jwt.StandardClaims
	*models.User
}

func newUsers(c *k8sHelper) *users {
	return &users{k: c}
}

// GetWithTeams gets the user with respective teams that he/she is member.
// If the user is admin, return all teams.
// If the user is member of any team, the team will have a flag true to "IAmMember",
// that represents the users is member of the team.
func (c users) getWithTeams(userEmail string, l *log.Entry) (user *models.User, err error) {
	su := storage.User{}
	if storage.DB.Where(&storage.User{Email: userEmail}).Preload("Teams").First(&su).RecordNotFound() {
		l.WithField("user", userEmail).Debug("user not found")
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
		allTeams, _ := c.k.Teams().ListAll(l)
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

// GetWithTeams get the user with teams, checking first token permissions
func (c users) GetWithTeams(userEmail string, tk *Token, l *log.Entry) (user *models.User, err error) {
	// only admins can see other users of the system
	if *tk.IsAdmin == false && *tk.Email != userEmail {
		l.WithField("user", userEmail).Debug("user token is not allowed to see another user")
		return nil, NewUnauthorizedError("token is not allowed to see this user")
	}
	user, err = c.getWithTeams(userEmail, l)
	if err != nil {
		return nil, err
	}
	return
}

// LoadUserToToken load the user inside the token
func (c users) LoadUserToToken(tk *Token, l *log.Entry) error {
	u, err := c.getWithTeams(*tk.Email, l)
	if err != nil {
		return err
	}
	tk.User = u
	return nil
}

// IsAuthorized checks if the user (from token) is authorized to do something with
// the team passed as param
func (t *Token) IsAuthorized(teamName string) bool {
	for _, team := range t.Teams {
		if *team.Name == teamName {
			return true
		}
	}
	return false
}

// IToToken executes a type assertion from the interface to a Token
func IToToken(i interface{}) *Token {
	return i.(*Token)
}
