package database

import (
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// BaseModel declares base fields that are to be used on all models
type BaseModel struct {
	ID        uint      `gorm:"primary_key;"`
	CreatedAt time.Time `gorm:"not null;"`
	UpdatedAt time.Time `gorm:"not null;"`
}

// Team represents a team of developers
type Team struct {
	BaseModel
	Name  string `gorm:"size:128;not null;unique_index;"`
	Email string `gorm:"size:64;"`
	URL   string `gorm:"size:1024;"`
	Users []User `gorm:"many2many:teams_users;"`
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
