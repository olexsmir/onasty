package notesrv

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
)

type NoteServicer interface {
	// Create create note
	// if slug is empty it will be generated, otherwise used as is
	// if userID is empty it means user isn't authorized so it will be used
	Create(ctx context.Context, note dtos.CreateNoteDTO, userID uuid.UUID) (dtos.NoteSlugDTO, error)
	GetBySlugAndRemoveIfNeeded(ctx context.Context, slug dtos.NoteSlugDTO) (dtos.NoteDTO, error)
}

var _ NoteServicer = (*NoteSrv)(nil)

type NoteSrv struct {
	noterepo noterepo.NoteStorer
}

func New(noterepo noterepo.NoteStorer) NoteServicer {
	return &NoteSrv{
		noterepo: noterepo,
	}
}

func (n *NoteSrv) Create(
	ctx context.Context,
	inp dtos.CreateNoteDTO,
	userID uuid.UUID,
) (dtos.NoteSlugDTO, error) {
	if inp.Slug == "" {
		inp.Slug = uuid.Must(uuid.NewV4()).String()
	}

	err := n.noterepo.Create(ctx, inp)
	if err != nil {
		return "", err
	}

	if !userID.IsNil() {
		if err := n.noterepo.SetAuthorIDBySlug(ctx, dtos.NoteSlugDTO(inp.Slug), userID); err != nil {
			return "", err
		}
	}

	return dtos.NoteSlugDTO(inp.Slug), nil
}

func (n *NoteSrv) GetBySlugAndRemoveIfNeeded(
	ctx context.Context,
	slug dtos.NoteSlugDTO,
) (dtos.NoteDTO, error) {
	note, err := n.noterepo.GetBySlug(ctx, slug)
	if err != nil {
		return dtos.NoteDTO{}, err
	}

	// TODO: there should be a better way to do it
	m := models.Note{
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
	// to shot user that note was alreasy seen
	return note, n.noterepo.DeleteBySlug(ctx, dtos.NoteSlugDTO(note.Slug))
}
