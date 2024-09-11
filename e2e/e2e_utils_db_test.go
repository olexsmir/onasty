package e2e_test

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
		Where(pgq.Eq{"username": username}).
		SQL()
	e.require.NoError(err)

	var user models.User
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.LastLoginAt)
	e.require.NoError(err)

	return user
}

func (e *AppTestSuite) insertUserIntoDB(uname, email, passwd string, activated ...bool) uuid.UUID {
	p, err := e.hasher.Hash(passwd)
	e.require.NoError(err)

	var a bool
	if len(activated) == 1 {
		a = activated[0]
	}

	query, args, err := pgq.
		Insert("users").
		Columns("username", "email", "password", "activated", "created_at", "last_login_at").
		Values(uname, email, p, a, time.Now(), time.Now()).
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
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Session{} //nolint:exhaustruct
	}

	e.require.NoError(err)
	return session
}

func (e *AppTestSuite) getLastInsertedUserByEmail(em string) models.User {
	query, args, err := pgq.
		Select("id", "username", "activated", "email", "password").
		From("users").
		Where(pgq.Eq{"email": em}).
		OrderBy("created_at DESC").
		Limit(1).
		SQL()
	e.require.NoError(err)

	var u models.User
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&u.ID, &u.Username, &u.Activated, &u.Email, &u.Password)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{} //nolint:exhaustruct
	}

	e.require.NoError(err)
	return u
}

func (e *AppTestSuite) getNoteFromDBbySlug(slug string) models.Note {
	query, args, err := pgq.
		Select("id", "content", "slug", "burn_before_expiration", "created_at", "expires_at").
		From("notes").
		Where(pgq.Eq{"slug": slug}).
		SQL()
	e.require.NoError(err)

	var note models.Note
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&note.ID, &note.Content, &note.Slug, &note.BurnBeforeExpiration, &note.CreatedAt, &note.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Note{} //nolint:exhaustruct
	}

	e.require.NoError(err)
	return note
}

type noteAuthorModel struct {
	noteID uuid.UUID
	userID uuid.UUID
}

func (e *AppTestSuite) getLastNoteAuthorsRecordByAuthorID(uid uuid.UUID) noteAuthorModel {
	qeuery, args, err := pgq.
		Select("note_id", "user_id").
		From("notes_authors").
		Where(pgq.Eq{"user_id": uid.String()}).
		OrderBy("created_at DESC").
		Limit(1).
		SQL()
	e.require.NoError(err)

	var na noteAuthorModel
	err = e.postgresDB.QueryRow(e.ctx, qeuery, args...).Scan(&na.noteID, &na.userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return noteAuthorModel{} //nolint:exhaustruct
	}

	e.require.NoError(err)
	return na
}

type userVerificationToken struct {
	Token  string
	UsedAt *time.Time
}

func (e *AppTestSuite) getVerificationTokenByUserID(u uuid.UUID) userVerificationToken {
	query, args, err := pgq.
		Select("token", "used_at").
		From("verification_tokens").
		Where(pgq.Eq{"user_id": u.String()}).
		SQL()
	e.require.NoError(err)
	var r userVerificationToken
	err = e.postgresDB.QueryRow(e.ctx, query, args...).Scan(&r.Token, &r.UsedAt)
	e.require.NoError(err)
	return r
}
