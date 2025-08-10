package userepo

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type UserStorer interface {
	// Create creates a new user.
	Create(ctx context.Context, inp models.User) (uuid.UUID, error)

	// GetByEmail returns user by email and password
	// the password should be hashed.
	GetByEmail(ctx context.Context, email string) (models.User, error)

	// GetByID returns user by id.
	// If user not found, returns [models.ErrUserNotFound].
	GetByID(ctx context.Context, userID uuid.UUID) (models.User, error)

	// GetUserIDByEmail returns user id that is associated with their email.
	// If user not found, returns [models.ErrUserNotFound].
	GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error)

	// MakrUserAsActivated marks user as activated by their id
	MarkUserAsActivated(ctx context.Context, id uuid.UUID) error

	// ChangePassword changes user password from oldPassword to newPassword
	// and oldPassword and newPassword should be hashed
	ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) error

	// SetPassword sets new password for user by their id
	// password should be hashed
	SetPassword(ctx context.Context, userID uuid.UUID, newPassword string) error

	GetByOAuthID(ctx context.Context, provider, providerID string) (models.User, error)
	LinkOAuthIdentity(ctx context.Context, userID uuid.UUID, provider, providerID string) error

	CheckIfUserExists(ctx context.Context, userID uuid.UUID) (bool, error)
	CheckIfUserIsActivated(ctx context.Context, userID uuid.UUID) (bool, error)
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

func (r *UserRepo) Create(ctx context.Context, inp models.User) (uuid.UUID, error) {
	query, args, err := pgq.
		Insert("users").
		Columns("email", "password", "activated", "created_at", "last_login_at").
		Values(inp.Email, inp.Password, inp.Activated, inp.CreatedAt, inp.LastLoginAt).
		Returning("id").
		SQL()
	if err != nil {
		return uuid.UUID{}, err
	}

	var id uuid.UUID
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)

	// FIXME: somehow this does return errors but i can't errors.Is them in api layer
	if psqlutil.IsDuplicateErr(err, "users_email_key") {
		return uuid.UUID{}, models.ErrUserEmailIsAlreadyInUse
	}

	return id, err
}

func (r *UserRepo) GetByEmail(
	ctx context.Context,
	email string,
) (models.User, error) {
	query, args, err := pgq.
		Select("id", "email", "password", "activated", "created_at", "last_login_at").
		From("users").
		Where(pgq.Eq{"email": email}).
		SQL()
	if err != nil {
		return models.User{}, err
	}

	var user models.User
	err = r.db.QueryRow(ctx, query, args...).
		Scan(&user.ID, &user.Email, &user.Password, &user.Activated, &user.CreatedAt, &user.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrUserNotFound
	}

	return user, err
}

func (r *UserRepo) GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	query, args, err := pgq.
		Select("id").
		From("users").
		Where(pgq.Eq{"email": email}).
		SQL()
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, models.ErrUserNotFound
	}

	return id, err
}

func (r *UserRepo) GetByID(ctx context.Context, userID uuid.UUID) (models.User, error) {
	query := `--sql
select id, email, password, activated, created_at, last_login_at
from users
where id = $1`

	var user models.User
	err := r.db.QueryRow(ctx, query, userID).
		Scan(&user.ID, &user.Email, &user.Password, &user.Activated, &user.CreatedAt, &user.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrUserNotFound
	}

	return user, err
}

func (r *UserRepo) GetByOAuthID(
	ctx context.Context,
	provider, providerID string,
) (models.User, error) {
	query := `--sql
	select u.id, u.email, u.password, u.activated, u.created_at, u.last_login_at
	from users u
	join oauth_identities oi on u.id = oi.user_id
	where oi.provider = $1
		and oi.provider_id = $2
	limit 1`

	var user models.User
	err := r.db.QueryRow(ctx, query, provider, providerID).
		Scan(&user.ID, &user.Email, &user.Password, &user.Activated, &user.CreatedAt, &user.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrUserNotFound
	}

	return user, err
}

func (r *UserRepo) LinkOAuthIdentity(
	ctx context.Context,
	userID uuid.UUID,
	provider, providerID string,
) error {
	query := `--sql
insert into oauth_identities (user_id, provider, provider_id)
values ($1, $2, $3)
on conflict (provider, provider_id) do update
set user_id = $1,
	provider = $2,
	provider_id = $3`

	_, err := r.db.Exec(ctx, query, userID, provider, providerID)
	return err
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

func (r *UserRepo) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	newPasswd string,
) error {
	query, args, err := pgq.
		Update("users").
		Set("password", newPasswd).
		Where(pgq.Eq{"id": userID.String()}).
		SQL()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *UserRepo) SetPassword(ctx context.Context, userID uuid.UUID, password string) error {
	query, args, err := pgq.
		Update("users").
		Set("password", password).
		Where(pgq.Eq{"id": userID.String()}).
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
