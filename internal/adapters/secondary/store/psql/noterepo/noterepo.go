package noterepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/henvic/pgq"

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
	query, args, err := pgq.Insert("notes").
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
