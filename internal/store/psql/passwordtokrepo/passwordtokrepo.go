package passwordtokrepo

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type PasswordResetTokenStorer interface {
	Create(
		ctx context.Context,
		token string,
		userID uuid.UUID,
		createdAt, expiresAt time.Time,
	) error
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

func (r *PasswordResetTokenRepo) Create(
	ctx context.Context,
	token string,
	userID uuid.UUID,
	createdAt, expiresAt time.Time,
) error {
	query, aggs, err := pgq.
		Insert("password_reset_tokens").
		Columns("user_id", "token", "created_at", "expires_at").
		Values(userID, token, createdAt, expiresAt).
		SQL()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, aggs...)
	return err
}
