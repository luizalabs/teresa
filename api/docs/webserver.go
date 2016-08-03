package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	// receive the swagger spec path from the command line
	if len(os.Args) != 2 {
		log.Fatal("Missing swagger file as argument")
	}
	swaggerfile := os.Args[1]

	// listen on a free port specified by the OS
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("listening on: %s", listener.Addr().String())

	// serve the whole content of docs/ on '/'
	fs := http.FileServer(http.Dir("docs"))
	http.Handle("/", fs)

	// swagger.yml is accessible through /swagger.yml
	http.HandleFunc("/swagger.yml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.WriteHeader(http.StatusOK)
		data, err := ioutil.ReadFile(swaggerfile)
		if err != nil {
			log.Fatal(err.Error())
		}
		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		fmt.Fprint(w, string(data))
	})

	url := listener.Addr().String()
	log.Printf("Opening %s...", url)
	cmd := exec.Command("open", fmt.Sprintf("http://%s", url))
	if err := cmd.Start(); err != nil {
		log.Fatal(err.Error())
	}

	panic(http.Serve(listener, nil))
}
