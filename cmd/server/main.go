package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/olexsmir/onasty/internal/config"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
	httptransport "github.com/olexsmir/onasty/internal/transport/http"
	"github.com/olexsmir/onasty/internal/transport/http/httpserver"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := config.NewConfig()
	if err := setupLogger(cfg); err != nil {
		return err
	}

	if !cfg.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	psqlDB, err := psqlutil.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}

	// app deps
	sha256Hasher := hasher.NewSHA256Hasher(cfg.PasswordSalt)
	jwtTokenizer := jwtutil.NewJWTUtil(cfg.JwtSigningKey, cfg.JwtAccessTokenTTL)

	userepo := userepo.New(psqlDB)
	usersrv := usersrv.New(userepo, sha256Hasher, jwtTokenizer)

	handler := httptransport.NewTransport(usersrv)

	// http server
	srv := httpserver.NewServer(cfg.ServerPort, handler.Handler())
	go func() {
		slog.Info("starting http server", "port", cfg.ServerPort)
		if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start http server", "error", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	if err := srv.Stop(ctx); err != nil {
		return errors.Join(errors.New("failed to stop http server"), err)
	}

	if err := psqlDB.Close(); err != nil {
		return errors.Join(errors.New("failed to close postgres connection"), err)
	}

	return nil
}

func setupLogger(cfg *config.Config) error {
	loggerLevels := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
	}

	logLevel, ok := loggerLevels[cfg.LogLevel]
	if !ok {
		return errors.New("unknown log level")
	}

	var slogHandler slog.Handler
	switch cfg.LogFormat {
	case "json":
		slogHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	case "text":
		slogHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	default:
		return errors.New("unknown log format")
	}

	slog.SetDefault(slog.New(slogHandler))

	return nil
}
