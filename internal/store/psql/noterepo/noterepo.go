package noterepo

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/henvic/pgq"
	"github.com/jackc/pgx/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type NoteStorer interface {
	// Create creates a note.
	Create(ctx context.Context, note models.Note) error

	// GetBySlug gets a note by slug.
	// Returns [models.ErrNoteNotFound] if note is not found.
	GetBySlug(ctx context.Context, slug dtos.NoteSlug) (models.Note, error)

	GetNotesByAuthorID(ctx context.Context, authorID uuid.UUID) ([]models.Note, error)

	// GetBySlugAndPassword gets a note by slug and password.
	// the "password" should be hashed.
	//
	// Returns [models.ErrNoteNotFound] if note is not found.
	GetBySlugAndPassword(
		ctx context.Context,
		slug dtos.NoteSlug,
		password string,
	) (models.Note, error)

	// RemoveBySlug marks note as read, deletes it's content, and keeps meta data
	// Returns [models.ErrNoteNotFound] if note is not found.
	RemoveBySlug(ctx context.Context, slug dtos.NoteSlug, readAt time.Time) error

	// SetAuthorIDBySlug assigns author to note by slug.
	// Returns [models.ErrNoteNotFound] if note is not found.
	SetAuthorIDBySlug(ctx context.Context, slug dtos.NoteSlug, authorID uuid.UUID) error
}

var _ NoteStorer = (*NoteRepo)(nil)

type NoteRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) *NoteRepo {
	return &NoteRepo{db}
}

func (s *NoteRepo) Create(ctx context.Context, inp models.Note) error {
	query, args, err := pgq.
		Insert("notes").
		Columns("content", "slug", "password", "burn_before_expiration", "created_at", "expires_at").
		Values(inp.Content, inp.Slug, inp.Password, inp.BurnBeforeExpiration, inp.CreatedAt, inp.ExpiresAt).
		SQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	if psqlutil.IsDuplicateErr(err) {
		return models.ErrNoteSlugIsAlreadyInUse
	}

	return err
}

func (s *NoteRepo) GetBySlug(ctx context.Context, slug dtos.NoteSlug) (models.Note, error) {
	query, args, err := pgq.
		Select("content", "slug", "burn_before_expiration", "read_at", "created_at", "expires_at").
		From("notes").
		Where("(password is null or password = '')").
		Where(pgq.Eq{"slug": slug}).
		SQL()
	if err != nil {
		return models.Note{}, err
	}

	var note models.Note
	err = s.db.QueryRow(ctx, query, args...).
		Scan(&note.Content, &note.Slug, &note.BurnBeforeExpiration, &note.ReadAt, &note.CreatedAt, &note.ExpiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Note{}, models.ErrNoteNotFound
	}

	return note, err
}

func (s *NoteRepo) GetNotesByAuthorID(
	ctx context.Context,
	authorID uuid.UUID,
) ([]models.Note, error) {
	query := `--sql
	select n.content, n.slug, n.burn_before_expiration, n.read_at, n.created_at, n.expires_at
	from notes n
	right join notes_authors na on n.id = na.note_id
	where na.user_id = $1`

	rows, err := s.db.Query(ctx, query, authorID.String())
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.Content, &note.Slug, &note.BurnBeforeExpiration,
			&note.ReadAt, &note.CreatedAt, &note.ExpiresAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

func (s *NoteRepo) GetBySlugAndPassword(
	ctx context.Context,
	slug dtos.NoteSlug,
	passwd string,
) (models.Note, error) {
	query, args, err := pgq.
		Select("content", "slug", "burn_before_expiration", "read_at", "created_at", "expires_at").
		From("notes").
		Where(pgq.Eq{
			"slug":     slug,
			"password": passwd,
		}).
		SQL()
	if err != nil {
		return models.Note{}, err
	}

	var note models.Note
	err = s.db.QueryRow(ctx, query, args...).
		Scan(&note.Content, &note.Slug, &note.BurnBeforeExpiration, &note.ReadAt, &note.CreatedAt, &note.ExpiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Note{}, models.ErrNoteNotFound
	}

	return note, err
}

func (s *NoteRepo) RemoveBySlug(
	ctx context.Context,
	slug dtos.NoteSlug,
	readAt time.Time,
) error {
	query, args, err := pgq.
		Update("notes").
		Set("content", "").
		Set("read_at", readAt).
		Where(pgq.Eq{
			"slug":    slug,
			"read_at": time.Time{}, // check if time is null
		}).
		SQL()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, query, args...)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ErrNoteNotFound
	}

	return err
}

func (s *NoteRepo) SetAuthorIDBySlug(
	ctx context.Context,
	slug dtos.NoteSlug,
	authorID uuid.UUID,
) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var noteID uuid.UUID
	err = tx.QueryRow(ctx, "select id from notes where slug = $1", slug).Scan(&noteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrNoteNotFound
		}
		return err
	}

	_, err = tx.Exec(
		ctx,
		"insert into notes_authors (note_id, user_id) values ($1, $2)",
		noteID, authorID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
