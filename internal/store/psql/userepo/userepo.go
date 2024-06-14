package userepo

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type UserStorer interface {
	Create(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error)
	GetUserByCredentials(ctx context.Context, email, password string) (dtos.UserDTO, error)
}

type UserRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) UserStorer {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Create(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error) {
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

func (r *UserRepo) GetUserByCredentials(
	ctx context.Context,
	email, password string,
) (dtos.UserDTO, error) {
	query, args, err := pgq.
		Select("id", "username", "email", "password", "created_at", "last_login_at").
		From("users").
		Where(pgq.Eq{
			"email":    email,
			"password": password,
		}).
		SQL()
	if err != nil {
		return dtos.UserDTO{}, err
	}

	var user dtos.UserDTO
	err = r.db.QueryRow(ctx, query, args...).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return dtos.UserDTO{}, models.ErrUserNotFound
	}

	return user, err
}
