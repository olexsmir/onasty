package sessionrepo

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type SessionStorer interface {
	// Set creates new session associated with user.
	Set(ctx context.Context, usedID uuid.UUID, refreshToken string, expiresAt time.Time) error

	// GetUserIDByRefreshToken returns user ID associated with the refresh token.
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (uuid.UUID, error)

	// Update updates refresh token with newer.
	Update(ctx context.Context, userID uuid.UUID, refreshToken string, newRefreshToken string) error

	// Delete deletes session by user ID and their refresh token.
	Delete(ctx context.Context, userID uuid.UUID, refreshToken string) error

	// DeleteAllByUserID deletes all sessions associated with user.
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

var _ SessionStorer = (*SessionRepo)(nil)

type SessionRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) *SessionRepo {
	return &SessionRepo{
		db: db,
	}
}

func (s *SessionRepo) Set(
	ctx context.Context,
	userID uuid.UUID,
	refreshToken string,
	expiresAt time.Time,
) error {
	query, args, err := pgq.
		Insert("sessions").
		Columns("user_id", "refresh_token", "expires_at").
		Values(userID, refreshToken, expiresAt).
		SQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	return err
}

func (s *SessionRepo) Update(
	ctx context.Context,
	userID uuid.UUID,
	refreshToken string,
	newRefreshToken string,
) error {
	query := `--sql
update sessions
set refresh_token = $1
where
  user_id = $2
  and refresh_token = $3
  -- and expires_at < now()
`

	res, err := s.db.Exec(ctx, query, newRefreshToken, userID, refreshToken)
	if res.RowsAffected() != 1 {
		return models.ErrSessionNotFound
	}

	return err
}

func (s *SessionRepo) GetUserIDByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (uuid.UUID, error) {
	query, args, err := pgq.
		Select("user_id").
		From("sessions").
		Where(pgq.Eq{"refresh_token": refreshToken}).
		SQL()
	if err != nil {
		return uuid.UUID{}, err
	}

	var userID uuid.UUID
	err = s.db.QueryRow(ctx, query, args...).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.UUID{}, models.ErrUserNotFound
	}

	return userID, err
}

func (s *SessionRepo) Delete(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	query := `--sql
DELETE FROM sessions
WHERE user_id = $1
  AND refresh_token = $2`

	_, err := s.db.Exec(ctx, query, userID, refreshToken)
	return err
}

func (s *SessionRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `--sql
delete from sessions
where user_id = $1`

	_, err := s.db.Exec(ctx, query, userID)
	return err
}
