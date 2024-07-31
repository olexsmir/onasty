package vertokrepo

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type VerificationTokenStorer interface {
	Create(
		ctx context.Context,
		token string,
		userId uuid.UUID,
		createdAt, expiresAt time.Time,
	) error

	GetUserIDByTokenAndMarkAsUsed(
		ctx context.Context,
		token string,
		usedAT time.Time,
	) (uuid.UUID, error)
}

var _ VerificationTokenStorer = (*VerificationTokenRepo)(nil)

type VerificationTokenRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) *VerificationTokenRepo {
	return &VerificationTokenRepo{
		db: db,
	}
}

func (r *VerificationTokenRepo) Create(
	ctx context.Context,
	token string,
	userId uuid.UUID,
	createdAt, expiresAt time.Time,
) error {
	query, aggs, err := pgq.Insert("verification_tokens ").
		Columns("user_id", "token", "created_at", "expires_at").
		Values(userId, token, createdAt, expiresAt).
		SQL()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, aggs...)
	return err
}

func (r *VerificationTokenRepo) GetUserIDByTokenAndMarkAsUsed(
	ctx context.Context,
	token string,
	usedAT time.Time,
) (uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var isUsed bool
	err = tx.QueryRow(ctx, "select used_at is not null from verification_tokens where token = $1", token).
		Scan(&isUsed)
	if err != nil {
		return uuid.Nil, err
	}

	if isUsed {
		return uuid.Nil, models.ErrUserIsAlreeadyVerified
	}

	var userID uuid.UUID
	err = r.db.QueryRow(ctx, "update verification_tokens set used_at = $1 where token = $2 returning user_id",
		usedAT, token).
		Scan(&userID)

	return userID, err
}
