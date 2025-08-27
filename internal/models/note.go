package models

import (
	"errors"
	"regexp"
	"time"

	"github.com/gofrs/uuid/v5"
)

// read and unread are not allowed because those slugs might and will be interpreted as api routes
var notAllowedSlugs = map[string]struct{}{
	"read":   {},
	"unread": {},
}

var (
	ErrNoteContentIsEmpty     = errors.New("note: content is empty")
	ErrNoteSlugIsAlreadyInUse = errors.New("note: slug is already in use")
	ErrNoteSlugIsInvalid      = errors.New("note: slug is invalid")
	ErrNoteCannotBeBurnt      = errors.New(
		"note: cannot be burn before expiration if expiration time is not provided",
	)
	ErrNoteExpired  = errors.New("note: expired")
	ErrNoteNotFound = errors.New("note: not found")
)

type Note struct {
	ID                   uuid.UUID
	Content              string
	Slug                 string
	Password             string
	BurnBeforeExpiration bool
	ReadAt               time.Time
	CreatedAt            time.Time
	ExpiresAt            time.Time
}

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func (n Note) Validate() error {
	if n.Content == "" {
		return ErrNoteContentIsEmpty
	}

	if n.Slug != "" && !slugPattern.MatchString(n.Slug) {
		return ErrNoteSlugIsInvalid
	}

	if n.IsExpired() {
		return ErrNoteExpired
	}

	if n.BurnBeforeExpiration && n.ExpiresAt.IsZero() {
		return ErrNoteCannotBeBurnt
	}

	if _, exists := notAllowedSlugs[n.Slug]; exists {
		return ErrNoteSlugIsAlreadyInUse
	}

	return nil
}

func (n Note) IsExpired() bool {
	return !n.ExpiresAt.IsZero() &&
		n.ExpiresAt.Before(time.Now())
}

func (n Note) ShouldBeBurnt() bool {
	return !n.ExpiresAt.IsZero() &&
		n.BurnBeforeExpiration
}

func (n Note) IsRead() bool {
	return !n.ReadAt.IsZero()
}
