package notesrv

import (
	"context"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
)

type NoteServicer interface {
	// Create creates note
	// if slug is empty it will be generated, otherwise used as is
	// if userID is empty it means user isn't authorized so it will be used
	Create(ctx context.Context, note dtos.CreateNoteDTO, userID uuid.UUID) (dtos.NoteSlugDTO, error)

	// GetBySlugAndRemoveIfNeeded returns note by slug, and removes if if needed
	GetBySlugAndRemoveIfNeeded(ctx context.Context, input GetNoteBySlugInput) (dtos.NoteDTO, error)
}

var _ NoteServicer = (*NoteSrv)(nil)

type NoteSrv struct {
	noterepo noterepo.NoteStorer
	hasher   hasher.Hasher
}

func New(noterepo noterepo.NoteStorer, hasher hasher.Hasher) *NoteSrv {
	return &NoteSrv{
		noterepo: noterepo,
		hasher:   hasher,
	}
}

func (n *NoteSrv) Create(
	ctx context.Context,
	inp dtos.CreateNoteDTO,
	userID uuid.UUID,
) (dtos.NoteSlugDTO, error) {
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

	if err := n.noterepo.Create(ctx, inp); err != nil {
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
) (dtos.NoteDTO, error) {
	if err := inp.Validate(); err != nil {
		return dtos.NoteDTO{}, err
	}

	note, err := n.noterepo.GetBySlug(ctx, inp.Slug)
	if err != nil {
		return dtos.NoteDTO{}, err
	}

	m := models.Note{ //nolint:exhaustruct
		ExpiresAt:            note.ExpiresAt,
		BurnBeforeExpiration: note.BurnBeforeExpiration,
	}

	if m.IsExpired() {
		return dtos.NoteDTO{}, models.ErrNoteExpired
	}

	// since not every note should be burn before expiration
	// we return early if it's not
	if m.ShouldBeBurnt() {
		return note, nil
	}

	// TODO: in future not remove, leave some metadata
	// to shot user that note was already seen
	return note, n.noterepo.DeleteBySlug(ctx, note.Slug)
}
