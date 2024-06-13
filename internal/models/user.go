package models

import "errors"

var (
	ErrUserEmailIsAlreadyInUse = errors.New("user: email is already in use")
	ErrUsernameIsAlreadyInUse  = errors.New("user: username is already in use")
)

type User struct{}
