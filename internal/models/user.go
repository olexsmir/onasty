package models

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	ErrUserEmailIsAlreadyInUse = errors.New("user: email is already in use")
	ErrUsernameIsAlreadyInUse  = errors.New("user: username is already in use")

	ErrUserNotFound = errors.New("user: not found")
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
	return nil
}
