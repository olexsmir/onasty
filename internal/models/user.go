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

	ErrResetPasswordTokenAlreadyUsed = errors.New("reset password token is already used")
	ErrVerificationTokenNotFound     = errors.New("user: verification token not found")
	ErrUserIsNotActivated            = errors.New("user: user is not activated")

	ErrUserNotFound         = errors.New("user: not found")
	ErrUserWrongCredentials = errors.New("user: wrong credentials")

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
	if isEmailValid(u.Email) {
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

func isEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
