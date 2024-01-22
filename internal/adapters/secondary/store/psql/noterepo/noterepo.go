package noterepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"

	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql"
	"github.com/olexsmir/onasty/internal/core/domain"
	"github.com/olexsmir/onasty/internal/ports"
)

var _ ports.NoteStorer = (*Store)(nil)

type Store struct {
	db *psql.DB
}

func New(db *psql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Create(ctx context.Context, inp domain.Note) (string, error) {
	query, args, err := pgq.
		Insert("notes").
		Columns("content", "slug", "created_at", "expires_at").
		Values(inp.Content, inp.Slug, inp.CreatedAt, inp.ExpiresAt).
		Returning("slug").
		SQL()
	if err != nil {
		return "", err
	}

	var res string
	err = s.db.QueryRow(ctx, query, args...).Scan(&res)
	return res, err
}

func (s *Store) GetBySlug(ctx context.Context, slug string) (domain.Note, error) {
	query, args, err := pgq.
		Select("id", "content", "slug", "burn_before_expiration", "created_at", "expires_at").
		From("notes").
		Where("slug = ?", slug).
		SQL()
	if err != nil {
		return domain.Note{}, err
	}

	var res domain.Note
	err = s.db.QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.Content, &res.Slug, &res.BurnBeforeExpiration, &res.CreatedAt, &res.ExpiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return res, domain.ErrNoteNotFound
	}

	return res, err
}

func (s *Store) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query, args, err := pgq.
		Delete("notes").
		Where("id = ?", id).
		SQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNoteNotFound
	}

	return err
}
