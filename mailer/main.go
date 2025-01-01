package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/olexsmir/onasty/internal/logger"
	"github.com/olexsmir/onasty/internal/transport/http/httpserver"

	_ "embed"
)

//go:embed version
var _version string

var version = strings.Trim(_version, "\n")

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

	if cfg.MetricsEnabled {
		srv := httpserver.NewServer(cfg.MetricsPort, MetricsHandler())
		go func() {
			slog.Info("starting metrics server", "port", cfg.MetricsPort)
			if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) {
				slog.Error("failed to start metrics server", "error", err)
			}
		}()
	}

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
