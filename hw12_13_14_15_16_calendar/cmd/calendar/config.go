package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger   LoggerConf   `yaml:"logger"`
	Storage  StorageConf  `yaml:"storage"`
	Server   ServerConf   `yaml:"server"`
	GRPC     ServerConf   `yaml:"grpc"`
	Database DatabaseConf `yaml:"database"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type StorageConf struct {
	Type string `yaml:"type"`
}

type ServerConf struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConf struct {
	DSN string `yaml:"dsn"`
}

func NewConfig(configPath string) (Config, error) {
	var config Config

	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(fileContent, &config)
	return config, err
}
