package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/server/http"
	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		fmt.Println("failed to read config: " + err.Error())
		os.Exit(1)
	}
	logg := logger.New(config.Logger.Level)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var eventStorage storage.EventStorage = memorystorage.New()
	if config.Storage.Type == "sql" {
		sqlStorage := sqlstorage.New(config.Database.DSN)
		if err = sqlStorage.Connect(ctx); err != nil {
			logg.Error("failed to connect to database: " + err.Error())
			os.Exit(1) //nolint:gocritic
		}
		defer sqlStorage.Close()
		eventStorage = sqlStorage
	}

	calendar := app.New(logg, eventStorage)

	server := internalhttp.NewServer(logg, calendar, config.Server.Host, config.Server.Port)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err = server.Start(); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1)
	}
}
