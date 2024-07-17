package notesrv

import (
	"context"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/store/psql/noterepo"
)

type NoteServicer interface {
	Create(ctx context.Context, note dtos.CreateNoteDTO) (dtos.NoteSlugDTO, error)
	GetBySlug(ctx context.Context, slug dtos.NoteSlugDTO) (dtos.NoteDTO, error)
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

func (n *NoteSrv) GetBySlug(ctx context.Context, slug dtos.NoteSlugDTO) (dtos.NoteDTO, error) {
	return dtos.NoteDTO{}, nil
}
