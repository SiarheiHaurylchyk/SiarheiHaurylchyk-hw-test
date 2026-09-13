package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	env, err := ReadDir("testdata/env")
	require.NoError(t, err)

	expectedEnv := Environment{
		"BAR":   {Value: "bar"},
		"EMPTY": {Value: ""},
		"FOO":   {Value: "   foo\nwith new line"},
		"HELLO": {Value: `"hello"`},
		"UNSET": {NeedRemove: true},
	}
	require.Equal(t, expectedEnv, env)
}

func TestReadDirReturnsErrorForMissingDir(t *testing.T) {
	_, err := ReadDir("testdata/no-such-dir")
	require.Error(t, err)
}
