package usersrv

import (
	"context"
	"errors"
	"strings"
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

func (u *UserSrv) GetOauthURL(providerName string) (string, error) {
	switch providerName {
	case googleProvider:
		return u.googleOauth.GetAuthURL(""), nil
	case githubProvider:
		return u.githubOauth.GetAuthURL(""), nil
	default:
		return "", ErrProviderNotSupported
	}
}

func (u *UserSrv) HandleOatuhLogin(
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
	case "google":
		userInfo, err = u.googleOauth.ExchangeCode(ctx, code)
	case "github":
		userInfo, err = u.githubOauth.ExchangeCode(ctx, code)
	default:
		return oauth.UserInfo{}, ErrProviderNotSupported
	}

	return userInfo, err
}

func getUsernameFromEmail(email string) string {
	p := strings.Split(email, "@")
	return p[0]
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
				Username:    getUsernameFromEmail(info.Email),
				Email:       info.Email,
				Activated:   true,
				Password:    "",
				CreatedAt:   time.Now(),
				LastLoginAt: time.Now(),
			})
			if cerr != nil {
				return uuid.Nil, cerr
			}
			return uid, nil
		}
		return uuid.Nil, err
	}

	return user.ID, nil
}
