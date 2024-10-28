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
	"github.com/olexsmir/onasty/internal/logger"
	"github.com/olexsmir/onasty/internal/mailer"
	"github.com/olexsmir/onasty/internal/metrics"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
	"github.com/olexsmir/onasty/internal/store/rdb"
	"github.com/olexsmir/onasty/internal/store/rdb/usercache"
	httptransport "github.com/olexsmir/onasty/internal/transport/http"
	"github.com/olexsmir/onasty/internal/transport/http/httpserver"
	"github.com/olexsmir/onasty/internal/transport/http/ratelimit"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

//nolint:err113
func run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := config.NewConfig()

	// logger
	logger, err := logger.NewCustomLogger(cfg.LogLevel, cfg.LogFormat, cfg.LogShowLine)
	if err != nil {
		return err
	}

	slog.SetDefault(logger)

	// semi dev mode
	if !cfg.IsDevMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	// app deps
	psqlDB, err := psqlutil.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}

	redisDB, err := rdb.Connect(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return err
	}

	sha256Hasher := hasher.NewSHA256Hasher(cfg.PasswordSalt)
	jwtTokenizer := jwtutil.NewJWTUtil(cfg.JwtSigningKey, cfg.JwtAccessTokenTTL)
	mailGunMailer := mailer.NewMailgun(cfg.MailgunFrom, cfg.MailgunDomain, cfg.MailgunAPIKey)

	sessionrepo := sessionrepo.New(psqlDB)
	vertokrepo := vertokrepo.New(psqlDB)

	userepo := userepo.New(psqlDB)
	usercache := usercache.New(redisDB, cfg.CacheUsersTTL)
	usersrv := usersrv.New(
		userepo,
		sessionrepo,
		vertokrepo,
		sha256Hasher,
		jwtTokenizer,
		mailGunMailer,
		usercache,
		cfg.JwtRefreshTokenTTL,
		cfg.VerificationTokenTTL,
		cfg.AppURL,
	)

	noterepo := noterepo.New(psqlDB)
	notesrv := notesrv.New(noterepo, sha256Hasher)

	rateLimiterConfig := ratelimit.Config{
		RPS:   cfg.RateLimiterRPS,
		TTL:   cfg.RateLimiterTTL,
		Burst: cfg.RateLimiterBurst,
	}

	handler := httptransport.NewTransport(
		usersrv,
		notesrv,
		rateLimiterConfig,
	)

	// http server
	srv := httpserver.NewServer(cfg.ServerPort, handler.Handler())
	go func() {
		slog.Info("starting http server", "port", cfg.ServerPort)
		if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start http server", "error", err)
		}
	}()

	// metrics
	if cfg.MetricsEnabled {
		mSrv := httpserver.NewServer(cfg.MetricsPort, metrics.Handler())
		go func() {
			slog.Info("starting metrics server", "port", cfg.MetricsPort)
			if err := mSrv.Start(); !errors.Is(err, http.ErrServerClosed) {
				slog.Error("failed to start metrics server", "error", err)
			}
		}()
	}

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

	if err := redisDB.Close(); err != nil {
		return errors.Join(errors.New("failed to close redis connection"), err)
	}

	return nil
}
