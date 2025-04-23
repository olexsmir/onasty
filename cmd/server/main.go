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
	"github.com/nats-io/nats.go"
	"github.com/olexsmir/onasty/internal/config"
	"github.com/olexsmir/onasty/internal/events/mailermq"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/logger"
	"github.com/olexsmir/onasty/internal/metrics"
	"github.com/olexsmir/onasty/internal/service/notesrv"
	"github.com/olexsmir/onasty/internal/service/usersrv"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
	"github.com/olexsmir/onasty/internal/store/rdb"
	"github.com/olexsmir/onasty/internal/store/rdb/notecache"
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

//nolint:err113,funlen
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
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return err
	}

	psqlDB, err := psqlutil.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}

	redisDB, err := rdb.Connect(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return err
	}

	userPasswordHasher := hasher.NewSHA256Hasher(cfg.PasswordSalt)
	notePasswordHasher := hasher.NewSHA256Hasher(cfg.NotePasswordSalt)
	jwtTokenizer := jwtutil.NewJWTUtil(cfg.JwtSigningKey, cfg.JwtAccessTokenTTL)

	mailermq := mailermq.New(nc)

	sessionrepo := sessionrepo.New(psqlDB)
	vertokrepo := vertokrepo.New(psqlDB)

	userepo := userepo.New(psqlDB)
	usercache := usercache.New(redisDB, cfg.CacheUsersTTL)
	usersrv := usersrv.New(
		userepo,
		sessionrepo,
		vertokrepo,
		userPasswordHasher,
		jwtTokenizer,
		mailermq,
		usercache,
		cfg.JwtRefreshTokenTTL,
		cfg.VerificationTokenTTL,
	)

	notecache := notecache.New(redisDB, cfg.CacheNoteTTL)
	noterepo := noterepo.New(psqlDB)
	notesrv := notesrv.New(noterepo, notePasswordHasher, notecache)

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
	srv := httpserver.NewServer(handler.Handler(), httpserver.Config{
		Port:            cfg.HTTPPort,
		ReadTimeout:     cfg.HTTPReadTimeout,
		WriteTimeout:    cfg.HTTPWriteTimeout,
		MaxHeaderSizeMb: cfg.HTTPHeaderMaxSizeMb,
	})
	go func() {
		slog.Info("starting http server", "port", cfg.HTTPPort)
		if err := srv.Start(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start http server", "error", err)
		}
	}()

	// metrics
	if cfg.MetricsEnabled {
		mSrv := httpserver.NewDefaultServer(metrics.Handler(), cfg.MetricsPort)
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
