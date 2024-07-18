package noterepo

import (
	"context"
	"errors"

	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type NoteStorer interface {
	Create(ctx context.Context, inp dtos.CreateNoteDTO) error
	GetBySlug(ctx context.Context, slug dtos.NoteSlugDTO) (dtos.NoteDTO, error)
	DeleteBySlug(ctx context.Context, slug dtos.NoteSlugDTO) error
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

func (s *NoteRepo) GetBySlug(ctx context.Context, slug dtos.NoteSlugDTO) (dtos.NoteDTO, error) {
	query, args, err := pgq.
		Select("content", "slug", "burn_before_expiration", "created_at", "expires_at").
		From("notes").
		Where("slug = ?", slug).
		SQL()
	if err != nil {
		return dtos.NoteDTO{}, err
	}

	var note dtos.NoteDTO
	err = s.db.QueryRow(ctx, query, args...).
		Scan(&note.Content, &note.Slug, &note.BurnBeforeExpiration, &note.CreatedAt, &note.ExpiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return dtos.NoteDTO{}, models.ErrNoteNotFound
	}

	return note, err
}

func (s *NoteRepo) DeleteBySlug(ctx context.Context, slug dtos.NoteSlugDTO) error {
	query, args, err := pgq.
		Delete("notes").
		Where(pgq.Eq{"slug": slug}).
		SQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ErrNoteNotFound
	}

	return err
}
