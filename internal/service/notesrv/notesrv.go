package notesrv

import (
	"context"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
)

type NoteServicer interface {
	Create(ctx context.Context, note dtos.CreateNoteDTO) (dtos.NoteSlugDTO, error)
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

func (n *NoteSrv) Create(ctx context.Context, inp dtos.CreateNoteDTO) (dtos.NoteSlugDTO, error) {
	if inp.Slug == "" {
		inp.Slug = uuid.New().String()
	}

	err := n.noterepo.Create(ctx, inp)
	return dtos.NoteSlugDTO(inp.Slug), err
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
