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
	"github.com/olexsmir/onasty/internal/adapters/secondary/hash/argon2"
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql"
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql/userrepo"
	"github.com/olexsmir/onasty/internal/adapters/secondary/tokens/jwt"
	"github.com/olexsmir/onasty/internal/core/config"
	"github.com/olexsmir/onasty/internal/core/services/notesrv"
	"github.com/olexsmir/onasty/internal/core/services/usersrv"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic("failed to load config")
	}

	setupLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	argon2Hasher := argon2.New()
	jwtTokenizer := jwt.New(cfg.JWTSigningKey)

	noterepo := noterepo.New(psqlDB)
	notesrv := notesrv.New(noterepo)

	userrepo := userrepo.New(psqlDB)
	usersrv := usersrv.New(userrepo, argon2Hasher, jwtTokenizer)

	handlers := web.NewHandler(web.HandlerDeps{
		NoteService: notesrv,
		UserService: usersrv,
	})

	// http server
	srv := web.NewServer(cfg.ServerPort, handlers.InitRoutes())
	go func() {
		slog.With("port", cfg.ServerPort).Info("starting http server")
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

func setupLogger(cfg *config.Config) {
	loggerLevels := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
	}

	logLevel, ok := loggerLevels[cfg.LogLevel]
	if !ok {
		panic("unknown log level")
	}

	var slogHandler slog.Handler
	switch cfg.LogFormat {
	case "json":
		slogHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	case "text":
		slogHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	default:
		panic("unknown log format")
	}

	slog.SetDefault(slog.New(slogHandler))
}
