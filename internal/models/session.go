package models

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

var ErrSessionNotFound = errors.New("user: session not found")

type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RefreshToken string
	ExpiresAt    time.Time
}
