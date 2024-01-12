package notesrv

import "github.com/olexsmir/onasty/internal/ports"

var _ ports.NoteServicer = (*Service)(nil)

type Service struct {
	store ports.NoteStorer
}

func New(store ports.NoteStorer) *Service {
	return &Service{
		store: store,
	}
}
