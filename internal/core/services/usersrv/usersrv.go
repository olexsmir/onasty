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

	// TODO: handle if eamil already in use

	p, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return err
	}

	inp.Password = p

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
		Password: passwordHash,
	})
	if err != nil {
		return domain.UserTokens{}, err
	}

	tokens, err := s.getTokens(user.ID)
	if err != nil {
		return domain.UserTokens{}, err
	}
	err = s.store.SetSession(ctx, user.ID, tokens.Refresh, time.Now().Add(s.refreshTokenTTL))

	return domain.UserTokens{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}, err
}

func (s *Service) RefreshTokens(
	ctx context.Context,
	refreshToken string,
) (domain.UserTokens, error) {
	slog.With("refreshToken", refreshToken).Debug("user: refreshing tokens")
	user, err := s.store.GetUserByRefreshToken(ctx, refreshToken)
	if err != nil {
		return domain.UserTokens{}, err
	}

	tokens, err := s.getTokens(user.ID)
	if err != nil {
		return domain.UserTokens{}, err
	}

	slog.With("user", user).Debug("user: refreshing tokens")
	slog.With("tokens", tokens).Debug("user: refreshing tokens")

	slog.Debug("updating session")
	err = s.store.UpdateSession(ctx, user.ID, refreshToken, tokens.Refresh)
	slog.Debug("session updated")

	return tokens, err
}

func (s *Service) Logout(ctx context.Context, userId uuid.UUID) error {
	panic("not implemented") // TODO: Implement
}

func (s *Service) getTokens(userID uuid.UUID) (domain.UserTokens, error) {
	accessToken, err := s.tokeniser.GetToken(userID.String(), s.accessTokenTTL)
	if err != nil {
		return domain.UserTokens{}, err
	}

	refreshToken, err := s.tokeniser.GetRefreshToken()
	if err != nil {
		return domain.UserTokens{}, err
	}

	return domain.UserTokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}
