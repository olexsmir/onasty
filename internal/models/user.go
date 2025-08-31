package models

import (
	"errors"
	"net/mail"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrUserEmailIsAlreadyInUse = errors.New("user: email is already in use")
	ErrUserIsAlreadyVerified   = errors.New("user: user is already verified")
	ErrUserIsNotActivated      = errors.New("user: user is not activated")
	ErrUserNotFound            = errors.New("user: not found")

	ErrResetPasswordTokenAlreadyUsed = errors.New("reset password token is already used")
	ErrVerificationTokenNotFound     = errors.New("user: verification token not found")

	ErrUserInvalidEmail    = errors.New("user: invalid email")
	ErrUserInvalidPassword = errors.New("user: password too short, minimum 6 chars")
)

type User struct {
	ID          uuid.UUID
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

	return u.ValidatePassword()
}

func (u User) ValidatePassword() error {
	if len(u.Password) < 6 {
		return ErrUserInvalidPassword
	}
	return nil
}

func (u User) IsActivated() bool {
	return u.Activated
}
