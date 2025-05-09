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
	"github.com/olexsmir/onasty/internal/oauth"
	"github.com/olexsmir/onasty/internal/store/psql/sessionrepo"
	"github.com/olexsmir/onasty/internal/store/psql/userepo"
	"github.com/olexsmir/onasty/internal/store/psql/vertokrepo"
	"github.com/olexsmir/onasty/internal/store/rdb/usercache"
)

type UserServicer interface {
	SignUp(ctx context.Context, inp dtos.SignUp) (uuid.UUID, error)
	SignIn(ctx context.Context, inp dtos.SignIn) (dtos.Tokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (dtos.Tokens, error)
	Logout(ctx context.Context, userID uuid.UUID) error

	ChangePassword(ctx context.Context, userID uuid.UUID, inp dtos.ChangeUserPassword) error

	GetOAuthURL(providerName string) (string, error)
	HandleOAuthLogin(ctx context.Context, providerName, code string) (dtos.Tokens, error)

	Verify(ctx context.Context, verificationKey string) error
	ResendVerificationEmail(ctx context.Context, credentials dtos.SignIn) error

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
	googleOauth  oauth.Provider
	githubOauth  oauth.Provider

	refreshTokenTTL      time.Duration
	verificationTokenTTL time.Duration
}

func New(
	userstore userepo.UserStorer,
	sessionstore sessionrepo.SessionStorer,
	vertokrepo vertokrepo.VerificationTokenStorer,
	hasher hasher.Hasher,
	jwtTokenizer jwtutil.JWTTokenizer,
	mailermq mailermq.Mailer,
	cache usercache.UserCacheer,
	googleOauth, githubOauth oauth.Provider,
	refreshTokenTTL, verificationTokenTTL time.Duration,
) *UserSrv {
	return &UserSrv{
		userstore:            userstore,
		sessionstore:         sessionstore,
		vertokrepo:           vertokrepo,
		hasher:               hasher,
		jwtTokenizer:         jwtTokenizer,
		mailermq:             mailermq,
		cache:                cache,
		googleOauth:          googleOauth,
		githubOauth:          githubOauth,
		refreshTokenTTL:      refreshTokenTTL,
		verificationTokenTTL: verificationTokenTTL,
	}
}

func (u *UserSrv) SignUp(ctx context.Context, inp dtos.SignUp) (uuid.UUID, error) {
	hashedPassword, err := u.hasher.Hash(inp.Password)
	if err != nil {
		return uuid.UUID{}, err
	}

	user := models.User{
		ID:          uuid.Nil, // nil, because it does not get used here
		Username:    inp.Username,
		Email:       inp.Email,
		Activated:   false,
		Password:    hashedPassword,
		CreatedAt:   inp.CreatedAt,
		LastLoginAt: inp.LastLoginAt,
	}
	if err = user.Validate(); err != nil {
		return uuid.Nil, err
	}

	userID, err := u.userstore.Create(ctx, user)
	if err != nil {
		return uuid.Nil, err
	}

	verificationToken := uuid.Must(uuid.NewV4()).String()
	if err := u.vertokrepo.Create(
		ctx,
		verificationToken,
		userID,
		time.Now(),
		time.Now().Add(u.verificationTokenTTL),
	); err != nil {
		return uuid.Nil, err
	}

	if err := u.mailermq.SendVerificationEmail(ctx, mailermq.SendVerificationEmailRequest{
		Receiver: inp.Email,
		Token:    verificationToken,
	}); err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (u *UserSrv) SignIn(ctx context.Context, inp dtos.SignIn) (dtos.Tokens, error) {
	user, err := u.userstore.GetByEmail(ctx, inp.Email)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err = u.hasher.Compare(user.Password, inp.Password); err != nil {
		if errors.Is(err, hasher.ErrMismatchedHashes) {
			return dtos.Tokens{}, models.ErrUserWrongCredentials
		}
		return dtos.Tokens{}, err
	}

	if !user.IsActivated() {
		return dtos.Tokens{}, models.ErrUserIsNotActivated
	}

	tokens, err := u.issueTokens(ctx, user.ID)
	return tokens, err
}

func (u *UserSrv) Logout(ctx context.Context, userID uuid.UUID) error {
	return u.sessionstore.Delete(ctx, userID)
}

func (u *UserSrv) RefreshTokens(ctx context.Context, rtoken string) (dtos.Tokens, error) {
	userID, err := u.sessionstore.GetUserIDByRefreshToken(ctx, rtoken)
	if err != nil {
		return dtos.Tokens{}, err
	}

	tokens, err := u.createTokens(userID)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err := u.sessionstore.Update(ctx, userID, rtoken, tokens.Refresh); err != nil {
		return dtos.Tokens{}, err
	}

	return dtos.Tokens{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}, nil
}

func (u *UserSrv) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	inp dtos.ChangeUserPassword,
) error {
	// TODO: compare current password with providede, and assert on mismatch

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

func (u *UserSrv) ResendVerificationEmail(ctx context.Context, inp dtos.SignIn) error {
	user, err := u.userstore.GetByEmail(ctx, inp.Email)
	if err != nil {
		return err
	}

	if err = u.hasher.Compare(user.Password, inp.Password); err != nil {
		return models.ErrUserWrongCredentials
	}

	if user.Activated {
		return models.ErrUserIsAlreadyVerified
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
	r, err := u.cache.GetIsExists(ctx, id.String())
	if err == nil {
		return r, nil
	}

	slog.ErrorContext(ctx, "usercache", "err", err)

	isExists, err := u.userstore.CheckIfUserExists(ctx, id)
	if err != nil {
		return false, err
	}

	if err := u.cache.SetIsExists(ctx, id.String(), isExists); err != nil {
		slog.ErrorContext(ctx, "usercache", "err", err)
	}

	return isExists, nil
}

func (u *UserSrv) CheckIfUserIsActivated(ctx context.Context, userID uuid.UUID) (bool, error) {
	r, err := u.cache.GetIsActivated(ctx, userID.String())
	if err == nil {
		return r, nil
	}

	slog.ErrorContext(ctx, "usercache", "err", err)

	isActivated, err := u.userstore.CheckIfUserIsActivated(ctx, userID)
	if err != nil {
		return false, err
	}

	if err := u.cache.SetIsActivated(ctx, userID.String(), isActivated); err != nil {
		slog.ErrorContext(ctx, "usercache", "err", err)
	}

	return isActivated, nil
}

func (u UserSrv) createTokens(userID uuid.UUID) (dtos.Tokens, error) {
	accessToken, err := u.jwtTokenizer.AccessToken(jwtutil.Payload{UserID: userID.String()})
	if err != nil {
		return dtos.Tokens{}, err
	}

	refreshToken, err := u.jwtTokenizer.RefreshToken()
	if err != nil {
		return dtos.Tokens{}, err
	}

	return dtos.Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, err
}

func (u UserSrv) issueTokens(ctx context.Context, userID uuid.UUID) (dtos.Tokens, error) {
	toks, err := u.createTokens(userID)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err := u.sessionstore.Set(ctx, userID, toks.Refresh, time.Now().Add(u.refreshTokenTTL)); err != nil {
		return dtos.Tokens{}, err
	}

	return toks, nil
}
