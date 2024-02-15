package userrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/henvic/pgq"
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql"
	"github.com/olexsmir/onasty/internal/core/domain"
	"github.com/olexsmir/onasty/internal/ports"
)

var _ ports.UserStorer = (*Store)(nil)

type Store struct {
	db *psql.DB
}

func New(db *psql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Create(ctx context.Context, inp domain.User) error {
	query, args, err := pgq.
		Insert("users").
		Columns("username", "email", "password", "created_at", "last_login_at").
		Values(inp.Username, inp.Email, inp.Password, inp.CreatedAt, inp.LastLoginAt).
		SQL()
	if err != nil {
		return err
	}

	if psql.IsDuplicateErr(err) {
		return domain.ErrUserEmailIsAlreadyInUse
	}

	_, err = s.db.Exec(ctx, query, args...)
	return err
}

func (s *Store) GetUserByCredentials(
	ctx context.Context,
	inp domain.UserCredentials,
) (domain.User, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Store) RemoveSession(ctx context.Context, userId uuid.UUID) error {
	panic("not implemented") // TODO: Implement
}
