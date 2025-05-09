package oauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGitHubProvider_GetAuthURL(t *testing.T) {
	provider := NewGithubProvider("client.id", "secret", "http://localhost/callback")
	url := provider.GetAuthURL("test")

	assert.Contains(t, url, "client_id=client.id")
	assert.Contains(t, url, "state=test")
	assert.Contains(t, url, "scope=user%3Aemail")
}

type mockClient func(*http.Request) (*http.Response, error)

func (m mockClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

func TestGitHubProvider_ExchangeCode(t *testing.T) {
	userID := "123123"
	userEmail := "test@testing.org"
	userLogin := "testing"

	resp := fmt.Sprintf(`{"id":%s, "email":"%s", "login":"%s"}`, userID, userEmail, userLogin)
	client := &http.Client{ //nolint:exhaustruct
		Transport: mockClient(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost {
				return &http.Response{ //nolint:exhaustruct
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body: io.NopCloser(
						strings.NewReader(`{"access_token":"fake",
							"token_type":"bearer",
							"expires_in":3600}`),
					),
				}, nil
			}
			return &http.Response{ //nolint:exhaustruct
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(resp)),
			}, nil
		}),
	}

	provider := NewGithubProvider("client.id", "secret", "http://localhost")
	ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, client)

	info, err := provider.ExchangeCode(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, "github", info.Provider)
	assert.Equal(t, userID, info.ProviderID)
	assert.Equal(t, userEmail, info.Email)
}

func TestGitHubProvider_ExchangeCode_tokenExcahnge_error(t *testing.T) {
	client := &http.Client{ //nolint:exhaustruct
		Transport: mockClient(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPost {
				return &http.Response{ //nolint:exhaustruct
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}
			return nil, errors.New("unexpected request")
		}),
	}

	provider := NewGithubProvider("client.id", "secret", "http://localhost")
	ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, client)

	_, err := provider.ExchangeCode(ctx, "")
	require.Error(t, err)
}
