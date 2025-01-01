package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/olexsmir/onasty/internal/logger"

	_ "embed"
)

//go:embed version
var version string

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := NewConfig()
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return err
	}

	logger, err := logger.NewCustomLogger(cfg.LogLevel, cfg.LogFormat, cfg.LogShowLine)
	if err != nil {
		return err
	}

	slog.SetDefault(logger)

	//nolint:exhaustruct
	svc, err := micro.AddService(nc, micro.Config{
		Name:    "mailer",
		Version: version,
	})
	if err != nil {
		return err
	}

	mg := NewMailgun(cfg.MailgunFrom, cfg.MailgunDomain, cfg.MailgunAPIKey)
	service := NewService(cfg.AppURL, mg)
	handlers := NewHandlers(service)

	if err := handlers.RegisterAll(svc); err != nil {
		return err
	}

	// TODO: add metrics

	slog.Info("the service is listening")

	// graceful shutdown
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)
	<-quitCh

	slog.Info("stopping the service")

	if err := svc.Stop(); err != nil {
		return err
	}

	if err := nc.Drain(); err != nil {
		return err
	}

	return nil
}
