package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var _ Provider = (*GitHubProvider)(nil)

const githubUserInfoEndpoint = "https://api.github.com/user"

type GitHubProvider struct {
	config oauth2.Config
}

func NewGithubProvider(clientID, secret, redirectURL string) GitHubProvider {
	return GitHubProvider{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: secret,
			RedirectURL:  redirectURL,
			Endpoint:     github.Endpoint,
			Scopes: []string{
				"user",
				"user:email",
			},
		},
	}
}

func (g GitHubProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state)
}

func (g GitHubProvider) ExchangeCode(ctx context.Context, code string) (UserInfo, error) {
	tok, err := g.config.Exchange(ctx, code)
	if err != nil {
		return UserInfo{}, err
	}

	client := g.config.Client(ctx, tok)
	resp, err := client.Get(githubUserInfoEndpoint)
	if err != nil {
		return UserInfo{}, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserInfo{}, err
	}

	var data struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&data); err != nil {
		return UserInfo{}, err
	}

	return UserInfo{
		Provider:      "github",
		ProviderID:    strconv.Itoa(data.ID),
		Email:         data.Email,
		EmailVerified: true,
	}, nil
}
