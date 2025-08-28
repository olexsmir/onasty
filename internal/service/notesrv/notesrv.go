package notesrv

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
	"github.com/olexsmir/onasty/internal/store/rdb/notecache"
)

var ErrNotePasswordNotProvided = errors.New("note: password was not provided")

type NoteServicer interface {
	// Create creates note
	// if slug is empty it will be generated, otherwise used as is
	// if userID is empty it means user isn't authorized so it will be used
	Create(ctx context.Context, note dtos.CreateNote, userID uuid.UUID) (dtos.NoteSlug, error)

	// GetBySlugAndRemoveIfNeeded returns note by slug, and removes if if needed.
	// If note is not found returns [models.ErrNoteNotFound].
	GetBySlugAndRemoveIfNeeded(
		ctx context.Context,
		input GetNoteBySlugInput,
	) (dtos.GetNote, error)

	// GetNoteMetadataBySlug returns note metadata by slug.
	// If note is not found returns [models.ErrNoteNotFound].
	GetNoteMetadataBySlug(ctx context.Context, slug dtos.NoteSlug) (dtos.NoteMetadata, error)

	// GetAllByAuthorID returns all notes by author id.
	GetAllByAuthorID(ctx context.Context, authorID uuid.UUID) ([]dtos.NoteDetailed, error)

	// GetAllReadByAuthorID returns all notes that ARE READ and authored by author id.
	GetAllReadByAuthorID(ctx context.Context, authorID uuid.UUID) ([]dtos.NoteDetailed, error)

	// GetAllUnreadByAuthorID returns all notes that ARE UNREAD and authored by author id.
	GetAllUnreadByAuthorID(ctx context.Context, authorID uuid.UUID) ([]dtos.NoteDetailed, error)

	// UpdateExpirationTimeSettings updates expiresAt and burnBeforeExpiration.
	// If notes is not found returns [models.ErrNoteNotFound].
	UpdateExpirationTimeSettings(
		ctx context.Context,
		patchData dtos.PatchNote,
		slug dtos.NoteSlug,
		userID uuid.UUID,
	) error

	// UpdatePassword sets or updates notes password.
	// If notes is not found returns [models.ErrNoteNotFound].
	UpdatePassword(ctx context.Context, slug dtos.NoteSlug, passwd string, userID uuid.UUID) error

	// DeleteBySlug deletes note by slug
	DeleteBySlug(ctx context.Context, slug dtos.NoteSlug, userID uuid.UUID) error
}

var _ NoteServicer = (*NoteSrv)(nil)

type NoteSrv struct {
	noterepo noterepo.NoteStorer
	hasher   hasher.Hasher
	cache    notecache.NoteCacher
}

func New(noterepo noterepo.NoteStorer, hasher hasher.Hasher, cache notecache.NoteCacher) *NoteSrv {
	return &NoteSrv{
		noterepo: noterepo,
		hasher:   hasher,
		cache:    cache,
	}
}

func (n *NoteSrv) Create(
	ctx context.Context,
	inp dtos.CreateNote,
	userID uuid.UUID,
) (dtos.NoteSlug, error) {
	slog.DebugContext(ctx, "creating", "inp", inp)

	if inp.Slug == "" {
		inp.Slug = uuid.Must(uuid.NewV4()).String()
	}

	if inp.Password != "" {
		hashedPassword, err := n.hasher.Hash(inp.Password)
		if err != nil {
			return "", err
		}
		inp.Password = hashedPassword
	}

	//nolint:exhaustruct // ID - cannot be predicted, and ReadAt will be set on read
	note := models.Note{
		Content:              inp.Content,
		Slug:                 inp.Slug,
		Password:             inp.Password,
		BurnBeforeExpiration: inp.BurnBeforeExpiration,
		CreatedAt:            inp.CreatedAt,
		ExpiresAt:            inp.ExpiresAt,
	}
	if err := note.Validate(); err != nil {
		return "", err
	}

	if err := n.noterepo.Create(ctx, note); err != nil {
		return "", err
	}

	if !userID.IsNil() {
		if err := n.noterepo.SetAuthorIDBySlug(ctx, inp.Slug, userID); err != nil {
			return "", err
		}
	}

	return inp.Slug, nil
}

