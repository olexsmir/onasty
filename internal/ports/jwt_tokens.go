package ports

import "time"

type JWTTokeniser interface {
	GetToken(subject string, ttl time.Duration) (string, error)
	GetRefreshToken() (string, error)
	Parse(token string) (string, error)
}
