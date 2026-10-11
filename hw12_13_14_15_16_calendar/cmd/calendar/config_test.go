package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	configText := `
logger:
  level: debug
storage:
  type: sql
server:
  host: 0.0.0.0
  port: 9090
database:
  dsn: postgres://user:password@localhost:5432/calendar
`
	if err := os.WriteFile(configPath, []byte(configText), 0o600); err != nil {
		t.Fatal(err)
	}

	config, err := NewConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if config.Logger.Level != "debug" || config.Storage.Type != "sql" {
		t.Fatalf("wrong logger or storage config: %+v", config)
	}
	if config.Server.Host != "0.0.0.0" || config.Server.Port != 9090 {
		t.Fatalf("wrong server config: %+v", config.Server)
	}
	if config.Database.DSN == "" {
		t.Fatal("database dsn is empty")
	}
}
