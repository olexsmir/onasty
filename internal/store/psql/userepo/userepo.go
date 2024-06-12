package userepo

import (
	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/store/psqlutil"
)

type UserStorer interface {
	SignUp(inp SignUpInput) (uuid.UUID, error)
}

type UserRepo struct {
	db *psqlutil.DB
}

func New(db *psqlutil.DB) UserStorer {
	return &UserRepo{
		db: db,
	}
}

type SignUpInput struct{}

func (r *UserRepo) SignUp(_ SignUpInput) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}
