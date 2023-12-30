package main

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
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
