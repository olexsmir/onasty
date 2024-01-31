package notesrv

import (
	"context"
	"log/slog"
	"time"

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
	// FIXME: dont log content
	slog.With("inp", inp).Info("createing note")

	if inp.Slug == "" {
		inp.Slug = uuid.New().String()
	}

	if inp.Validate() != nil {
		return "", inp.Validate()
	}

	return s.store.Create(ctx, inp)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (domain.Note, error) {
	slog.With("slug", slug).Info("getting note by slug")

	note, err := s.store.GetBySlug(ctx, slug)
	if err != nil {
		return domain.Note{}, err
	}

	isntExprTimeEmpty := !note.ExpiresAt.IsZero()

	if note.ExpiresAt.Before(time.Now()) && isntExprTimeEmpty {
		return domain.Note{}, domain.ErrNoteExpired
	}

	if !note.BurnBeforeExpiration && isntExprTimeEmpty {
		return note, nil
	}

	return note, s.store.DeleteByID(ctx, note.ID)
}
