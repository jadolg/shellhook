package main

import (
	"crypto/subtle"
	"fmt"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
)

type ClientError struct {
	Message  string `json:"message"`
	HTTPCode int    `json:"code"`
}

func executionHandler(c configuration, locks map[uuid.UUID]*sync.Mutex) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		scriptToRun, cliErr := getScript(r.URL.Query().Get("script"), c)
		if cliErr != nil {
			http.Error(w, cliErr.Message, cliErr.HTTPCode)
			return
		}

		cliErr = checkAuthorization(r.Header.Get("Authorization"), scriptToRun, c)
		if cliErr != nil {
			http.Error(w, cliErr.Message, cliErr.HTTPCode)
			return
		}

		if scriptToRun.Path != "" {
			log.Infof("Executing script: %v with path: '%s'", scriptToRun.ID, scriptToRun.Path)
		} else {
			log.Infof("Executing script: %v with inline", scriptToRun.ID)
		}

		if !scriptToRun.Concurrent {
			log.Debugf("Acquiring lock for script %v", scriptToRun.ID)
			locks[scriptToRun.ID].Lock()
			defer locks[scriptToRun.ID].Unlock()
		}

		output, err := executeScript(scriptToRun, c.Environment)
		if err != nil {
			reportError(err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, "%s", output)
		if err != nil {
			log.Errorf("error responding to request %v", err)
		}
	}
}

func reportError(err error, w http.ResponseWriter) {
	log.Error(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	errorsTotal.Inc()
}

func createTemporaryScriptFromInline(scriptToRun script) (string, error) {
	tempScript, err := os.CreateTemp(os.TempDir(), "*-shellhook")

	if err != nil {
		return "", fmt.Errorf("error creating temporary script file %v for %s", err, scriptToRun.User)
	}

	_, err = tempScript.WriteString(scriptToRun.Inline)
	if err != nil {
		return "", fmt.Errorf("error writing temporary script file %v for %s", err, scriptToRun.User)
	}

	return tempScript.Name(), nil
}

func checkAuthorization(authHeader string, scriptToRun script, c configuration) *ClientError {
	if authHeader == "" {
		return &ClientError{Message: "Missing authorization token", HTTPCode: http.StatusUnauthorized}
	}

	if (scriptToRun.Token != "" && subtle.ConstantTimeCompare([]byte(authHeader), []byte(scriptToRun.Token)) != 1) ||
		(scriptToRun.Token == "" && subtle.ConstantTimeCompare([]byte(authHeader), []byte(c.DefaultToken)) != 1) {
		return &ClientError{Message: "Invalid authorization token", HTTPCode: http.StatusUnauthorized}
	}
	return nil
}

func getScript(scriptUUID string, c configuration) (script, *ClientError) {
	scriptID, err := uuid.Parse(scriptUUID)

	if err != nil {
		return script{}, &ClientError{Message: "Missing script parameter or invalid script parameter", HTTPCode: http.StatusBadRequest}
	}

	for _, ascript := range c.Scripts {
		if ascript.ID == scriptID {
			return ascript, nil
		}
	}

	return script{}, &ClientError{Message: "Script not found", HTTPCode: http.StatusNotFound}
}

func healthcheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "OK")
	if err != nil {
		log.Errorf("error responding to request %v", err)
	}
}

func getRouter(c configuration) *http.ServeMux {
	locks := getLocks(c)
	mux := http.NewServeMux()
	mux.HandleFunc("/hook", executionHandler(c, locks))
	mux.HandleFunc("/health", healthcheckHandler)
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}
