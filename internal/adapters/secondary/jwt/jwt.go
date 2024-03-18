package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/olexsmir/onasty/internal/ports"
)

var (
	ErrUnexpectedSingningMethod = errors.New("unexpected signing method")
	ErrCannotParseClaims        = errors.New("cannot parse claims")
)

var _ ports.JWTTokenProvider = (*Tokens)(nil)

type Tokens struct {
	signingKey string
}

func New(key string) *Tokens {
	return &Tokens{signingKey: key}
}

func (t *Tokens) GetToken(subject string, ttl time.Duration) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		Subject:   subject,
	})
	return tok.SignedString([]byte(t.signingKey))
}

func (t *Tokens) Parse(userToken string) (string, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(userToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSingningMethod
		}
		return []byte(t.signingKey), nil
	},
	)
	if err != nil {
		return "", err
	}

	return claims.Subject, nil
}

func (t *Tokens) GetRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
