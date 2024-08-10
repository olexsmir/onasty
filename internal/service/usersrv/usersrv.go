package usersrv

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/mailer"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
)

type UserServicer interface {
	SignUp(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error)
	SignIn(ctx context.Context, inp dtos.SignInDTO) (dtos.TokensDTO, error)
	RefreshTokens(ctx context.Context, refreshToken string) (dtos.TokensDTO, error)
	Logout(ctx context.Context, userID uuid.UUID) error

	Verify(ctx context.Context, verificationKey string) error
	ResendVerificationEmail(ctx context.Context, credentials dtos.SignInDTO) error

	ParseJWTToken(token string) (jwtutil.Payload, error)

	CheckIfUserExists(ctx context.Context, userID uuid.UUID) (bool, error)
	CheckIfUserIsActivated(ctx context.Context, userID uuid.UUID) (bool, error)
}

var _ UserServicer = (*UserSrv)(nil)

type UserSrv struct {
	userstore    userepo.UserStorer
	sessionstore sessionrepo.SessionStorer
	vertokrepo   vertokrepo.VerificationTokenStorer
	hasher       hasher.Hasher
	jwtTokenizer jwtutil.JWTTokenizer
	mailer       mailer.Mailer

	refreshTokenTTL      time.Duration
	verificationTokenTTL time.Duration
}

func New(
	userstore userepo.UserStorer,
	sessionstore sessionrepo.SessionStorer,
	vertokrepo vertokrepo.VerificationTokenStorer,
	hasher hasher.Hasher,
	jwtTokenizer jwtutil.JWTTokenizer,
	mailer mailer.Mailer,
	refreshTokenTTL, verificationTokenTTL time.Duration,
) UserServicer {
	return &UserSrv{
		userstore:            userstore,
		sessionstore:         sessionstore,
		vertokrepo:           vertokrepo,
		hasher:               hasher,
		jwtTokenizer:         jwtTokenizer,
		mailer:               mailer,
		refreshTokenTTL:      refreshTokenTTL,
		verificationTokenTTL: verificationTokenTTL,
	}
}

func (u *UserSrv) SignUp(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error) {
	hashedPassword, err := u.hasher.Hash(inp.Password)
	if err != nil {
		return uuid.UUID{}, err
	}

	uid, err := u.userstore.Create(ctx, dtos.CreateUserDTO{
		Username:    inp.Username,
		Email:       inp.Email,
		Password:    hashedPassword,
		CreatedAt:   inp.CreatedAt,
		LastLoginAt: inp.LastLoginAt,
	})
	if err != nil {
		return uuid.Nil, err
	}

	vtok := uuid.Must(uuid.NewV4()).String()
	if err := u.vertokrepo.Create(ctx, vtok, uid, time.Now(), time.Now().Add(u.verificationTokenTTL)); err != nil {
		return uuid.Nil, err
	}

	// TODO: handle the error that might be returned
	// i dont think that tehre's need to handle the error, just log it
	bgCtx, bgCancel := context.WithTimeout(context.Background(), 10*time.Second)
	go u.sendVerificationEmail(bgCtx, bgCancel, inp.Email, vtok) //nolint:errcheck

	return uid, nil
}

func (u *UserSrv) SignIn(ctx context.Context, inp dtos.SignInDTO) (dtos.TokensDTO, error) {
	hashedPassword, err := u.hasher.Hash(inp.Password)
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	user, err := u.userstore.GetUserByCredentials(ctx, inp.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return dtos.TokensDTO{}, models.ErrUserWrongCredentials
		}
		return dtos.TokensDTO{}, err
	}

	if !user.Activated {
		return dtos.TokensDTO{}, models.ErrUserIsNotActivated
	}

	tokens, err := u.getTokens(user.ID)
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	if err := u.sessionstore.Set(ctx, user.ID, tokens.Refresh, time.Now().Add(u.refreshTokenTTL)); err != nil {
		return dtos.TokensDTO{}, err
	}

	return dtos.TokensDTO{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}, nil
}

func (u *UserSrv) Logout(ctx context.Context, userID uuid.UUID) error {
	return u.sessionstore.Delete(ctx, userID)
}

func (u *UserSrv) RefreshTokens(ctx context.Context, rtoken string) (dtos.TokensDTO, error) {
	userID, err := u.sessionstore.GetUserIDByRefreshToken(ctx, rtoken)
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	tokens, err := u.getTokens(userID)
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	if err := u.sessionstore.Update(ctx, userID, rtoken, tokens.Refresh); err != nil {
		return dtos.TokensDTO{}, err
	}

	return dtos.TokensDTO{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}, nil
}

func (u *UserSrv) Verify(ctx context.Context, verificationKey string) error {
	uid, err := u.vertokrepo.GetUserIDByTokenAndMarkAsUsed(ctx, verificationKey, time.Now())
	if err != nil {
		return err
	}

	return u.userstore.MarkUserAsActivated(ctx, uid)
}

func (u *UserSrv) ResendVerificationEmail(ctx context.Context, inp dtos.SignInDTO) error {
	hashedPassword, err := u.hasher.Hash(inp.Password)
	if err != nil {
		return err
	}

	user, err := u.userstore.GetUserByCredentials(ctx, inp.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.ErrUserWrongCredentials
		}
		return err
	}

	if user.Activated {
		return models.ErrUserIsAlreeadyVerified
	}

	token, err := u.vertokrepo.GetTokenOrUpdateTokenByUserID(
		ctx,
		user.ID,
		uuid.Must(uuid.NewV4()).String(),
		time.Now().Add(u.verificationTokenTTL))
	if err != nil {
		return err
	}

	bgCtx, bgCancel := context.WithTimeout(context.Background(), 10*time.Second)
	go u.sendVerificationEmail(bgCtx, bgCancel, inp.Email, token) //nolint:errcheck

	return nil
}

func (u *UserSrv) ParseJWTToken(token string) (jwtutil.Payload, error) {
	return u.jwtTokenizer.Parse(token)
}

func (u UserSrv) CheckIfUserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	return u.userstore.CheckIfUserExists(ctx, id)
}

func (u UserSrv) CheckIfUserIsActivated(ctx context.Context, userID uuid.UUID) (bool, error) {
	return u.userstore.CheckIfUserIsActivated(ctx, userID)
}

func (u UserSrv) getTokens(userID uuid.UUID) (dtos.TokensDTO, error) {
	accessToken, err := u.jwtTokenizer.AccessToken(jwtutil.Payload{UserID: userID.String()})
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	refreshToken, err := u.jwtTokenizer.RefreshToken()
	if err != nil {
		return dtos.TokensDTO{}, err
	}

	return dtos.TokensDTO{
		Access:  accessToken,
		Refresh: refreshToken,
	}, err
}
