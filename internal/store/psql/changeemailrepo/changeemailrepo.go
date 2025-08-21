package changeemailrepo

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v4"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type ChangeEmailStorer interface {
	// Create create a change email token.
	Create(ctx context.Context, input models.ChangeEmailToken) error

	GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error)

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

func (c *ChangeEmailRepo) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	query := `--sql
select user_id
from change_email_tokens
where token = $1
`

	var userID uuid.UUID
	err := c.db.QueryRow(ctx, query, token).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, models.ErrChangeEmailTokenNotFound
	}

	return userID, err
}

func (c *ChangeEmailRepo) MarkAsUsed(ctx context.Context, token string, usedAT time.Time) error {
	return nil
}
