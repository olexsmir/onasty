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
	ErrUserIsAlreeadyVerified  = errors.New("user: user is already verified")

	ErrVerificationTokenNotFound = errors.New("user: verification token not found")
	ErrUserIsNotActivated        = errors.New("user: user is not activated")

	ErrUserNotFound         = errors.New("user: not found")
	ErrUserWrongCredentials = errors.New("user: wrong credentials")
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
		return errors.New("user: invalid email") //nolint:err113
	}

	if len(u.Password) < 6 {
		return errors.New("user: password too short, minimum 6 chars") //nolint:err113
	}

	if len(u.Username) == 0 {
		return errors.New("user: username is required") //nolint:err113
	}

	return nil
}
