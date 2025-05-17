package models

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrResetPasswordTokenExpired  = errors.New("reset password token expired")
	ErrResetPasswordTokenNotFound = errors.New("reset password token not found")
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
