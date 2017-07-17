package cmd

import (
	"github.com/jinzhu/gorm"
	"github.com/kelseyhightower/envconfig"
	"github.com/luizalabs/teresa-api/pkg/server/database"
)

func getDB() (*gorm.DB, error) {
	dbConf := new(database.Config)
	if err := envconfig.Process("teresa_db", dbConf); err != nil {
		return nil, err
	}
	return database.New(dbConf)
}
