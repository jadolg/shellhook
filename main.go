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
	var configFile, logLevel, certFile, keyFile string

	flag.IntVar(&port, "port", 9081, "Port to listen on")
	flag.StringVar(&configFile, "config", "./config.yaml", "Path to config file (optional)")
	flag.StringVar(&logLevel, "loglevel", "info", "Log level (debug, info, warn, error, fatal, panic)")
	flag.StringVar(&certFile, "cert", "", "Path to TLS certificate file (optional)")
	flag.StringVar(&keyFile, "key", "", "Path to TLS key file (optional)")
	flag.Parse()

	err := configureLogs(logLevel)
	if err != nil {
		log.Fatal(err)
	}

	c, err := getConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if (certFile == "" && keyFile != "") || (certFile != "" && keyFile == "") {
		log.Fatal("Both cert and key must be provided together or left empty.")
	}

	router := getRouter(c)
	if certFile != "" && keyFile != "" {
		log.WithFields(log.Fields{
			"port": port,
			"cert": certFile,
			"key":  keyFile,
		}).Info("Starting TLS server")
		if err := http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, keyFile, router); err != nil {
			log.Fatalf("Error starting TLS server: %v", err)
		}
	} else {
		log.WithFields(log.Fields{
			"port": port,
		}).Info("Starting server")
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
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
