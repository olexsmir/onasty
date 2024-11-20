package notesrv

import (
	"errors"

	"github.com/olexsmir/onasty/internal/dtos"
)

// GetNoteBySlugInput used as input for [GetBySlugAndRemoveIfNeeded]
type GetNoteBySlugInput struct {
	// Slug is a note's slug :) *Required*
	Slug dtos.NoteSlugDTO

	// Password is a note's password.
	// Optional, needed only if note has one.
	Password string
}

func (i GetNoteBySlugInput) Validate() error {
	if i.Slug == "" {
		// TODO: make it as sep error
		//nolint:err113
		return errors.New("slug is required")
	}

	return nil
}

func (i GetNoteBySlugInput) HasPassword() bool {
	return i.Password != ""
}
