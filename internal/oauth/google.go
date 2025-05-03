package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var _ Provider = (*GoogleProvider)(nil)

const googleUserInfoEndpoint = "https://www.googleapis.com/oauth2/v3/userinfo"

type GoogleProvider struct {
	config oauth2.Config
}

func NewGoogleProvider(clientID, secret, redirectURL string) GoogleProvider {
	return GoogleProvider{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: secret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
			},
		},
	}
}

func (g GoogleProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

func (g GoogleProvider) ExchangeCode(ctx context.Context, code string) (UserInfo, error) {
	tok, err := g.config.Exchange(ctx, code)
	if err != nil {
		return UserInfo{}, err
	}

	client := g.config.Client(ctx, tok)
	resp, err := client.Get(googleUserInfoEndpoint)
	if err != nil {
		return UserInfo{}, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserInfo{}, err
	}

	var data struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}

	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&data); err != nil {
		return UserInfo{}, err
	}

	return UserInfo{
		Provider:      "google",
		ProviderID:    data.Sub,
		Email:         data.Email,
		EmailVerified: data.EmailVerified,
	}, nil
}
