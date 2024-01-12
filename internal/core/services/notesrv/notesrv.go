package notesrv

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/core/domain"
	"github.com/olexsmir/onasty/internal/ports"
)

var _ ports.NoteServicer = (*Service)(nil)

type Service struct {
	store ports.NoteStorer
}

func New(store ports.NoteStorer) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Create(ctx context.Context, inp domain.Note) (string, error) {
	// TODO: validate input

	slog.With("inp", inp).Info("createing note")

	if inp.Slug == "" {
		inp.Slug = uuid.New().String()
	}

	return s.store.Create(ctx, inp)
}

func (s *Service) GetBySlug(ctx context.Context, inp string) (domain.Note, error) {
	slog.With("slug", inp).Info("getting note by slug")

	return s.store.GetBySlug(ctx, inp)
}
