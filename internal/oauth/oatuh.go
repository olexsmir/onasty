package oauth

import "context"

type Provider interface {
	// GetAuthURL generates the URL to redirect users to for login.
	GetAuthURL(state string) string

	// ExchangeCode exchanges the authorization code for an access token and ID token.
	ExchangeCode(ctx context.Context, code string) (UserInfo, error)
}

type UserInfo struct {
	Provider      string
	ProviderID    string
	Email         string
	EmailVerified bool
}
