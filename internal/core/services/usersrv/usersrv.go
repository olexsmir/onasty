package usersrv

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/core/domain"
	"github.com/olexsmir/onasty/internal/ports"
)

var _ ports.UserServicer = (*Service)(nil)

type Service struct {
	store  ports.UserStorer
	hasher ports.Hasher
}

func New(store ports.UserStorer, hasher ports.Hasher) *Service {
	return &Service{
		store:  store,
		hasher: hasher,
	}
}

func (s *Service) SignUp(
	ctx context.Context,
	inp domain.User,
) error {
	// FIXME: dont log password
	slog.With("inp", inp).Info("user: signing up")

	if err := inp.Validate(); err != nil {
		return err
	}

	slog.Info("user: hashing password")
	p, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return err
	}

	inp.Password = string(p)

	return s.store.Create(ctx, inp)
}

func (s *Service) SignIn(
	ctx context.Context,
	inp domain.User,
) (domain.UserTokens, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Service) RefreshTokens(
	ctx context.Context,
	refreshToken string,
) (domain.UserTokens, error) {
	panic("not implemented") // TODO: Implement
}

func (s *Service) Logout(ctx context.Context, userId uuid.UUID) error {
	panic("not implemented") // TODO: Implement
}
