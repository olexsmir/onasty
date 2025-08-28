package e2e_test

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

// getUserByEmail queries user from db by it's email
func (e *AppTestSuite) getUserByEmail(email string) models.User {
	query, args, err := pgq.
		Select("id", "email", "password", "created_at", "last_login_at").
		From("users").
		Where(pgq.Eq{"email": email}).
		SQL()
	e.require.NoError(err)

	var user models.User
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.LastLoginAt)
	e.require.NoError(err)

	return user
}

// insertUser inserts user into db
func (e *AppTestSuite) insertUser(email, passwd string, activated ...bool) uuid.UUID {
	p, err := e.hasher.Hash(passwd)
	e.require.NoError(err)

	var a bool
	if len(activated) == 1 {
		a = activated[0]
	}

	query, args, err := pgq.
		Insert("users").
		Columns("email", "password", "activated", "created_at", "last_login_at").
		Values(email, p, a, time.Now(), time.Now()).
		Returning("id").
		SQL()
	e.require.NoError(err)

	var id uuid.UUID
	err = e.postgresDB.QueryRow(e.ctx, query, args...).Scan(&id)
	e.require.NoError(err)

	return id
}

// getLastSessionByUserID gets last inserted [models.Session] for particular user
func (e *AppTestSuite) getLastSessionByUserID(uid uuid.UUID) models.Session {
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
	session.UserID = uid
	return session
}

// getLastUserByEmail gets last inserted [models.User] by user's email
func (e *AppTestSuite) getLastUserByEmail(em string) models.User {
	query, args, err := pgq.
		Select("id", "activated", "email", "password", "created_at", "last_login_at").
		From("users").
		Where(pgq.Eq{"email": em}).
		OrderBy("created_at DESC").
		Limit(1).
		SQL()
	e.require.NoError(err)

	var u models.User
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&u.ID, &u.Activated, &u.Email, &u.Password, &u.CreatedAt, &u.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{} //nolint:exhaustruct
	}

	e.require.NoError(err)
	return u
}

// getNoteBySlug gets [models.Note] by slug
func (e *AppTestSuite) getNoteBySlug(slug string) models.Note {
	query, args, err := pgq.
		Select(
			"id",
			"content",
			"slug",
			"keep_before_expiration",
			"password",
			"read_at",
			"created_at",
			"expires_at",
		).
		From("notes").
		Where(pgq.Eq{"slug": slug}).
		SQL()
	e.require.NoError(err)

	var readAt sql.NullTime
	var note models.Note
	err = e.postgresDB.QueryRow(e.ctx, query, args...).
		Scan(&note.ID, &note.Content, &note.Slug, &note.KeepBeforeExpiration, &note.Password, &readAt, &note.CreatedAt, &note.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Note{} //nolint:exhaustruct
	}

	note.ReadAt = psqlutil.NullTimeToTime(readAt)

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
	Extra  string // Extra field (optional)
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

func (e *AppTestSuite) getResetPasswordTokenByUserID(u uuid.UUID) userVerificationToken {
	query, args, err := pgq.
		Select("token", "used_at").
		From("password_reset_tokens ").
		Where(pgq.Eq{"user_id": u.String()}).
		Limit(1).
		SQL()

	e.require.NoError(err)
	var r userVerificationToken
	err = e.postgresDB.QueryRow(e.ctx, query, args...).Scan(&r.Token, &r.UsedAt)
	e.require.NoError(err)
	return r
}

func (e *AppTestSuite) getChangeEmailTokenByUserID(u uuid.UUID) userVerificationToken {
	query, args, err := pgq.
		Select("token", "new_email", "used_at").
		From("change_email_tokens").
		Where(pgq.Eq{"user_id": u.String()}).
		Limit(1).
		SQL()

	e.require.NoError(err)
	var r userVerificationToken
	err = e.postgresDB.QueryRow(e.ctx, query, args...).Scan(&r.Token, &r.Extra, &r.UsedAt)
	e.require.NoError(err)
	return r
}
