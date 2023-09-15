package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	var port int
	var configFile string
	flag.IntVar(&port, "port", 9081, "Port to listen on")
	flag.StringVar(&configFile, "config", "./config.yaml", "Path to config file")
	flag.Parse()

	c, err := getConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	router := getRouter(c)
	log.Printf("Starting server on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
