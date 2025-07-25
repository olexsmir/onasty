package usersrv

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/olexsmir/onasty/internal/dtos"
	"github.com/olexsmir/onasty/internal/models"
	"github.com/olexsmir/onasty/internal/oauth"
)

var ErrProviderNotSupported = errors.New("oauth2 provider not supported")

const (
	googleProvider = "google"
	githubProvider = "github"
)

func (u *UserSrv) GetOAuthURL(providerName string) (dtos.OAuthRedirect, error) {
	state := uuid.Must(uuid.NewV4()).String()

	switch providerName {
	case googleProvider:
		return dtos.OAuthRedirect{
			URL:   u.googleOauth.GetAuthURL(state),
			State: state,
		}, nil
	case githubProvider:
		return dtos.OAuthRedirect{
			URL:   u.githubOauth.GetAuthURL(state),
			State: state,
		}, nil
	default:
		return dtos.OAuthRedirect{}, ErrProviderNotSupported
	}
}

func (u *UserSrv) HandleOAuthLogin(
	ctx context.Context,
	providerName, code string,
) (dtos.Tokens, error) {
	userInfo, err := u.getUserInfoBasedOnProvider(ctx, providerName, code)
	if err != nil {
		return dtos.Tokens{}, err
	}

	userID, err := u.getUserByOAuthIDOrCreateOne(ctx, userInfo)
	if err != nil {
		return dtos.Tokens{}, err
	}

	if err = u.userstore.LinkOAuthIdentity(ctx, userID, userInfo.Provider, userInfo.ProviderID); err != nil {
		slog.ErrorContext(ctx, "failed to link user identity", "user_id", userID, "err", err)
		return dtos.Tokens{}, err
	}

	tokens, err := u.issueTokens(ctx, userID)

	return tokens, err
}

func (u *UserSrv) getUserInfoBasedOnProvider(
	ctx context.Context,
	providerName, code string,
) (oauth.UserInfo, error) {
	var userInfo oauth.UserInfo
	var err error

	switch providerName {
	case googleProvider:
		userInfo, err = u.googleOauth.ExchangeCode(ctx, code)
	case githubProvider:
		userInfo, err = u.githubOauth.ExchangeCode(ctx, code)
	default:
		return oauth.UserInfo{}, ErrProviderNotSupported
	}

	return userInfo, err
}

func (u *UserSrv) getUserByOAuthIDOrCreateOne(
	ctx context.Context,
	info oauth.UserInfo,
) (uuid.UUID, error) {
	user, err := u.userstore.GetByOAuthID(ctx, info.Provider, info.ProviderID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			uid, cerr := u.userstore.Create(ctx, models.User{
				ID:          uuid.Nil,
				Email:       info.Email,
				Activated:   true,
				Password:    "",
				CreatedAt:   time.Now(),
				LastLoginAt: time.Now(),
			})
			return uid, cerr
		}
		return uuid.Nil, err
	}

	return user.ID, nil
}
