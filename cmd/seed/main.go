package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/olexsmir/onasty/internal/config"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/logger"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg := config.NewConfig()
	cfg.PostgresDSN = os.Getenv("MIGRATION_DSN")

	logger, err := logger.NewCustomLogger(cfg.LogLevel, cfg.LogFormat, cfg.LogShowLine)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	psql, err := psqlutil.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}

	hasher := hasher.NewSHA256Hasher(cfg.PasswordSalt)
	userrepo := userepo.New(psql)
	noterepo := noterepo.New(psql)

	if err := seedUsers(ctx, hasher, userrepo); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	slog.Info("Users seeded successfully")

	if err := seedNotes(ctx, noterepo); err != nil {
		return fmt.Errorf("failed to seed notes: %w", err)
	}

	slog.Info("Notes seeded successfully")

	return nil
}
