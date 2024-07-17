package noterepo

import (
	"context"

	"github.com/henvic/pgq"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type NoteStorer interface {
	Create(ctx context.Context, inp dtos.CreateNoteDTO) error
}

var _ NoteStorer = (*NoteRepo)(nil)

type NoteRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) NoteStorer {
	return &NoteRepo{db}
}

func (s *NoteRepo) Create(ctx context.Context, inp dtos.CreateNoteDTO) error {
	query, args, err := pgq.
		Insert("notes").
		Columns("content", "slug", "burn_before_expiration ", "created_at", "expires_at").
		Values(inp.Content, inp.Slug, inp.BurnBeforeExpiration, inp.CreatedAt, inp.ExpiresAt).
		SQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	if psqlutil.IsDuplicateErr(err) {
		return models.ErrNoteSlugIsAlreadyInUse
	}

	return err
}
