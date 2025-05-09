package oauth

import "context"

// Provider is an OAuth interface.
type Provider interface {
	// GetAuthURL return the provider's authorization page URL.
	GetAuthURL(state string) string

	// ExchangeCode exchanges the provided authorization code for user information.
	ExchangeCode(ctx context.Context, code string) (UserInfo, error)
}

// UserInfo represents the user information returned by the OAuth provider.
type UserInfo struct {
	// Provider is the name of the OAuth provider
	Provider string
	// ProviderID is user ID assigned by the provider
	ProviderID string
	// Email is user's email address returned by the provider
	Email string
	// EmailVerified indicates whether the email was verified by the provider
	EmailVerified bool
}
