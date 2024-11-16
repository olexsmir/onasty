package noterepo

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

type NoteStorer interface {
	// Create creates a note
	Create(ctx context.Context, inp dtos.CreateNoteDTO) error

	// GetBySlug gets a note by slug.
	// Returns [models.ErrNoteNotFound] if note is not found.
	GetBySlug(ctx context.Context, slug dtos.NoteSlugDTO) (dtos.NoteDTO, error)

	// GetBySlugAndPassword gets a note by slug and password.
	// the "password" should be hashed.
	//
	// Returns [models.ErrNoteNotFound] if note is not found.
	GetBySlugAndPassword(
		ctx context.Context,
		slug dtos.NoteSlugDTO,
		password string,
	) (dtos.NoteDTO, error)

	// DeleteBySlug deletes note by slug or returns [models.ErrNoteNotFound] if note if not found.
	DeleteBySlug(ctx context.Context, slug dtos.NoteSlugDTO) error

	// SetAuthorIDBySlug assigns author to note by slug.
	// Returns [models.ErrNoteNotFound] if note is not found.
	SetAuthorIDBySlug(ctx context.Context, slug dtos.NoteSlugDTO, authorID uuid.UUID) error
}

var _ NoteStorer = (*NoteRepo)(nil)

type NoteRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) *NoteRepo {
	return &NoteRepo{db}
}

func (s *NoteRepo) Create(ctx context.Context, inp dtos.CreateNoteDTO) error {
	query, args, err := pgq.
		Insert("notes").
		Columns("content", "slug", "password", "burn_before_expiration ", "created_at", "expires_at").
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

func (s *NoteRepo) GetBySlug(ctx context.Context, slug dtos.NoteSlugDTO) (dtos.NoteDTO, error) {
	query, args, err := pgq.
		Select("content", "slug", "burn_before_expiration", "created_at", "expires_at").
		From("notes").
		Where("slug = ?", slug).
		SQL()
	if err != nil {
		return dtos.NoteDTO{}, err
	}

	var note dtos.NoteDTO
	err = s.db.QueryRow(ctx, query, args...).
		Scan(&note.Content, &note.Slug, &note.BurnBeforeExpiration, &note.CreatedAt, &note.ExpiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return dtos.NoteDTO{}, models.ErrNoteNotFound
	}

	return note, err
}

func (s *NoteRepo) DeleteBySlug(ctx context.Context, slug dtos.NoteSlugDTO) error {
	query, args, err := pgq.
		Delete("notes").
		Where(pgq.Eq{"slug": slug}).
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
	slug dtos.NoteSlugDTO,
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
