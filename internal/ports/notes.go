package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/core/domain"
)

type NoteServicer interface {
	Create(context.Context, domain.Note) (string, error)
	GetBySlug(context.Context, string) (domain.Note, error)
}

type NoteStorer interface {
	Create(context.Context, domain.Note) (string, error)
	GetBySlug(context.Context, string) (domain.Note, error)
	DeleteByID(context.Context, uuid.UUID) error
}
