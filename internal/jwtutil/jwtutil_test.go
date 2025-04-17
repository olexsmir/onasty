package jwtutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTUtil_AccessToken(t *testing.T) {
	jwt := NewJWTUtil("key", time.Hour)
	payload := Payload{UserID: "user.123"}

	token, err := jwt.AccessToken(payload)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTUtil_RefreshToken(t *testing.T) {
	jwt := NewJWTUtil("key", time.Hour)

	tok, err := jwt.RefreshToken()
	require.NoError(t, err)
	assert.Len(t, tok, 64)

	secondTok, err := jwt.RefreshToken()
	require.NoError(t, err)

	// tokens should be unique
	assert.NotEqual(t, tok, secondTok)
}

func TestJWTUtil_Parse(t *testing.T) {
	jwt := NewJWTUtil("key", time.Hour)
	payload := Payload{UserID: "qwerty"}

	token, err := jwt.AccessToken(payload)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedPayload, err := jwt.Parse(token)
	require.NoError(t, err)

	assert.Equal(t, payload, parsedPayload)
}

func TestJWTUtil_Parse_expired(t *testing.T) {
	ttl := 100 * time.Millisecond
	jwt := NewJWTUtil("key", ttl)
	payload := Payload{UserID: "qwerty"}

	token, err := jwt.AccessToken(payload)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	time.Sleep(ttl)
	parsedPayload, err := jwt.Parse(token)
	require.Error(t, err)

	assert.Equal(t, payload, parsedPayload)
}
