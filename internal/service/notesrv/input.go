package notesrv

import "github.com/olexsmir/onasty/internal/dtos"

// GetNoteBySlugInput used as input for [GetBySlugAndRemoveIfNeeded]
type GetNoteBySlugInput struct {
	// Slug is a note's slug :) *Required*
	Slug dtos.NoteSlug

	// Password is a note's password.
	// Optional, needed only if note has one.
	Password string
}

func (i GetNoteBySlugInput) HasPassword() bool {
	return i.Password != ""
}
