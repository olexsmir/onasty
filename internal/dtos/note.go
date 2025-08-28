package dtos

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type NoteSlug = string

type GetNote struct {
	Content              string
	KeepBeforeExpiration bool
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
	KeepBeforeExpiration bool
	Password             string
	CreatedAt            time.Time
	ExpiresAt            time.Time
}

type NoteDetailed struct {
	Content              string
	Slug                 NoteSlug
	KeepBeforeExpiration bool
	HasPassword          bool
	CreatedAt            time.Time
	ExpiresAt            time.Time
	ReadAt               time.Time
}

type PatchNote struct {
	ExpiresAt            *time.Time
	KeepBeforeExpiration *bool
}
