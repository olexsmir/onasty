package dtos

import (
	"time"
)

type SignUp struct {
	Username    string
	Email       string
	Password    string
	CreatedAt   time.Time
	LastLoginAt time.Time
}

type SignIn struct {
	Email    string
	Password string
}

type ChangeUserPassword struct {
	CurrentPassword string
	NewPassword     string
}

type ResetPassword struct {
	Email string
}

type Tokens struct {
	Access  string
	Refresh string
}
