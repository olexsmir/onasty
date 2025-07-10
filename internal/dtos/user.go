package dtos

import (
	"time"
)

type SignUp struct {
	Email       string
	Password    string
	CreatedAt   time.Time
	LastLoginAt time.Time
}

type SignIn struct {
	Email    string
	Password string
}

type ResendVerificationEmail struct {
	Email string
}

type ChangeUserPassword struct {
	CurrentPassword string
	NewPassword     string
}

type RequestResetPassword struct {
	Email string
}

type ResetPassword struct {
	Token       string
	NewPassword string
}

type OAuthRedirect struct {
	URL   string
	State string
}

type Tokens struct {
	Access  string
	Refresh string
}

type UserInfo struct {
	Email        string
	CreatedAt    time.Time
	LastLoginAt  time.Time
	NotesCreated int
}
