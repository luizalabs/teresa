package storage

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // blank import sqlite
	"golang.org/x/crypto/bcrypt"
)

// DB object to access connection poll
var DB *gorm.DB

type BaseModel struct {
	ID        uint      `gorm:"primary_key;"`
	CreatedAt time.Time `gorm:"not null;"`
	UpdatedAt time.Time `gorm:"not null;"`
}

type Team struct {
	BaseModel
	Name  string        `gorm:"size:128;not null;unique_index;"`
	Email string        `gorm:"size:64;"`
	URL   string        `gorm:"size:1024;"`
	Users []User        `gorm:"many2many:teams_users;"`
	Apps  []Application `gorm:"ForeignKey:TeamID"`
}

type User struct {
	BaseModel
	Name     string `gorm:"size:128;not null;unique_index;"`
	Email    string `gorm:"size:64;not null;unique_index;"`
	Password string `gorm:"size:60;not null;"`
	Teams    []Team `gorm:"many2many:teams_users;"`
}

// Application ...
type Application struct {
	BaseModel
	Name        string `gorm:"size(128);not null;unique_index:idx_application_team_unique_key;"`
	Scale       int16  `gorm:"not null"`
	Addresses   []AppAddress
	Deployments []Deployment `gorm:"ForeignKey:AppID"`
	EnvVars     []EnvVar     `gorm:"ForeignKey:AppID"`
	Team        Team
	TeamID      uint `gorm:"unique_index:idx_application_team_unique_key;"`
}

// EnvVar ...
type EnvVar struct {
	BaseModel
	Key   string `gorm:"size(64);unique_index:idx_envvar_unique_key"`
	Value string `gorm:"size(1024)"`
	AppID uint   `gorm:"unique_index:idx_envvar_unique_key;"`
}

// AppAddress ...
type AppAddress struct {
	BaseModel
	Address string `orm:"unique;size(1024)"`
}

type deploymentOrigin string

// DeploymentOrigin const
const (
	CliAppDeploy deploymentOrigin = "cli_app_deploy"
	GIT          deploymentOrigin = "git"
	CI           deploymentOrigin = "ci"
)

// Deployment ...
type Deployment struct {
	BaseModel
	UUID        string           `gorm:"size(36)"`
	Description string           `gorm:"size(1024)"`
	Origin      deploymentOrigin `gorm:"size(14);index"`
	Error       string           `gorm:"size(2048)"`
	AppID       uint
}

// Authenticate ...
func (u *User) Authenticate(p *string) (err error) {
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(*p))
	if err == nil {
		log.Printf("Authentication succeeded for user [%s]\n", u.Email)
	} else {
		log.Printf("Authentication failed for user [%s]\n", u.Email)
	}
	return err
}

func init() {
	var err error
	DB, err = gorm.Open("sqlite3", "teresa.sqlite") // FIXME: make it configurable and per-env
	if err != nil {
		panic("failed to connect database")
	}
	// DB.DB().SetMaxIdleConns(10)
	// DB.DB().SetMaxOpenConns(50)
	// Print log.
	DB.LogMode(true)

	// only create, never change
	DB.AutoMigrate(&Team{}, &User{}, &Application{}, &EnvVar{}, &AppAddress{}, &Deployment{})

}
