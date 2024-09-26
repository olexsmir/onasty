package dtos

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type UserDTO struct {
	ID          uuid.UUID
	Username    string
	Email       string
	Password    string
	Activated   bool
	CreatedAt   time.Time
	LastLoginAt time.Time
}

type ResetUserPasswordDTO struct {
	// NOTE: probably userID shouldn't be here
	UserID          uuid.UUID
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
