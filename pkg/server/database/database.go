package database

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
)

// Duplicated for now
type Config struct {
	Hostname string
	Port     int
	Username string
	Password string
	Database string
}

func New(conf Config) (*gorm.DB, error) {
	dialect := "sqlite"
	uri := "teresa.sqlite"
	if conf.Hostname != "" {
		dialect = "mysql"
		uri = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true",
			conf.Username, conf.Password,
			conf.Hostname, conf.Port, conf.Database,
		)
	} else if conf.Database != "" {
		uri = conf.Database
	}

	log.Printf("Using %s to connect to %s", dialect, conf.Database)
	db, err := gorm.Open(dialect, uri)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)
	return db, nil
}
