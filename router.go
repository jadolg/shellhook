package main

import (
	"crypto/subtle"
	"fmt"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

func executionHandler(c configuration, locks map[uuid.UUID]*sync.Mutex) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		scriptToRun, done := getScript(w, r, c)
		if done {
			return
		}

		if isUnauthorized(w, r, scriptToRun, c) {
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

		shell := getShell(scriptToRun)
		scriptPath := scriptToRun.Path

		if scriptToRun.Inline != "" {
			tempScript, err := createTemporaryScriptFromInline(scriptToRun)
			if err != nil {
				reportError(err, w)
				return
			}
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					errorsTotal.Inc()
					log.Error(err)
				}
			}(tempScript)

			scriptPath = tempScript
		}

		cmd := exec.Command(shell, scriptPath)
		if scriptToRun.User != "" {
			err := injectUserInCmd(scriptToRun.User, cmd)
			if err != nil {
				reportError(fmt.Errorf("%v for %s", err, scriptToRun.User), w)
				return
			}
		}
		output, err := cmd.Output()
		execsTotal.Inc()
		if err != nil {
			reportError(fmt.Errorf("%s\n%s\n%v", output, err.(*exec.ExitError).Stderr, err), w)
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

func isUnauthorized(w http.ResponseWriter, r *http.Request, scriptToRun script, c configuration) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return true
	}

	if (scriptToRun.Token != "" && subtle.ConstantTimeCompare([]byte(authHeader), []byte(scriptToRun.Token)) != 1) ||
		(scriptToRun.Token == "" && subtle.ConstantTimeCompare([]byte(authHeader), []byte(c.DefaultToken)) != 1) {
		http.Error(w, "Invalid authorization token", http.StatusUnauthorized)
		return true
	}
	return false
}

func getScript(w http.ResponseWriter, r *http.Request, c configuration) (script, bool) {
	scriptID, err := uuid.Parse(r.URL.Query().Get("script"))

	if err != nil {
		http.Error(w, "Missing script parameter or invalid script parameter", http.StatusBadRequest)
		return script{}, true
	}

	for _, ascript := range c.Scripts {
		if ascript.ID == scriptID {
			return ascript, false
		}
	}

	http.Error(w, "Script not found", http.StatusNotFound)
	return script{}, true
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
