package userrepo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
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
	query, args, err := pgq.
		Select("id", "username", "email", "password", "created_at", "last_login_at").
		From("users").
		Where(pgq.Eq{
			"email":    inp.Email,
			"password": inp.Password,
		}).
		SQL()
	if err != nil {
		return domain.User{}, err
	}

	var res domain.User
	err = s.db.QueryRow(ctx, query, args...).
		Scan(&res.ID, &res.Username, &res.Email, &res.Password, &res.CreatedAt, &res.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return res, domain.ErrUserNotFound
	}

	return res, err
}

func (s *Store) GetUserByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (domain.User, error) {
	query := `
select id, username, email, password, created_at, last_login_at from users
where id = (select user_id from sessions where refresh_token = $1);`

	var res domain.User
	err := s.db.QueryRow(ctx, query, refreshToken).
		Scan(&res.ID, &res.Username, &res.Email, &res.Password, &res.CreatedAt, &res.LastLoginAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return res, domain.ErrUsersSessionNotFound
	}

	return res, nil
}

func (s *Store) SetSession(
	ctx context.Context,
	id uuid.UUID,
	refreshToken string,
	expiresAt time.Time,
) error {
	query, args, err := pgq.
		Insert("sessions").
		Columns("user_id", "refresh_token", "expires_at").
		Values(id, refreshToken, expiresAt).
		SQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	return err
}

func (s *Store) UpdateSession(
	ctx context.Context,
	userID uuid.UUID,
	refreshToken string,
	newRefreshToken string,
) error {
	query := `
update sessions
set
  refresh_token = $1
where
  user_id = $2
  and refresh_token = $3
  and expires_at < now()
`

	_, err := s.db.Exec(ctx, query, newRefreshToken, userID, refreshToken)
	// if res.RowsAffected() != 1 {
	// 	return domain.ErrUsersSessionNotFound
	// }

	return err
}

func (s *Store) RemoveSession(ctx context.Context, userID uuid.UUID) error {
	query, args, err := pgq.
		Delete("sessions").
		Where(pgq.Eq{
			// NOTE: also, add refreshToken?
			"user_id": userID,
		}).
		SQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	return err
}
