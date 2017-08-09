package main

import (
	"io/ioutil"
	"log"

	"github.com/luizalabs/teresa-api/pkg/client"
	"github.com/luizalabs/teresa-api/pkg/client/cmd"

	"google.golang.org/grpc/grpclog"
)

func init() {
	grpclog.SetLogger(log.New(ioutil.Discard, "", 0))
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		client.PrintErrorAndExit(err.Error())
	}
}