func (n *NoteSrv) GetBySlugAndRemoveIfNeeded(
	ctx context.Context,
	inp GetNoteBySlugInput,
) (dtos.GetNote, error) {
	note, err := n.getNote(ctx, inp)
	if err != nil {
		return dtos.GetNote{}, err
	}

	if note.IsExpired() {
		return dtos.GetNote{}, models.ErrNoteExpired
	}

	respNote := dtos.GetNote{
		Content:              note.Content,
		BurnBeforeExpiration: note.BurnBeforeExpiration,
		ReadAt:               note.ReadAt,
		CreatedAt:            note.CreatedAt,
		ExpiresAt:            note.ExpiresAt,
	}

	// since not every note should be burn before expiration
	// we return early if it's not
	if note.ShouldPreserveOnRead() {
		return respNote, nil
	}

	return respNote, n.noterepo.RemoveBySlug(ctx, inp.Slug, time.Now())
}

func (n *NoteSrv) GetNoteMetadataBySlug(
	ctx context.Context,
	slug dtos.NoteSlug,
) (dtos.NoteMetadata, error) {
	note, err := n.noterepo.GetMetadataBySlug(ctx, slug)
	return note, err
}

func (n *NoteSrv) GetAllByAuthorID(
	ctx context.Context,
	authorID uuid.UUID,
) ([]dtos.NoteDetailed, error) {
	notes, err := n.noterepo.GetAllByAuthorID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	return n.mapNoteModelToDto(notes), nil
}

func (n *NoteSrv) GetAllReadByAuthorID(
	ctx context.Context,
	authorID uuid.UUID,
) ([]dtos.NoteDetailed, error) {
	notes, err := n.noterepo.GetAllReadByAuthorID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	return n.mapNoteModelToDto(notes), nil
}

func (n *NoteSrv) GetAllUnreadByAuthorID(
	ctx context.Context,
	authorID uuid.UUID,
) ([]dtos.NoteDetailed, error) {
	notes, err := n.noterepo.GetAllUnreadByAuthorID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	return n.mapNoteModelToDto(notes), nil
}

func (n *NoteSrv) UpdateExpirationTimeSettings(
	ctx context.Context,
	patchData dtos.PatchNote,
	slug dtos.NoteSlug,
	userID uuid.UUID,
) error {
	return n.noterepo.UpdateExpirationTimeSettingsBySlug(ctx, slug, patchData, userID)
}

func (n *NoteSrv) UpdatePassword(
	ctx context.Context,
	slug dtos.NoteSlug,
	passwd string,
	userID uuid.UUID,
) error {
	if len(passwd) == 0 {
		return ErrNotePasswordNotProvided
	}

	hashedPassword, err := n.hasher.Hash(passwd)
	if err != nil {
		return err
	}

	return n.noterepo.UpdatePasswordBySlug(ctx, slug, userID, hashedPassword)
}

func (n *NoteSrv) DeleteBySlug(
	ctx context.Context,
	slug dtos.NoteSlug,
	authorID uuid.UUID,
) error {
	return n.noterepo.DeleteNoteBySlug(ctx, slug, authorID)
}

func (n *NoteSrv) getNote(ctx context.Context, inp GetNoteBySlugInput) (models.Note, error) {
	if note, err := n.cache.GetNote(ctx, inp.Slug); err == nil {
		return note, nil
	}

	note, err := n.getNoteFromDBasedOnInput(ctx, inp)
	if err != nil {
		return models.Note{}, err
	}

	if note.IsRead() {
		if err = n.cache.SetNote(ctx, inp.Slug, note); err != nil {
			slog.ErrorContext(ctx, "notecache", "err", err)
		}
	}

	return note, err
}

func (n *NoteSrv) getNoteFromDBasedOnInput(
	ctx context.Context,
	inp GetNoteBySlugInput,
) (models.Note, error) {
	if inp.HasPassword() {
		hashedPassword, err := n.hasher.Hash(inp.Password)
		if err != nil {
			return models.Note{}, err
		}

		return n.noterepo.GetBySlugAndPassword(ctx, inp.Slug, hashedPassword)
	}
	return n.noterepo.GetBySlug(ctx, inp.Slug)
}

func (n *NoteSrv) mapNoteModelToDto(notes []models.Note) []dtos.NoteDetailed {
	var resNotes []dtos.NoteDetailed
	for _, note := range notes {
		resNotes = append(resNotes, dtos.NoteDetailed{
			Content:              note.Content,
			Slug:                 note.Slug,
			BurnBeforeExpiration: note.BurnBeforeExpiration,
			HasPassword:          note.Password != "",
			CreatedAt:            note.CreatedAt,
			ExpiresAt:            note.ExpiresAt,
			ReadAt:               note.ReadAt,
		})
	}

	return resNotes
}
