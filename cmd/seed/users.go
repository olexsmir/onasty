package main

import (
	"context"
	"errors"
	"time"

	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
)

func seedUsers(ctx context.Context, hash hasher.Hasher, repo userepo.UserStorer) error {
	var errs error

	adminPassword, err := hash.Hash("admin")
	if err != nil {
		return errors.Join(errs, err)
	}
	_, err = repo.Create(ctx, models.User{ //nolint:exhaustruct
		Email:       "admin@onasty.local",
		Activated:   true,
		Password:    adminPassword,
		CreatedAt:   time.Now(),
		LastLoginAt: time.Now(),
	})
	if err != nil {
		errs = errors.Join(errs, err)
	}

	userPassword, err := hash.Hash("qwerty")
	if err != nil {
		return errors.Join(errs, err)
	}
	_, err = repo.Create(ctx, models.User{ //nolint:exhaustruct
		Email:       "user@onasty.local",
		Activated:   false,
		Password:    userPassword,
		CreatedAt:   time.Now(),
		LastLoginAt: time.Now(),
	})
	if err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}
