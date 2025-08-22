package changeemailrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type ChangeEmailStorer interface {
	// Create create a change email token.
	Create(ctx context.Context, input models.ChangeEmailToken) error

	// GetByToken returns change email token by its token.
	// Returns [models.ErrChangeEmailTokenNotFound] if not found.
	GetByToken(ctx context.Context, token string) (models.ChangeEmailToken, error)

	// MarkAsUsed marks change email token as used.
	// If not found, returns [models.ErrChangeEmailTokenNotFound].
	// If token is already used, returns [models.ErrChangeEmailTokenIsAlreadyUsed].
	// If token is expired, returns [models.ErrChangeEmailTokenExpired]
	MarkAsUsed(ctx context.Context, token string, usedAT time.Time) error
}

var _ ChangeEmailStorer = (*ChangeEmailRepo)(nil)

type ChangeEmailRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) *ChangeEmailRepo {
	return &ChangeEmailRepo{
		db: db,
	}
}

func (c *ChangeEmailRepo) Create(ctx context.Context, inp models.ChangeEmailToken) error {
	query := `--sql
insert into change_email_tokens (user_id, new_email, token, created_at, expires_at)
values ($1, $2, $3, $4, $5)
`

	_, err := c.db.Exec(ctx, query,
		inp.UserID, inp.NewEmail, inp.Token, inp.CreatedAt, inp.ExpiresAt)
	return err
}

func (c *ChangeEmailRepo) GetByToken(
	ctx context.Context,
	token string,
) (models.ChangeEmailToken, error) {
	query := `--sql
select user_id, new_email, token, created_at, expires_at
from change_email_tokens
where token = $1
`

	var res models.ChangeEmailToken
	err := c.db.QueryRow(ctx, query, token).
		Scan(&res.UserID, &res.NewEmail, &res.Token, &res.CreatedAt, &res.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ChangeEmailToken{}, models.ErrChangeEmailTokenNotFound
	}

	return res, err
}

func (c *ChangeEmailRepo) MarkAsUsed(ctx context.Context, token string, usedAT time.Time) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var isUsed bool
	var expiresAt time.Time
	err = tx.QueryRow(ctx,
		"select (used_at is not null), expires_at from change_email_tokens where token = $1",
		token).
		Scan(&isUsed, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrChangeEmailTokenNotFound
		}
		return err
	}

	if isUsed {
		return models.ErrChangeEmailTokenIsAlreadyUsed
	}

	if time.Now().After(expiresAt) {
		return models.ErrChangeEmailTokenExpired
	}

	query := `--sql
update change_email_tokens
set used_at = $1
where token = $2`

	_, err = tx.Exec(ctx, query, usedAT, token)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
