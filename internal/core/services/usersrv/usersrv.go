package usersrv

import (
	"context"

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
	panic("not implemented") // TODO: Implement
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
