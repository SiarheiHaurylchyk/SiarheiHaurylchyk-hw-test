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
	internalgrpc "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/server/http"
	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/lib/pq"
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

	httpServer := internalhttp.NewServer(logg, calendar, config.Server.Host, config.Server.Port)
	grpcServer := internalgrpc.NewServer(logg, calendar, config.GRPC.Host, config.GRPC.Port)

	go func() {
		<-ctx.Done()

		stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer stopCancel()

		if err := httpServer.Stop(stopCtx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
		if err := grpcServer.Stop(stopCtx); err != nil {
			logg.Error("failed to stop grpc server: " + err.Error())
		}
	}()

	go func() {
		if err := grpcServer.Start(); err != nil {
			logg.Error("failed to start grpc server: " + err.Error())
			cancel()
		}
	}()

	logg.Info("calendar is running...")

	if err = httpServer.Start(); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1)
	}
}
