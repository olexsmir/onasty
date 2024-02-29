package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/olexsmir/onasty/internal/core/domain"
)

type UserServicer interface {
	SignUp(context.Context, domain.User) error
	SignIn(context.Context, domain.User) (domain.UserTokens, error)
	RefreshTokens(context.Context, string) (domain.UserTokens, error)
	Logout(context.Context, uuid.UUID) error
}

type UserStorer interface {
	Create(context.Context, domain.User) error
	GetUserByCredentials(context.Context, domain.UserCredentials) (domain.User, error)
	SetSession(context.Context, uuid.UUID, string, time.Time) error
	RemoveSession(context.Context, uuid.UUID) error
}
