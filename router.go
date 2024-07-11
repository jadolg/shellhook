package main

import (
	"crypto/subtle"
	"fmt"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net"
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
		scriptToRun, err := c.getScript(r.URL.Query().Get("script"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		remoteIP := getRemoteIP(r)

		cliErr := checkAuthorization(r.Header.Get("Authorization"), scriptToRun, c)
		if cliErr != nil {
			log.WithFields(log.Fields{
				"Error":  cliErr.Message,
				"Client": remoteIP,
			}).Warning("Authorization error")
			http.Error(w, cliErr.Message, cliErr.HTTPCode)
			return
		}

		log.WithFields(log.Fields{
			"ID":         scriptToRun.ID,
			"Path":       scriptToRun.Path,
			"Inline":     scriptToRun.Inline != "",
			"Concurrent": scriptToRun.Concurrent,
			"Shell":      scriptToRun.Shell,
			"User":       scriptToRun.User,
			"Client":     remoteIP,
		}).Info("Executing script")

		if !scriptToRun.Concurrent {
			log.WithFields(log.Fields{"ID": scriptToRun.ID}).Debug("Acquiring lock for script")
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

func getRemoteIP(r *http.Request) string {
	clientIP := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}
	ip, _, splitErr := net.SplitHostPort(clientIP)
	if splitErr == nil {
		clientIP = ip
	}
	return clientIP
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
