package database

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
)

const (
	maxAttempts    = 20
	defaultDialect = "sqlite3"
	defaultUri     = "teresa.sqlite"
)

type Config struct {
	Hostname string
	Port     int `default:"3306"`
	Username string
	Password string
	Database string
	ShowLogs bool `split_words:"true" default:"false"`
}

func New(conf *Config) (*gorm.DB, error) {
	dialect := defaultDialect
	uri := defaultUri
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

	var db *gorm.DB
	var err error
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		db, err = gorm.Open(dialect, uri)
		if err == nil {
			break
		}
		log.WithFields(log.Fields{
			"dialect": dialect,
			"db":      conf.Database,
			"host":    conf.Hostname,
			"user":    conf.Username,
			"err":     err,
		}).Warnf("failed to connect to database, retrying in %d seconds", attempts)
		time.Sleep(time.Duration(attempts) * time.Second)
	}

	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"dialect": dialect,
		"db":      conf.Database,
		"host":    conf.Hostname,
		"user":    conf.Username,
	}).Info("connected to database")

	db.LogMode(conf.ShowLogs)
	return db, nil
}
