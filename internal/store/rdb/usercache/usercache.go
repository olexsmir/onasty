package usercache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserCacheer interface {
	SetUserExists(ctx context.Context, userID string, isExists bool) error
	GetUserExists(ctx context.Context, userID string) (isExists bool, err error)

	SetUserIsActivated(ctx context.Context, userID string, isActivated bool) error
	GetUserIsActivated(ctx context.Context, userID string) (isActivated bool, err error)
}

var _ UserCacheer = (*UserCache)(nil)

type UserCache struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *UserCache {
	return &UserCache{rdb}
}

func (u *UserCache) SetUserExists(ctx context.Context, userID string, val bool) error {
	_, err := u.rdb.Set(
		ctx,
		fmt.Sprintf("user:%s:exists", userID),
		val,
		time.Hour,
	).Result()
	return err
}

func (u *UserCache) GetUserExists(ctx context.Context, userID string) (bool, error) {
	res, err := u.rdb.Get(ctx, fmt.Sprintf("user:%s:exists", userID)).Bool()
	if err != nil {
		slog.ErrorContext(ctx, "usercache", "err", err)
		return false, err
	}

	return res, nil
}

func (u *UserCache) SetUserIsActivated(ctx context.Context, userID string, val bool) error {
	_, err := u.rdb.Set(
		ctx,
		fmt.Sprintf("user:%s:activated", userID),
		val,
		time.Hour,
	).Result()
	return err
}

func (u *UserCache) GetUserIsActivated(ctx context.Context, userID string) (bool, error) {
	res, err := u.rdb.Get(ctx, fmt.Sprintf("user:%s:activated", userID)).Bool()
	if err != nil {
		slog.ErrorContext(ctx, "usercache", "err", err)
		return false, err
	}
	return res, nil
}
