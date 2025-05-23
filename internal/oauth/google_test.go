package oauth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGoogleProvider_GetAuthURL(t *testing.T) {
	provider := NewGoogleProvider("client.id", "secret", "http://localhost/callback")
	authURL := provider.GetAuthURL("test")

	assert.Contains(t, authURL, "client_id=client.id")
	assert.Contains(t, authURL, "state=test")
	assert.Contains(t, authURL, "scope="+
		url.QueryEscape("https://www.googleapis.com/auth/userinfo.email"))
}

func TestGoogleProvider_ExchangeCode(t *testing.T) {
	sub := "1234567890"
	email := "testemail@mail.com"
	resp := fmt.Sprintf(`{"sub":"%s", "email":"%s","email_verified":true}`, sub, email)
	client := &http.Client{
		Transport: mockClient(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body: io.NopCloser(
						strings.NewReader(`{"access_token":"fake",
							"token_type":"bearer",
							"expires_in":3600}`),
					),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(resp)),
			}, nil
		}),
	}

	provider := NewGoogleProvider("client.id", "secret", "http://localhost")
	ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, client)

	info, err := provider.ExchangeCode(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, "google", info.Provider)
	assert.Equal(t, sub, info.ProviderID)
	assert.Equal(t, email, info.Email)
	assert.True(t, info.EmailVerified)
}
