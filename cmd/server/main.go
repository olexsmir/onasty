//nolint:err113 // all errors are shown to the user so it's ok for them to be dynamic
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
	"github.com/olexsmir/onasty/internal/mailer"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
	httptransport "github.com/olexsmir/onasty/internal/transport/http"
	"github.com/olexsmir/onasty/internal/transport/http/httpserver"
	"github.com/olexsmir/onasty/internal/transport/http/reqid"
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

	// logger
	baseLoger, err := setupBasicLoggerHandler(cfg)
	if err != nil {
		return err
	}

	slog.SetDefault(slog.New(&CustomLogger{Handler: baseLoger}))

	// semi dev mode
	if !cfg.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	// app deps
	psqlDB, err := psqlutil.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}

	sha256Hasher := hasher.NewSHA256Hasher(cfg.PasswordSalt)
	jwtTokenizer := jwtutil.NewJWTUtil(cfg.JwtSigningKey, cfg.JwtAccessTokenTTL)
	mailGunMailer := mailer.NewMailgun(cfg.MailgunFrom, cfg.MailgunDomain, cfg.MailgunAPIKey)

	sessionrepo := sessionrepo.New(psqlDB)
	vertokrepo := vertokrepo.New(psqlDB)

	userepo := userepo.New(psqlDB)
	usersrv := usersrv.New(
		userepo,
		sessionrepo,
		vertokrepo,
		sha256Hasher,
		jwtTokenizer,
		mailGunMailer,
		cfg.JwtRefreshTokenTTL,
		cfg.VerficationTokenTTL,
		cfg.AppURL,
	)

	noterepo := noterepo.New(psqlDB)
	notesrv := notesrv.New(noterepo)

	handler := httptransport.NewTransport(usersrv, notesrv)

	// http server
	srv := httpserver.NewServer(cfg.ServerPort, handler.Handler())
	go func() {
		slog.DebugContext(ctx, "starting http server", "port", cfg.ServerPort)
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

type CustomLogger struct {
	slog.Handler
}

func (l *CustomLogger) Handle(ctx context.Context, r slog.Record) error {
	if requestID := reqid.GetFromContext(ctx); requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}

	return l.Handler.Handle(ctx, r)
}

func setupBasicLoggerHandler(cfg *config.Config) (slog.Handler, error) {
	loggerLevels := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
	}

	logLevel, ok := loggerLevels[cfg.LogLevel]
	if !ok {
		return nil, errors.New("unknown log level")
	}

	handlerOptions := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: cfg.LogShowLine,
	}

	var slogHandler slog.Handler
	switch cfg.LogFormat {
	case "json":
		slogHandler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	case "text":
		slogHandler = slog.NewTextHandler(os.Stdout, handlerOptions)
	default:
		return nil, errors.New("unknown log format")
	}

	return slogHandler, nil
}
