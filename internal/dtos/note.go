package dtos

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type NoteSlugDTO = string

type NoteDTO struct {
	Content              string
	Slug                 string
	BurnBeforeExpiration bool
	Password             string
	IsRead               bool
	ReadAt               *time.Time
	CreatedAt            time.Time
	ExpiresAt            time.Time
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
