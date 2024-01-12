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
	ID        uuid.UUID
	Content   string
	Slug      string
	CreatedAt time.Time
	ExpiresAt time.Time
}
