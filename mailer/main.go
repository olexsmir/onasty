package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	cfg := NewConfig()

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		panic(err)
	}

	svc, err := micro.AddService(nc, micro.Config{ //nolint:exhaustruct
		Name:    "mailer",
		Version: "0.0.1",
	})
	if err != nil {
		panic(err)
	}

	mg := NewMailgun(cfg.MailgunFrom, cfg.MailgunDomain, cfg.MailgunAPIKey)
	service := NewService(cfg.AppURL, mg)
	handlers := NewHandlers(service)

	if err := handlers.RegisterAll(svc); err != nil {
		slog.Error("failed to register handlers", "err", err)
	}

	slog.Info("should be running")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("stopping")

	if err := svc.Stop(); err != nil {
		slog.Error("failed stopping service", "err", err)
	}

	if err := nc.Drain(); err != nil {
		slog.Error("failed to close nats connection", "err", err)
	}
}
