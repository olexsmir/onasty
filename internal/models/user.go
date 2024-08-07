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

	ErrUserNotFound         = errors.New("user: not found")
	ErrUserWrongCredentials = errors.New("user: wrong credentials")
)

type User struct {
	ID          uuid.UUID
	Username    string
	Email       string
	Password    string
	CreatedAt   time.Time
	LastLoginAt time.Time
}

func (u User) Validate() error {
	// NOTE: there's probably a better way to validate emails
	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return errors.New("user: invalid email")
	}

	if len(u.Password) < 6 {
		return errors.New("user: password too short, minimum 6 chars")
	}

	if len(u.Username) == 0 {
		return errors.New("user: username is required")
	}

	return nil
}
