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
		userID uuid.UUID,
		createdAt, expiresAt time.Time,
	) error

	GetUserIDByTokenAndMarkAsUsed(
		ctx context.Context,
		token string,
		usedAT time.Time,
	) (uuid.UUID, error)

	GetTokenOrUpdateTokenByUserID(
		ctx context.Context,
		userID uuid.UUID,
		token string,
		tokenExpirationTime time.Time,
	) (string, error)
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
	userID uuid.UUID,
	createdAt, expiresAt time.Time,
) error {
	query, aggs, err := pgq.Insert("verification_tokens ").
		Columns("user_id", "token", "created_at", "expires_at").
		Values(userID, token, createdAt, expiresAt).
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
	usedAt time.Time,
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

	query := `--sql
update verification_tokens
set used_at = $1
where token = $2
returning user_id`

	var userID uuid.UUID
	err = tx.QueryRow(ctx, query, usedAt, token).Scan(&userID)

	return userID, err
}

func (r *VerificationTokenRepo) GetTokenOrUpdateTokenByUserID(
	ctx context.Context,
	userID uuid.UUID,
	token string,
	tokenExpirationTime time.Time,
) (string, error) {
	query := `--sql
insert into verification_tokens (user_id, token, expires_at)
values ($1, $2, $3)
on conflict (user_id)
  do update set
    token = $2,
    expires_at = $3
returning token`

	var res string
	err := r.db.QueryRow(ctx, query, userID, token, tokenExpirationTime).Scan(&res)
	return res, err
}
