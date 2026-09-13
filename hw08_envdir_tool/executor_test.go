package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmdReturnsCommandExitCode(t *testing.T) {
	exitCode := RunCmd([]string{"sh", "-c", "exit 42"}, nil)
	require.Equal(t, 42, exitCode)
}

func TestRunCmdAppliesEnv(t *testing.T) {
	t.Setenv("NEED_REMOVE", "old_value")

	env := Environment{
		"NEED_REMOVE": {NeedRemove: true},
		"NEW_VAR":     {Value: "new_value"},
	}

	shellCheck := `[ -z "$NEED_REMOVE" ] && [ "$NEW_VAR" = "new_value" ]`
	exitCode := RunCmd([]string{"sh", "-c", shellCheck}, env)
	require.Equal(t, 0, exitCode)
}

func TestRunCmdWithEmptyCommand(t *testing.T) {
	exitCode := RunCmd(nil, nil)
	require.Equal(t, 1, exitCode)
}
