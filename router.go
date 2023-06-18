package main

import (
	"crypto/subtle"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"
)

func executionHandler(c configuration, locks map[uuid.UUID]*sync.Mutex) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		scriptToRun, err, done := getScript(w, r, c)
		if done {
			return
		}

		if checkAuthorization(w, r, scriptToRun, c) {
			return
		}

		log.Debugf("Executing script: %s on with path: %s", scriptToRun.ID, scriptToRun.Path)
		if !scriptToRun.Concurrent {
			locks[scriptToRun.ID].Lock()
			defer locks[scriptToRun.ID].Unlock()
		}

		shell := getShell(scriptToRun)

		cmd := exec.Command(shell, scriptToRun.Path)
		if scriptToRun.User != "" {
			if handleUser(w, scriptToRun, cmd) {
				return
			}
		}
		output, err := cmd.Output()
		if err != nil {
			errorMsg := fmt.Sprintf("%s\n%v", output, err)
			log.Errorf(errorMsg)
			http.Error(w, errorMsg, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, "%s", output)
		if err != nil {
			log.Errorf("error responding to request %v", err)
		}
	}
}

func handleUser(w http.ResponseWriter, scriptToRun script, cmd *exec.Cmd) bool {
	u, err := user.Lookup(scriptToRun.User)
	if err != nil {
		errorMsg := fmt.Sprintf("%v for %s", err, scriptToRun.User)
		log.Error(errorMsg)
		http.Error(w, errorMsg, http.StatusInternalServerError)
	}
	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	gid, err := strconv.ParseInt(u.Gid, 10, 32)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	return false
}

func getShell(scriptToRun script) string {
	shell := scriptToRun.Shell
	if shell == "" {
		shellFromEnv, exists := os.LookupEnv("SHELL")
		if !exists {
			shell = "/bin/bash"
		} else {
			shell = shellFromEnv
		}
	}
	return shell
}

func checkAuthorization(w http.ResponseWriter, r *http.Request, scriptToRun script, c configuration) bool {
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

func getScript(w http.ResponseWriter, r *http.Request, c configuration) (script, error, bool) {
	scriptID, err := uuid.Parse(r.URL.Query().Get("script"))

	if err != nil || scriptID == uuid.Nil {
		http.Error(w, "Missing script parameter or invalid script parameter", http.StatusBadRequest)
		return script{}, nil, true
	}

	scriptToRun := script{}
	for _, ascript := range c.Scripts {
		if ascript.ID == scriptID {
			scriptToRun = ascript
			break
		}
	}

	if scriptToRun.ID == uuid.Nil {
		http.Error(w, "Script not found", http.StatusNotFound)
		return script{}, nil, true
	}
	return scriptToRun, err, false
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
	return mux
}
