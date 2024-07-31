package vertokrepo

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

type VerificationTokenStorer interface {
	Create(
		ctx context.Context,
		token string,
		userId uuid.UUID,
		createdAt, expiresAt time.Time,
	) error

	GetUserIdByTookenAndMarkAsUsed(
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

func (r *VerificationTokenRepo) GetUserIdByTookenAndMarkAsUsed(
	ctx context.Context,
	token string,
	usedAT time.Time,
) (uuid.UUID, error) {
	query, aggs, err := pgq.
		Update("verification_tokens").
		Set("used_at", usedAT).
		Where(pgq.Eq{"token": token}).
		Returning("user_id").
		SQL()
	if err != nil {
		return uuid.Nil, err
	}

	var userID uuid.UUID
	err = r.db.QueryRow(ctx, query, aggs...).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, models.ErrUserNotFound
	}

	return userID, err
}
