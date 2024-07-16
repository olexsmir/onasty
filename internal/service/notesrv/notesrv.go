package notesrv

import "github.com/olexsmir/onasty/internal/store/psql/noterepo"

type NoteServicer interface{}

var _ NoteServicer = (*NoteSrv)(nil)

type NoteSrv struct {
	noterepo noterepo.NoteStorer
}

func New(noterepo noterepo.NoteStorer) NoteServicer {
	return &NoteSrv{
		noterepo: noterepo,
	}
}
