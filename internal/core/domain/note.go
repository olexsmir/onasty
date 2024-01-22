package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNoteNotFound       = errors.New("note: not found")
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

	if !n.ExpiresAt.IsZero() &&
		n.ExpiresAt.Before(time.Now()) {
		return ErrNoteExpired
	}

	return nil
}
