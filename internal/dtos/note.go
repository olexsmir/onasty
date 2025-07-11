package dtos

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type NoteSlug = string

type GetNote struct {
	Content              string
	BurnBeforeExpiration bool
	ReadAt               time.Time
	CreatedAt            time.Time
	ExpiresAt            time.Time
}

type NoteMetadata struct {
	HasPassword bool
	CreatedAt   time.Time
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

type NoteDetailed struct {
	Content              string
	Slug                 NoteSlug
	BurnBeforeExpiration bool
	HasPassword          bool
	CreatedAt            time.Time
	ExpiresAt            time.Time
	ReadAt               time.Time
}

type PatchNote struct {
	ExpiresAt            *time.Time
	BurnBeforeExpiration *bool
}
