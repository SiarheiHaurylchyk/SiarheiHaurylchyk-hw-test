package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: go-envdir /path/to/env/dir command [args...]")
		os.Exit(1)
	}

	envDirPath := os.Args[1]
	commandWithArgs := os.Args[2:]

	env, err := ReadDir(envDirPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(RunCmd(commandWithArgs, env))
}
