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
	ReadAt               time.Time
	CreatedAt            time.Time
	ExpiresAt            time.Time
}

type NoteMetadataDTO struct {
	Slug      string    `json:"slug"`
	IsRead    bool      `json:"is_read"`
	ReadAt    time.Time `json:"read_at"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
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
