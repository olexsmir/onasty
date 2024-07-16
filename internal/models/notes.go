package models

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrNoteContentIsEmpty = errors.New("note: content is empty")
	ErrNoteExpired        = errors.New("note: expired")
)

type Note struct {
	ID                   uuid.UUID
	Content              string
	Slug                 string
	BurnBeforeExpiration bool
	CreatedAt            time.Time
	ExpiresAt            time.Time
}

func (n Note) Validate() error {
	if n.Content == "" {
		return ErrNoteContentIsEmpty
	}

	if n.IsExpired() {
		return ErrNoteExpired
	}

	return nil
}

func (n Note) IsExpired() bool {
	return !n.ExpiresAt.IsZero() &&
		n.ExpiresAt.Before(time.Now())
}