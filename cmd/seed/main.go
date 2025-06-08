package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/olexsmir/onasty/internal/config"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/logger"
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

	if err := logger.SetDefault(cfg.LogLevel, cfg.LogFormat, cfg.LogShowLine); err != nil {
		return err
	}

	psql, err := psqlutil.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}

	userHasher := hasher.NewSHA256Hasher(cfg.PasswordSalt)
	noteHasher := hasher.NewSHA256Hasher(cfg.NotePasswordSalt)

	if err := seedUsers(ctx, userHasher, psql); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	slog.Info("Users seeded successfully")

	if err := seedNotes(ctx, noteHasher, psql); err != nil {
		return fmt.Errorf("failed to seed notes: %w", err)
	}

	slog.Info("Notes seeded successfully")

	return nil
}
