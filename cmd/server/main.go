package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/luizalabs/teresa/pkg/server/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("teresa-server exited")
	}
}
