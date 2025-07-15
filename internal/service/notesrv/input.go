package notesrv

import "github.com/olexsmir/onasty/internal/dtos"

const EmptyPassword = ""

// GetNoteBySlugInput used as input for [GetBySlugAndRemoveIfNeeded]
type GetNoteBySlugInput struct {
	// Slug is a note's slug :) *Required*
	Slug dtos.NoteSlug

	// Password is a note's password.
	// Leave it `""` if note has no password.
	Password string
}

func (i GetNoteBySlugInput) HasPassword() bool {
	return i.Password != EmptyPassword
}
