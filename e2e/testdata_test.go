package e2e

import (
	"time"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/core/domain"
)

var (
	note = domain.Note{
		ID:                   uuid.New(),
		Content:              "first, testing content",
		Slug:                 "first-testing-content",
		BurnBeforeExpiration: false,
		CreatedAt:            time.Now(),
	}

	noteWithExpiration = domain.Note{
		ID:                   uuid.New(),
		Content:              "testing",
		Slug:                 uuid.New().String(),
		BurnBeforeExpiration: false,
		CreatedAt:            time.Now(),
		ExpiresAt:            time.Now().Add(5 * time.Minute),
	}

	noteExpired = domain.Note{
		ID:                   uuid.New(),
		Content:              "testing",
		Slug:                 uuid.New().String(),
		BurnBeforeExpiration: false,
		CreatedAt:            time.Now().Add(5 * time.Minute),
		ExpiresAt:            time.Now().Add(-(5 * time.Minute)),
	}
)
