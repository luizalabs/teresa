package main

import (
	"log"

	"github.com/luizalabs/teresa-api/pkg/server/cmd"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := cmd.RunCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
