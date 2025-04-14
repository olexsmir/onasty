package dtos

import (
	"time"
)

type CreateUserDTO struct {
	Username    string
	Email       string
	Password    string
	CreatedAt   time.Time
	LastLoginAt time.Time
}

type SignInDTO struct {
	Email    string
	Password string
}

type ChangeUserPasswordDTO struct {
	CurrentPassword string
	NewPassword     string
}

type TokensDTO struct {
	Access  string
	Refresh string
}
