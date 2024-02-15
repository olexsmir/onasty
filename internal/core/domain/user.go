package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailIsInvalid = errors.New("user: email is invalid")

	ErrUserNotFound            = errors.New("user: not found")
	ErrUsersSessionNotFound    = errors.New("user: session not found")
	ErrUserEmailIsAlreadyInUse = errors.New("user: email is already in use")
)

type User struct {
	ID             uuid.UUID
	Username       string
	Email          string
	Password       string
	CreatedAt      time.Time
	LastLoginnedAt time.Time
}

func (u User) Validate() error {
	// TODO: check if email is valid

	return nil
}

type UserTokens struct {
	Access  string
	Refresh string
}

type UserCredentials struct {
	Email    string
	Password string
}

type UserSession struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RefreshToken string
	ExpiresAt    time.Time
}
