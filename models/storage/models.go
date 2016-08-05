package storage

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"  // mysql used in production
	_ "github.com/jinzhu/gorm/dialects/sqlite" // used in dev or test
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/crypto/bcrypt"
)

// DB object to access connection poll
var DB *gorm.DB

// BaseModel declares base fields that are to be used on all models
type BaseModel struct {
	ID        uint      `gorm:"primary_key;"`
	CreatedAt time.Time `gorm:"not null;"`
	UpdatedAt time.Time `gorm:"not null;"`
}

// Team represents a team of developers
type Team struct {
	BaseModel
	Name  string        `gorm:"size:128;not null;unique_index;"`
	Email string        `gorm:"size:64;"`
	URL   string        `gorm:"size:1024;"`
	Users []User        `gorm:"many2many:teams_users;"`
	Apps  []Application `gorm:"ForeignKey:TeamID"`
}

// User represents a developer
type User struct {
	BaseModel
	Name     string `gorm:"size:128;not null;unique_index;"`
	Email    string `gorm:"size:64;not null;unique_index;"`
	Password string `gorm:"size:60;not null;"`
	IsAdmin  bool   `gorm:"not null;"`
	Teams    []Team `gorm:"many2many:teams_users;"`
}

// Application represents an application
type Application struct {
	BaseModel
	Name        string       `gorm:"size(128);not null;unique_index:idx_application_team_unique_key;"`
	Scale       int16        `gorm:"not null"`
	Addresses   []AppAddress `gorm:"ForeignKey:AppID"`
	Deployments []Deployment `gorm:"ForeignKey:AppID"`
	EnvVars     []EnvVar     `gorm:"ForeignKey:AppID"`
	Team        Team
	TeamID      uint `gorm:"unique_index:idx_application_team_unique_key;"`
}

// EnvVar represents an application environment variable
type EnvVar struct {
	BaseModel
	Key   string `gorm:"size(64);unique_index:idx_envvar_unique_key"`
	Value string `gorm:"size(1024)"`
	AppID uint   `gorm:"unique_index:idx_envvar_unique_key;"`
}

// AppAddress represents an application fqdn
type AppAddress struct {
	BaseModel
	Address string `orm:"unique;size(1024)"`
	AppID   uint
}

// Deployment ...
type Deployment struct {
	BaseModel
	UUID        string `gorm:"size(36)"`
	Description string `gorm:"size(1024)"`
	// Origin      deploymentOrigin `gorm:"size(14);index"`
	Error string `gorm:"size(2048)"`
	AppID uint
}

// Authenticate check if the user's password matches via bcrypt
func (u *User) Authenticate(p *string) (err error) {
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(*p))
	if err == nil {
		log.Printf("Authentication succeeded for user [%s]\n", u.Email)
	} else {
		log.Printf("Authentication failed for user [%s]\n", u.Email)
	}
	return err
}

// DatabaseConfig holds the database configuration variables
type DatabaseConfig struct {
	Hostname string
	Port     int
	Username string
	Password string
	Database string
}

func init() {
	var err error
	var conf DatabaseConfig
	var dialect, uri string

	// on production, we read the database configuration only from envvars
	env := os.Getenv("TERESA_ENVIRONMENT")

	err = envconfig.Process("teresadb", &conf)
	if env == "PRODUCTION" && err != nil {
		log.Fatalf("Failed to read configuration from environment: %s", err.Error())
	}

	// we got to read the conf from env
	if env == "PRODUCTION" || (err == nil && conf.Hostname != "") {
		dialect = "mysql"
		uri = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", conf.Username, conf.Password, conf.Hostname, conf.Port, conf.Database)
	} else {
		dialect = "sqlite3"
		uri = "teresa.sqlite"
	}
	log.Printf("Using %s to connect to %s", dialect, uri)

	DB, err = gorm.Open(dialect, uri)
	if err != nil {
		log.Fatalf("failed to connect database: %s", err.Error())
	}

	// Print log.
	DB.LogMode(true)

	// only create, never change
	DB.AutoMigrate(&Team{}, &User{}, &Application{}, &EnvVar{}, &AppAddress{}, &Deployment{})
}
