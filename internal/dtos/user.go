package dtos

import (
	"time"
)

type ChangeUserPasswordDTO struct {
	CurrentPassword string
	NewPassword     string
}

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
