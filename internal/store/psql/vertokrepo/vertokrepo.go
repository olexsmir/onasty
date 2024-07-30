package vertokrepo

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type VerificationTokenStorer interface {
	Create(
		ctx context.Context,
		token string,
		userId uuid.UUID,
		createdAt, expiresAt time.Time,
	) error

	MarkAsVerified(ctx context.Context, token string, verifiedAt time.Time) error
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

func (r *VerificationTokenRepo) MarkAsVerified(
	ctx context.Context,
	token string,
	verifiedAt time.Time,
) error {
	return nil
}
