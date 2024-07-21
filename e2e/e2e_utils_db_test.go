package e2e

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
	"github.com/olexsmir/onasty/internal/models"
)

func (e *AppTestSuite) getUserFromDBByUsername(username string) models.User {
	query, args, err := pgq.
		Select("id", "username", "email", "password", "created_at", "last_login_at").
		From("users").
		Where(pgq.Eq{
			"username": username,
		}).
		SQL()
	e.require.NoError(err)

	var user models.User
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.LastLoginAt)
	e.require.NoError(err)

	return user
}

func (e *AppTestSuite) insertUserIntoDB(uname, email, passwd string) uuid.UUID {
	p, err := e.hasher.Hash(passwd)
	e.require.NoError(err)

	query, args, err := pgq.
		Insert("users").
		Columns("username", "email", "password", "activated", "created_at", "last_login_at").
		Values(uname, email, p, true, time.Now(), time.Now()).
		Returning("id").
		SQL()
	e.require.NoError(err)

	var id uuid.UUID
	err = e.postgresDB.QueryRow(e.ctx, query, args...).Scan(&id)
	e.require.NoError(err)

	return id
}

func (e *AppTestSuite) getLastUserSessionByUserID(uid uuid.UUID) models.Session {
	query, args, err := pgq.
		Select("refresh_token", "expires_at").
		From("sessions").
		Where(pgq.Eq{"user_id": uid.String()}).
		OrderBy("expires_at DESC").
		SQL()
	e.require.NoError(err)

	var session models.Session
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&session.RefreshToken, &session.ExpiresAt)
	if errors.Is(pgx.ErrNoRows, err) {
		return models.Session{}
	}

	e.require.NoError(err)
	return session
}

func (e *AppTestSuite) getNoteFromDBbySlug(slug string) models.Note {
	query, args, err := pgq.
		Select("content", "slug", "burn_before_expiration", "created_at", "expires_at").
		From("notes").
		Where("slug = ?", slug).
		SQL()
	e.require.NoError(err)

	var note models.Note
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&note.Content, &note.Slug, &note.BurnBeforeExpiration, &note.CreatedAt, &note.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Note{}
	}

	e.require.NoError(err)
	return note
}
