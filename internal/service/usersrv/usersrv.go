package usersrv

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
)

type UserServicer interface {
	SignUp(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error)
	SignIn(ctx context.Context, inp dtos.SignInDTO) (dtos.TokensDTO, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	ParseToken(token string) (jwtutil.Payload, error)
}

var _ UserServicer = (*UserSrv)(nil)

type UserSrv struct {
	userstore    userepo.UserStorer
	sessionstore sessionrepo.SessionStorer
	hasher       *hasher.SHA256Hasher
	jwtTokenizer jwtutil.JWTTokenizer

	refreshTokenExpiredAt time.Time
}

func New(
	userstore userepo.UserStorer,
	sessionstore sessionrepo.SessionStorer,
	hasher *hasher.SHA256Hasher,
	jwtTokenizer jwtutil.JWTTokenizer,
) UserServicer {
	return &UserSrv{
		userstore:    userstore,
		sessionstore: sessionstore,
		hasher:       hasher,
		jwtTokenizer: jwtTokenizer,
	}
}

func (u *UserSrv) SignUp(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error) {
	hashedPassword, err := u.hasher.Hash(inp.Password)
	if err != nil {
		return uuid.UUID{}, err
	}

	return u.userstore.Create(ctx, dtos.CreateUserDTO{
		Username:    inp.Username,
		Email:       inp.Email,
		Password:    hashedPassword,
		CreatedAt:   inp.CreatedAt,
		LastLoginAt: inp.LastLoginAt,
	})
}

func (u *UserSrv) SignIn(ctx context.Context, inp dtos.SignInDTO) (dtos.TokensDTO, error) {
	hashedPassword, err := u.hasher.Hash(inp.Password)
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	user, err := u.userstore.GetUserByCredentials(ctx, inp.Email, hashedPassword)
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	accessToken, err := u.jwtTokenizer.AccessToken(jwtutil.Payload{UserID: user.ID.String()})
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	refreshToken, err := u.jwtTokenizer.RefreshToken()
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	if err := u.sessionstore.Set(ctx, user.ID, refreshToken, u.refreshTokenExpiredAt); err != nil {
		return dtos.TokensDTO{}, err
	}

	return dtos.TokensDTO{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (u *UserSrv) ParseToken(token string) (jwtutil.Payload, error) {
	return u.jwtTokenizer.Parse(token)
}

func (u *UserSrv) Logout(ctx context.Context, userID uuid.UUID) error {
	return u.sessionstore.Delete(ctx, userID)
}
