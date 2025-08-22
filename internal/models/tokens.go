package models

import (
	"errors"
	"net/mail"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrResetPasswordTokenExpired  = errors.New("reset password token expired")
	ErrResetPasswordTokenNotFound = errors.New("reset password token not found")

	ErrChangeEmailTokenExpired       = errors.New("change email token expired")
	ErrChangeEmailTokenNotFound      = errors.New("change email token not found")
	ErrChangeEmailTokenIsAlreadyUsed = errors.New("change email token is already used")
)

type ResetPasswordToken struct {
	UserID    uuid.UUID
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (p ResetPasswordToken) IsExpired() bool {
	return p.ExpiresAt.Before(time.Now())
}

type VerificationToken struct {
	UserID    uuid.UUID
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type ChangeEmailToken struct {
	UserID    uuid.UUID
	Token     string
	NewEmail  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (c ChangeEmailToken) IsExpired() bool {
	return c.ExpiresAt.Before(time.Now())
}

func (c ChangeEmailToken) Validate() error {
	_, err := mail.ParseAddress(c.NewEmail)
	if err != nil {
		return ErrUserInvalidEmail
	}

	return nil
}
