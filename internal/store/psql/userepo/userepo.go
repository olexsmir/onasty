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
	MarkUserAsActivated(ctx context.Context, id uuid.UUID) error

	CheckIfUserExists(ctx context.Context, id uuid.UUID) (bool, error)
	CheckIfUserIsActivated(ctx context.Context, id uuid.UUID) (bool, error)
}

var _ UserStorer = (*UserRepo)(nil)

type UserRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) *UserRepo {
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

func (r *UserRepo) MarkUserAsActivated(ctx context.Context, id uuid.UUID) error {
	query, args, err := pgq.
		Update("users").
		Set("activated ", true).
		Where(pgq.Eq{"id": id.String()}).
		SQL()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *UserRepo) CheckIfUserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`,
		id.String(),
	).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, models.ErrUserNotFound
	}

	return exists, err
}

func (r *UserRepo) CheckIfUserIsActivated(ctx context.Context, id uuid.UUID) (bool, error) {
	var activated bool
	err := r.db.QueryRow(ctx, `SELECT activated FROM users WHERE id = $1`, id.String()).
		Scan(&activated)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, models.ErrUserNotFound
	}
	return activated, err
}
