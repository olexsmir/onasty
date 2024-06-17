package usersrv

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
)

type UserServicer interface {
	SignUp(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error)
	SignIn(ctx context.Context, inp dtos.SignInDTO) (dtos.TokensDTO, error)
	ParseToken(token string) (jwtutil.Payload, error)
}

type UserSrv struct {
	store        userepo.UserStorer
	hasher       *hasher.SHA256Hasher
	jwtTokenizer jwtutil.JWTTokenizer
}

func New(
	store userepo.UserStorer,
	hasher *hasher.SHA256Hasher,
	jwtTokenizer jwtutil.JWTTokenizer,
) UserServicer {
	return &UserSrv{
		store:        store,
		hasher:       hasher,
		jwtTokenizer: jwtTokenizer,
	}
}

func (s *UserSrv) SignUp(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error) {
	hashedPassword, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return uuid.UUID{}, err
	}

	return s.store.Create(ctx, dtos.CreateUserDTO{
		Username:    inp.Username,
		Email:       inp.Email,
		Password:    hashedPassword,
		CreatedAt:   inp.CreatedAt,
		LastLoginAt: inp.LastLoginAt,
	})
}

func (s *UserSrv) SignIn(ctx context.Context, inp dtos.SignInDTO) (dtos.TokensDTO, error) {
	hashedPassword, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	user, err := s.store.GetUserByCredentials(ctx, inp.Email, hashedPassword)
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	accessToken, err := s.jwtTokenizer.AccessToken(jwtutil.Payload{UserID: user.ID.String()})
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	refreshToken, err := s.jwtTokenizer.RefreshToken()
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	return dtos.TokensDTO{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *UserSrv) ParseToken(token string) (jwtutil.Payload, error) {
	return s.jwtTokenizer.Parse(token)
}
