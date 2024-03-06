package usersrv

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/core/domain"
	"github.com/olexsmir/onasty/internal/ports"
)

var _ ports.UserServicer = (*Service)(nil)

type Service struct {
	store     ports.UserStorer
	hasher    ports.Hasher
	tokeniser ports.JWTTokenProvider

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func New(
	store ports.UserStorer,
	hasher ports.Hasher,
	tokeniser ports.JWTTokenProvider,
	accessTokenTTL, refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		store:           store,
		hasher:          hasher,
		tokeniser:       tokeniser,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
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
	passwordHash, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return domain.UserTokens{}, err
	}

	user, err := s.store.GetUserByCredentials(ctx, domain.UserCredentials{
		Email:    inp.Email,
		Password: string(passwordHash),
	})
	if err != nil {
		return domain.UserTokens{}, err
	}

	return s.getTokensAndSetSession(ctx, user.ID)
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

func (s *Service) getTokensAndSetSession(
	ctx context.Context,
	userID uuid.UUID,
) (domain.UserTokens, error) {
	accessToken, err := s.tokeniser.GetToken(userID.String(), s.accessTokenTTL)
	if err != nil {
		return domain.UserTokens{}, err
	}

	refreshToken, err := s.tokeniser.GetRefreshToken()
	if err != nil {
		return domain.UserTokens{}, err
	}

	err = s.store.SetSession(ctx, userID, refreshToken, time.Now().Add(s.refreshTokenTTL))

	return domain.UserTokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, err
}
