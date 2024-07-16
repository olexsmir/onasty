package jwtutil

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenizer interface {
	// AccessToken generates a new access token with the given payload
	AccessToken(pl Payload) (string, error)

	// RefreshToken generates a new refresh token
	RefreshToken() (string, error)

	// Parse parses the token and returns the payload
	Parse(token string) (Payload, error)
}

type Payload struct {
	UserID string
}

var _ JWTTokenizer = (*JWTUtil)(nil)

type JWTUtil struct {
	signingKey     string
	accessTokenTTL time.Duration
}

func NewJWTUtil(signingKey string, accessTokenTTL time.Duration) *JWTUtil {
	return &JWTUtil{
		signingKey:     signingKey,
		accessTokenTTL: accessTokenTTL,
	}
}

func (j *JWTUtil) AccessToken(pl Payload) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   pl.UserID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
	})
	return tok.SignedString([]byte(j.signingKey))
}

func (j *JWTUtil) RefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (j *JWTUtil) Parse(token string) (Payload, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.signingKey), nil
	})
	return Payload{
		UserID: claims.Subject,
	}, err
}
