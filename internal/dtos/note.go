package dtos

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type NoteSlug = string

type GetNote struct {
	Content   string
	ReadAt    time.Time
	CreatedAt time.Time
	ExpiresAt time.Time
}

type CreateNote struct {
	Content              string
	UserID               uuid.UUID
	Slug                 NoteSlug
	BurnBeforeExpiration bool
	Password             string
	CreatedAt            time.Time
	ExpiresAt            time.Time
}
