package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
	"net/http"
	"os"
)

func main() {
	var port int
	var configFile string
	var logLevel string

	flag.IntVar(&port, "port", 9081, "Port to listen on")
	flag.StringVar(&configFile, "config", "./config.yaml", "Path to config file")
	flag.StringVar(&logLevel, "loglevel", "info", "Log level")
	flag.Parse()

	c, err := getConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	err = configureLogs(logLevel)
	if err != nil {
		log.Fatal(err)
	}

	router := getRouter(c)
	log.Printf("Starting server on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func configureLogs(logLevel string) error {
	parsedLogLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	log.SetLevel(parsedLogLevel)
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		log.SetFormatter(&log.JSONFormatter{})
	}
	return err
}
