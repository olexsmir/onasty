package usersrv

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
)

type UserServicer interface {
	SignUp(ctx context.Context, inp SignUpInput) (uuid.UUID, error)
}

type UserSrv struct {
	store  userepo.UserStorer
	hasher *hasher.SHA256Hasher
}

func New(store userepo.UserStorer, hasher *hasher.SHA256Hasher) UserServicer {
	return &UserSrv{
		store:  store,
		hasher: hasher,
	}
}

type SignUpInput struct {
	Username    string
	Email       string
	Password    string
	CreatedAt   time.Time
	LastLoginAt time.Time
}

func (s *UserSrv) SignUp(ctx context.Context, inp SignUpInput) (uuid.UUID, error) {
	hashedPassword, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return uuid.UUID{}, err
	}

	return s.store.Create(ctx, userepo.CreateInput{
		Username:    inp.Username,
		Email:       inp.Email,
		Password:    hashedPassword,
		CreatedAt:   inp.CreatedAt,
		LastLoginAt: inp.LastLoginAt,
	})
}
