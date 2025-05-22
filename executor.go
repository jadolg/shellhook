package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func executeScript(scriptToRun script, globalEnvironment []environment) ([]byte, error) {
	shell := getShell(scriptToRun)
	scriptPath := scriptToRun.Path

	if scriptToRun.Inline != "" {
		tempScript, err := createTemporaryScriptFromInline(scriptToRun)
		if err != nil {
			return nil, err
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
			return nil, fmt.Errorf("%v for %s", err, scriptToRun.User)
		}
	}

	injectEnvironmentVariables(scriptToRun.Environment, globalEnvironment, cmd)

	output, err := cmd.Output()
	execsTotal.Inc()
	if err != nil {
		return nil, fmt.Errorf("%s%v", output, err)
	}
	log.WithFields(log.Fields{"output": string(output), "script": scriptPath}).Debug("Script output")
	return output, nil
}

func injectEnvironmentVariables(scriptEnvironment []environment, globalEnvironment []environment, cmd *exec.Cmd) {
	for _, env := range globalEnvironment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", env.Key, env.Value))
	}

	for _, env := range scriptEnvironment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", env.Key, env.Value))
	}
}

func getUser(scriptToRun script) string {
	if scriptToRun.User != "" {
		return scriptToRun.User
	}

	currentUser, err := user.Current()
	if err != nil {
		log.Errorf("Error fetching current user: %v. Using default (root)", err)
		return "root"
	}
	return currentUser.Username
}

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
			username := getUser(scriptToRun)
			shell = detectDefaultShell(username, "/etc/passwd")
		} else {
			shell = shellFromEnv
		}
	}
	return shell
}

func injectUserInCmd(username string, cmd *exec.Cmd) error {
	u, err := user.Lookup(username)
	if err != nil {
		return err
	}
	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	if err != nil {
		return err
	}
	gid, err := strconv.ParseInt(u.Gid, 10, 32)
	if err != nil {
		return err
	}
	groups, err := u.GroupIds()
	if err != nil {
		return err
	}
	groupIDs := make([]uint32, len(groups))
	for i, group := range groups {
		gid, err := strconv.ParseInt(group, 10, 32)
		if err != nil {
			return err
		}
		groupIDs[i] = uint32(gid)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	envForUser, err := getEnvironmentVariables(username)
	if err != nil {
		return err
	}

	cmd.Env = append(cmd.Env, envForUser...)
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid), Groups: groupIDs}
	return nil
}

func detectDefaultShell(user, etcpasswdPath string) string {
	file, err := os.Open(etcpasswdPath)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", etcpasswdPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) >= 7 && fields[0] == user {
			return fields[6]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading %s: %v", etcpasswdPath, err)
	}

	return "/bin/bash"
}
