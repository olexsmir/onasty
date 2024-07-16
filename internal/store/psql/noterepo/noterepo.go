package noterepo

import "github.com/olexsmir/onasty/internal/store/psqlutil"

type NoteStorer interface{}

var _ NoteStorer = (*NoteRepo)(nil)

type NoteRepo struct{}

func New(db *psqlutil.DB) NoteStorer {
	return &NoteRepo{}
}
