package main

import (
	"context"
	"errors"
	"time"

	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
)

func seedNotes(ctx context.Context, repo noterepo.NoteStorer) error {
	var errs error
	err := repo.Create(ctx, models.Note{ //nolint:exhaustruct
		Content:              "that is the test note",
		Slug:                 "test1",
		Password:             "",
		BurnBeforeExpiration: false,
		CreatedAt:            time.Now(),
	})
	if err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}
