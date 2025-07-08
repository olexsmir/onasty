package e2e_test

import (
	"context"

	"github.com/olexsmir/onasty/internal/oauth"
)

var _ oauth.Provider = (*oauthProviderStub)(nil)

type oauthProviderStub struct{}

func newOauthProviderStub() *oauthProviderStub {
	return &oauthProviderStub{}
}

func (o *oauthProviderStub) GetAuthURL(_ string) string {
	return "https://example.com/oauth/authorize"
}

func (o *oauthProviderStub) ExchangeCode(_ context.Context, _ string) (oauth.UserInfo, error) {
	return oauth.UserInfo{
		Provider:      "google",
		ProviderID:    "1234567890",
		Email:         "testing@mail.org",
		EmailVerified: false,
	}, nil
}
