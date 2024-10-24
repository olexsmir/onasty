package usercache

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserCacheer interface {
	SetUserIsExists(ctx context.Context, userID string, isExists bool) error
	GetUserIsExists(ctx context.Context, userID string) (isExists bool, err error)

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

func (u *UserCache) SetUserIsExists(ctx context.Context, userID string, val bool) error {
	_, err := u.rdb.
		Set(ctx, getKey("exists", userID), val, time.Hour).
		Result()
	return err
}

func (u *UserCache) GetUserIsExists(ctx context.Context, userID string) (bool, error) {
	res, err := u.rdb.Get(ctx, getKey(userID, "exists")).Bool()
	if err != nil {
		slog.ErrorContext(ctx, "usercache", "err", err)
		return false, err
	}

	return res, nil
}

func (u *UserCache) SetUserIsActivated(ctx context.Context, userID string, val bool) error {
	_, err := u.rdb.
		Set(ctx, getKey("activated", userID), val, time.Hour).
		Result()
	return err
}

func (u *UserCache) GetUserIsActivated(ctx context.Context, userID string) (bool, error) {
	res, err := u.rdb.Get(ctx, getKey(userID, "activated")).Bool()
	if err != nil {
		slog.ErrorContext(ctx, "usercache", "err", err)
		return false, err
	}
	return res, nil
}

// getKey return a key for redis in this format user:<userID>:<key>
func getKey(userID, key string) string {
	var sb strings.Builder
	sb.WriteString("user:")
	sb.WriteString(userID)
	sb.WriteString(":")
	sb.WriteString(key)
	return sb.String()
}
