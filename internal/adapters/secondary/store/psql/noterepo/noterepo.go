package noterepo

import (
	"github.com/olexsmir/onasty/internal/adapters/secondary/store/psql"
	"github.com/olexsmir/onasty/internal/ports"
)

var _ ports.NoteStorer = (*Store)(nil)

type Store struct {
	db *psql.DB
}

func New(db *psql.DB) *Store {
	return &Store{
		db: db,
	}
}
