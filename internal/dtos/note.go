package dtos

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type NoteSlugDTO string

func (n NoteSlugDTO) String() string { return string(n) }

type NoteDTO struct {
	Content              string
	Slug                 string
	BurnBeforeExpiration bool
	CreatedAt            time.Time
	ExpiresAt            time.Time
}

type CreateNoteDTO struct {
	Content              string
	UserID               uuid.UUID
	Slug                 string
	BurnBeforeExpiration bool
	CreatedAt            time.Time
	ExpiresAt            time.Time
}
