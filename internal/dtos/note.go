package dtos

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type NoteSlugDTO = string

type GetNoteDTO struct {
	Content   string
	ReadAt    time.Time
	CreatedAt time.Time
	ExpiresAt time.Time
}

type CreateNoteDTO struct {
	Content              string
	UserID               uuid.UUID
	Slug                 string
	BurnBeforeExpiration bool
	Password             string
	CreatedAt            time.Time
	ExpiresAt            time.Time
}
