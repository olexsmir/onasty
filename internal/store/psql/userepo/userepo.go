package userepo

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type UserStorer interface {
	Create(ctx context.Context, inp CreateInput) (uuid.UUID, error)
}

type UserRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) UserStorer {
	return &UserRepo{
		db: db,
	}
}

type CreateInput struct {
	Username    string
	Email       string
	Password    string
	CreatedAt   time.Time
	LastLoginAt time.Time
}

func (r *UserRepo) Create(ctx context.Context, inp CreateInput) (uuid.UUID, error) {
	query, args, err := pgq.
		Insert("users").
		Columns("username", "email", "password", "created_at", "last_login_at").
		Values(inp.Username, inp.Email, inp.Password, inp.CreatedAt, inp.LastLoginAt).
		Returning("id").
		SQL()
	if err != nil {
		return uuid.UUID{}, err
	}

	var id uuid.UUID
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)

	// FIXME: somehow this does return errors but i can't errors.Is them in api layer
	if psqlutil.IsDuplicateErr(err, "users_username_key") {
		return uuid.UUID{}, models.ErrUsernameIsAlreadyInUse
	}

	if psqlutil.IsDuplicateErr(err, "users_email_key") {
		return uuid.UUID{}, models.ErrUserEmailIsAlreadyInUse
	}

	return id, err
}
