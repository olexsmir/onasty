package oauth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var _ Provider = (*GoogleProvider)(nil)

type GoogleProvider struct {
	Config oauth2.Config
}

func NewGoogleProvider(clientID, secret, redirectURL string) GoogleProvider {
	return GoogleProvider{
		Config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: secret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		},
	}
}

func (g GoogleProvider) GetAuthURL(state string) string {
	return g.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g GoogleProvider) ExchangeCode(ctx context.Context, code string) (UserInfo, error) {
	return UserInfo{}, nil
}
