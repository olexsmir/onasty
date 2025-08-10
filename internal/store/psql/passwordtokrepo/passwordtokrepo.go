package passwordtokrepo

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

type PasswordResetTokenStorer interface {
	// Create a new password reset token.
	Create(ctx context.Context, input models.ResetPasswordToken) error

	// GetUserIDByTokenAndMarkAsUsed gets the token, and marks it as used.
	//
	// In case the token is not found, returns [model.ErrResetPasswordTokenNotFound]
	// If token if used, or expired, returns [model.ErrResetPasswordTokenAlreadyUsed],
	// or [models.ErrResetPasswordTokenExpired].
	GetUserIDByTokenAndMarkAsUsed(
		ctx context.Context,
		token string,
		usedAT time.Time,
	) (uuid.UUID, error)
}

var _ PasswordResetTokenStorer = (*PasswordResetTokenRepo)(nil)

type PasswordResetTokenRepo struct {
	db *psqlutil.DB
}

func NewPasswordResetTokenRepo(db *psqlutil.DB) *PasswordResetTokenRepo {
	return &PasswordResetTokenRepo{
		db: db,
	}
}

func (r *PasswordResetTokenRepo) Create(ctx context.Context, token models.ResetPasswordToken,
) error {
	query, aggs, err := pgq.
		Insert("password_reset_tokens").
		Columns("user_id", "token", "created_at", "expires_at").
		Values(token.UserID, token.Token, token.CreatedAt, token.ExpiresAt).
		SQL()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, aggs...)
	return err
}

func (r *PasswordResetTokenRepo) GetUserIDByTokenAndMarkAsUsed(
	ctx context.Context,
	token string,
	usedAt time.Time,
) (uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var isUsed bool
	var expiresAt time.Time
	err = tx.QueryRow(ctx, "select (used_at is not null), expires_at from password_reset_tokens where token = $1", token).
		Scan(&isUsed, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, models.ErrResetPasswordTokenNotFound
		}
		return uuid.Nil, err
	}

	if isUsed {
		return uuid.Nil, models.ErrResetPasswordTokenAlreadyUsed
	}

	if time.Now().After(expiresAt) {
		return uuid.Nil, models.ErrResetPasswordTokenExpired
	}

	query := `--sql
update password_reset_tokens
set used_at = $1
where token = $2
returning user_id`

	var userID uuid.UUID
	err = tx.QueryRow(ctx, query, usedAt, token).Scan(&userID)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, tx.Commit(ctx)
}
