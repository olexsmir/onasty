package authsrv

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

type AuthServicer interface {
	// SignUp creates a new user and sends a verification email.
	//
	// Uses [models.User.Validate] to validate credentials (see more possible returned errors).
	//
	// If provided email already in use returns [models.ErrUserEmailIsAlreadyInUse].
	//
	SignUp(ctx context.Context, credentials dtos.SignUp) error

	// SignIn authenticates a user and returns access and refresh tokens.
	//
	// If user not found returns [models.ErrUserNotFound], and if credentials don't match [models.ErrUserWrongCredentials]
	//
	// If inactivated user tries to login, returns [models.ErrUserIsNotActivated]
	//
	SignIn(ctx context.Context, credentials dtos.SignIn) (dtos.Tokens, error)

	// RefreshTokens refreshes the access and refresh tokens using the provided refresh token.
	//
	// If couldn't find a user liked with token, returns [models.ErrUserNotFound]
	//
	RefreshTokens(ctx context.Context, refreshToken string) (dtos.Tokens, error)

	// Logout logs out a user by deleting the session associated with the provided refresh token.
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error

	// LogoutAll logs out a user by deleting all sessions associated with the user ID.
	LogoutAll(ctx context.Context, userID uuid.UUID) error

	// GetOAuthURL retrieves the OAuth URL for the specified provider.
	//
	// If [providerName] is incorrect returns [ErrProviderNotSupported]
	//
	GetOAuthURL(providerName string) (dtos.OAuthRedirect, error)

	// HandleOAuthLogin handles the OAuth login process by exchanging the code for tokens.
	//
	HandleOAuthLogin(ctx context.Context, providerName, code string) (dtos.Tokens, error)

	// ParseJWTToken parses the JWT token and returns the payload.
	//
	// If token is expired, returns [jwtutil.ErrTokenExpired],
	//
	// If token is invalid returns: [jwturil.ErrTokenSignatureInvalid], [jwt.ErrUnexpectedSigningMethod]
	//
	ParseJWTToken(token string) (jwtutil.Payload, error)

	// CheckIfUserExists checks if a user exists by user ID.
	CheckIfUserExists(ctx context.Context, userID uuid.UUID) (bool, error)

	// CheckIfUserIsActivated checks if a user is activated by user ID.
	CheckIfUserIsActivated(ctx context.Context, userID uuid.UUID) (bool, error)
}

var _ AuthServicer = (*AuthSrv)(nil)

type AuthSrv struct {
	userstore    userepo.UserStorer
	sessionstore sessionrepo.SessionStorer
	vertokrepo   vertokrepo.VerificationTokenStorer
	cache        usercache.UserCacheer

	hasher       hasher.Hasher
	jwtTokenizer jwtutil.JWTTokenizer
	mailermq     mailermq.Mailer

	googleOauth oauth.Provider
	githubOauth oauth.Provider

	refreshTokenTTL      time.Duration
	verificationTokenTTL time.Duration
}

func New(
	userstore userepo.UserStorer,
	sessionstore sessionrepo.SessionStorer,
	vertokrepo vertokrepo.VerificationTokenStorer,
	cache usercache.UserCacheer,
	hasher hasher.Hasher,
	jwtTokenizer jwtutil.JWTTokenizer,
	mailermq mailermq.Mailer,
	googleOauth, githubOauth oauth.Provider,
	refreshTokenTTL, verificationTokenTTL time.Duration,
) *AuthSrv {
	return &AuthSrv{
		userstore:            userstore,
		sessionstore:         sessionstore,
		vertokrepo:           vertokrepo,
		cache:                cache,
		hasher:               hasher,
		jwtTokenizer:         jwtTokenizer,
		mailermq:             mailermq,
		googleOauth:          googleOauth,
		githubOauth:          githubOauth,
		refreshTokenTTL:      refreshTokenTTL,
		verificationTokenTTL: verificationTokenTTL,
	}
}

func (a *AuthSrv) SignUp(ctx context.Context, inp dtos.SignUp) error {
	user := models.User{
		ID:          uuid.Nil, // nil, since we do not know it yet
		Email:       inp.Email,
		Activated:   false,
		Password:    inp.Password,
		CreatedAt:   inp.CreatedAt,
		LastLoginAt: inp.LastLoginAt,
	}
	if err := user.Validate(); err != nil {
		return err
	}

	hashedPassword, err := a.hasher.Hash(inp.Password)
	if err != nil {
		return err
	}

	user.Password = hashedPassword

	userID, err := a.userstore.Create(ctx, user)
	if err != nil {
		return err
	}

	verificationToken := uuid.Must(uuid.NewV4()).String()
	if err := a.vertokrepo.Create(ctx, models.VerificationToken{
		UserID:    userID,
		Token:     verificationToken,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(a.verificationTokenTTL),
	}); err != nil {
		return err
	}

	if err := a.mailermq.SendVerificationEmail(ctx, mailermq.SendVerificationEmailRequest{
		Receiver: inp.Email,
		Token:    verificationToken,
	}); err != nil {
		return err
	}

	return nil
}

func (a *AuthSrv) SignIn(ctx context.Context, inp dtos.SignIn) (dtos.Tokens, error) {
	user, err := a.userstore.GetByEmail(ctx, inp.Email)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err = a.hasher.Compare(user.Password, inp.Password); err != nil {
		if errors.Is(err, hasher.ErrMismatchedHashes) {
			return dtos.Tokens{}, models.ErrUserWrongCredentials
		}
		return dtos.Tokens{}, err
	}

	if !user.IsActivated() {
		return dtos.Tokens{}, models.ErrUserIsNotActivated
	}

	return a.issueTokens(ctx, user.ID)
}

func (a *AuthSrv) RefreshTokens(ctx context.Context, rtoken string) (dtos.Tokens, error) {
	userID, err := a.sessionstore.GetUserIDByRefreshToken(ctx, rtoken)
	if err != nil {
		return dtos.Tokens{}, err
	}

	tokens, err := a.createTokens(userID)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err := a.sessionstore.Update(ctx, userID, rtoken, tokens.Refresh); err != nil {
		return dtos.Tokens{}, err
	}

	return dtos.Tokens{
		Access:  tokens.Access,
		Refresh: tokens.Refresh,
	}, nil
}

func (a *AuthSrv) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	return a.sessionstore.Delete(ctx, userID, refreshToken)
}

func (a *AuthSrv) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return a.sessionstore.DeleteAllByUserID(ctx, userID)
}

func (a *AuthSrv) CheckIfUserExists(ctx context.Context, uid uuid.UUID) (bool, error) {
	if isExists, err := a.cache.GetIsExists(ctx, uid.String()); err == nil {
		return isExists, nil
	}
	isExists, err := a.userstore.CheckIfUserExists(ctx, uid)
	if err != nil {
		return false, err
	}

	if err := a.cache.SetIsExists(ctx, uid.String(), isExists); err != nil {
		slog.ErrorContext(ctx, "failed to update 'is user exists' cache", "err", err)
	}

	return isExists, nil
}

func (a *AuthSrv) CheckIfUserIsActivated(ctx context.Context, uid uuid.UUID) (bool, error) {
	if isActivated, err := a.cache.GetIsActivated(ctx, uid.String()); err == nil {
		return isActivated, nil
	}

	isActivated, err := a.userstore.CheckIfUserIsActivated(ctx, uid)
	if err != nil {
		return false, err
	}

	if err := a.cache.SetIsActivated(ctx, uid.String(), isActivated); err != nil {
		slog.ErrorContext(ctx, "failed to update 'is user activated' cache", "err", err)
	}

	return isActivated, nil
}
