package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	childProcess := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec // запуск произвольной команды
	childProcess.Stdin = os.Stdin
	childProcess.Stdout = os.Stdout
	childProcess.Stderr = os.Stderr
	childProcess.Env = buildChildProcessEnv(env)

	err := childProcess.Run()
	if err == nil {
		return 0
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return 1
}

func buildChildProcessEnv(env Environment) []string {
	childProcessEnv := make([]string, 0, len(os.Environ())+len(env))

	for _, osEnvVar := range os.Environ() {
		envVarName, _, _ := strings.Cut(osEnvVar, "=")
		if _, isOverridden := env[envVarName]; !isOverridden {
			childProcessEnv = append(childProcessEnv, osEnvVar)
		}
	}

	for envVarName, envValue := range env {
		if !envValue.NeedRemove {
			childProcessEnv = append(childProcessEnv, envVarName+"="+envValue.Value)
		}
	}

	return childProcessEnv
}
