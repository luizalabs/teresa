package main

import (
	"io/ioutil"
	"log"

	"github.com/luizalabs/teresa-api/cmd/client/cmd"

	"google.golang.org/grpc/grpclog"
)

func init() {
	grpclog.SetLogger(log.New(ioutil.Discard, "", 0))
}

func main() {
	cmd.Execute()
}
