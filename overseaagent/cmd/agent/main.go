package main

import (
	"flag"
	"fmt"
	"net/http"

	"overseaagent/pkg/config"
	"overseaagent/pkg/store"
)

var conf *string = flag.String("conf", "./config.json", "path of the configure file.")

func main() {
	flag.Parse()

	c := config.New(*conf)
	mux := store.New(c)

	fmt.Println("overseg agent is running at ", c.Host.Port)
	http.ListenAndServe(c.Host.Port, mux)
}
