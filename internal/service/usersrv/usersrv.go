package usersrv

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/events/mailermq"
	"github.com/olexsmir/onasty/internal/hasher"
	"github.com/olexsmir/onasty/internal/jwtutil"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
	"github.com/olexsmir/onasty/internal/store/rdb/usercache"
)

type UserServicer interface {
	SignUp(ctx context.Context, inp dtos.CreateUserDTO) (uuid.UUID, error)
	SignIn(ctx context.Context, inp dtos.SignInDTO) (dtos.TokensDTO, error)
	RefreshTokens(ctx context.Context, refreshToken string) (dtos.TokensDTO, error)
	Logout(ctx context.Context, userID uuid.UUID) error

	ChangePassword(ctx context.Context, userID uuid.UUID, inp dtos.ResetUserPasswordDTO) error

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
	mailermq     mailermq.Mailer
	cache        usercache.UserCacheer

	refreshTokenTTL      time.Duration
	verificationTokenTTL time.Duration
	appURL               string
}

func New(
	userstore userepo.UserStorer,
	sessionstore sessionrepo.SessionStorer,
	vertokrepo vertokrepo.VerificationTokenStorer,
	hasher hasher.Hasher,
	jwtTokenizer jwtutil.JWTTokenizer,
	mailermq mailermq.Mailer,
	cache usercache.UserCacheer,
	refreshTokenTTL, verificationTokenTTL time.Duration,
	appURL string,
) *UserSrv {
	return &UserSrv{
		userstore:            userstore,
		sessionstore:         sessionstore,
		vertokrepo:           vertokrepo,
		hasher:               hasher,
		jwtTokenizer:         jwtTokenizer,
		mailermq:             mailermq,
		cache:                cache,
		refreshTokenTTL:      refreshTokenTTL,
		verificationTokenTTL: verificationTokenTTL,
		appURL:               appURL,
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

	if err := u.mailermq.SendVerificationEmail(ctx, mailermq.SendVerificationEmailRequest{
		Receiver: inp.Email,
		Token:    vtok,
	}); err != nil {
		return uuid.Nil, err
	}

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

func (u *UserSrv) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	inp dtos.ResetUserPasswordDTO,
) error {
	oldPass, err := u.hasher.Hash(inp.CurrentPassword)
	if err != nil {
		return err
	}

	newPass, err := u.hasher.Hash(inp.NewPassword)
	if err != nil {
		return err
	}

	if err := u.userstore.ChangePassword(ctx, userID, oldPass, newPass); err != nil {
		return err
	}

	return nil
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

	if err := u.mailermq.SendVerificationEmail(ctx, mailermq.SendVerificationEmailRequest{
		Receiver: inp.Email,
		Token:    token,
	}); err != nil {
		return err
	}

	return nil
}

func (u *UserSrv) ParseJWTToken(token string) (jwtutil.Payload, error) {
	return u.jwtTokenizer.Parse(token)
}

func (u UserSrv) CheckIfUserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if r, err := u.cache.GetIsExists(ctx, id.String()); err == nil {
		return r, nil
	} else { //nolint:revive
		slog.ErrorContext(ctx, "usercache", "err", err)
	}

	isExists, err := u.userstore.CheckIfUserExists(ctx, id)
	if err != nil {
		return false, err
	}

	if err := u.cache.SetIsExists(ctx, id.String(), isExists); err != nil {
		slog.Error("usercache", "err", err)
	}

	return isExists, nil
}

func (u UserSrv) CheckIfUserIsActivated(ctx context.Context, userID uuid.UUID) (bool, error) {
	if r, err := u.cache.GetIsActivated(ctx, userID.String()); err == nil {
		return r, nil
	} else { //nolint:revive
		slog.ErrorContext(ctx, "usercache", "err", err)
	}

	isActivated, err := u.userstore.CheckIfUserExists(ctx, userID)
	if err != nil {
		return false, err
	}

	if err := u.cache.SetIsActivated(ctx, userID.String(), isActivated); err != nil {
		slog.Error("usercache", "err", err)
	}

	return isActivated, nil
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
