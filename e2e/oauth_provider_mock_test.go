package e2e_test

import (
	"context"

	"github.com/olexsmir/onasty/internal/oauth"
)

var _ oauth.Provider = (*oauthProviderMock)(nil)

type oauthProviderMock struct{}

func newOauthProviderMock() *oauthProviderMock {
	return &oauthProviderMock{}
}

func (o *oauthProviderMock) GetAuthURL(_ string) string {
	return "https://example.com/oauth/authorize"
}

func (o *oauthProviderMock) ExchangeCode(_ context.Context, _ string) (oauth.UserInfo, error) {
	return oauth.UserInfo{
		Provider:   "google",
		ProviderID: "1234567890",
	}, nil
}
