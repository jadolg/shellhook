package main

import (
	"os"
	"os/exec"
	"strings"
)

func getEnvironmentVariables(username string) ([]string, error) {
	cmd := exec.Command("sudo", "-Hiu", username, "env")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	envVars := strings.Split(string(output), "\n")
	return envVars, nil
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
