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

func (u *UserSrv) GetOauthURL(_ context.Context, providerName string) (string, error) {
	switch providerName {
	case "google":
		return u.googleOauth.GetAuthURL("randomstate"), nil
	case "github":
		fallthrough
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
		fallthrough
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
			us, err := u.userstore.Create(ctx, models.User{
				ID:          uuid.Nil, // nil, because it does not get used here
				Username:    getUsernameFromEmail(info.Email),
				Email:       info.Email,
				Activated:   true,
				Password:    "",
				CreatedAt:   time.Now(),
				LastLoginAt: time.Now(),
			})
			if err != nil {
				return uuid.Nil, err
			}
			user.ID = us
		}
	}

	return user.ID, nil
}
