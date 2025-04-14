package models

import (
	"errors"
	"net/mail"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrUserEmailIsAlreadyInUse = errors.New("user: email is already in use")
	ErrUsernameIsAlreadyInUse  = errors.New("user: username is already in use")
	ErrUserIsAlreadyVerified   = errors.New("user: user is already verified")

	ErrVerificationTokenNotFound = errors.New("user: verification token not found")
	ErrUserIsNotActivated        = errors.New("user: user is not activated")

	ErrUserNotFound         = errors.New("user: not found")
	ErrUserWrongCredentials = errors.New("user: wrong credentials")

	ErrUserInvalidEmail    = errors.New("user: invalid email")
	ErrUserInvalidPassword = errors.New("user: password too short, minimum 6 chars")
	ErrUserInvalidUsername = errors.New("user: username is required")
)

type User struct {
	ID          uuid.UUID
	Username    string
	Email       string
	Activated   bool
	Password    string
	CreatedAt   time.Time
	LastLoginAt time.Time
}

func (u User) Validate() error {
	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return ErrUserInvalidEmail
	}

	if len(u.Password) < 6 {
		return ErrUserInvalidPassword
	}

	if len(u.Username) == 0 {
		return ErrUserInvalidUsername
	}

	return nil
}

func (u User) IsActivated() bool {
	return u.Activated
}
