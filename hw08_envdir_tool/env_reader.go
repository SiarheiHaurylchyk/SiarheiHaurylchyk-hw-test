package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

type EnvValue struct {
	Value      string
	NeedRemove bool
}

func ReadDir(dir string) (Environment, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	environment := make(Environment, len(dirEntries))
	for _, dirEntry := range dirEntries {
		envVarName := dirEntry.Name()
		if dirEntry.IsDir() || strings.Contains(envVarName, "=") {
			continue
		}

		envValue, err := readEnvValueFromFile(filepath.Join(dir, envVarName))
		if err != nil {
			return nil, err
		}

		environment[envVarName] = envValue
	}

	return environment, nil
}

func readEnvValueFromFile(filePath string) (EnvValue, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return EnvValue{}, err
	}

	if len(fileContent) == 0 {
		return EnvValue{NeedRemove: true}, nil
	}

	firstLine := fileContent
	if newLineIndex := bytes.IndexByte(fileContent, '\n'); newLineIndex >= 0 {
		firstLine = fileContent[:newLineIndex]
	}
	firstLine = bytes.TrimRight(firstLine, "\r")
	firstLine = bytes.ReplaceAll(firstLine, []byte{0x00}, []byte{'\n'})

	value := strings.TrimRight(string(firstLine), " \t")

	return EnvValue{Value: value}, nil
}
