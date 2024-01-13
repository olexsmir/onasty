package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/olexsmir/onasty/internal/adapters/primary/web"
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql"
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/core/config"
	"github.com/olexsmir/onasty/internal/core/services/notesrv"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		slog.With("error", err).Error("failed to load config")
	}

	ctx := context.Background()
	psqlDB, err := psql.Connect(ctx, psql.Credentials{
		Username: cfg.PostgresUsername,
		Password: cfg.PostgresPassword,
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		Database: cfg.PostgresDatabase,
	})
	if err != nil {
		slog.With("error", err).Error("failed to connect to database")
	}

	noterepo := noterepo.New(psqlDB)
	notesrv := notesrv.New(noterepo)

	handlers := web.NewHandler(web.HandlerDeps{
		NoteService: notesrv,
	})

	// http server
	srv := web.NewServer(cfg.ServerPort, handlers.InitRoutes())
	go func() {
		slog.With("port", cfg.ServerPort).Info("starting http server on")
		if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) {
			slog.With("error", err).Error("failed to start http server")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	if err := srv.Shutdown(context.Background()); err != nil {
		slog.With("error", err).Error("failed to shutdown http server")
	}

	if err := psqlDB.Close(); err != nil {
		slog.With("error", err).Error("failed to disconect form database")
	}
}
