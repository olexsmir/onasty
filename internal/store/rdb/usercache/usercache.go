package usercache

import (
	"context"
	"strings"
	"time"

	"github.com/olexsmir/onasty/internal/store/rdb"
)

type UserCacheer interface {
	SetIsExists(ctx context.Context, userID string, isExists bool) error
	GetIsExists(ctx context.Context, userID string) (isExists bool, err error)

	SetIsActivated(ctx context.Context, userID string, isActivated bool) error
	GetIsActivated(ctx context.Context, userID string) (isActivated bool, err error)
}

var _ UserCacheer = (*UserCache)(nil)

type UserCache struct {
	rdb *rdb.DB
	ttl time.Duration
}

func New(rdb *rdb.DB, ttl time.Duration) *UserCache {
	return &UserCache{
		rdb: rdb,
		ttl: ttl,
	}
}

func (u *UserCache) SetIsExists(ctx context.Context, userID string, val bool) error {
	_, err := u.rdb.
		Set(ctx, getKey("exists", userID), val, u.ttl).
		Result()
	return err
}

func (u *UserCache) GetIsExists(ctx context.Context, userID string) (bool, error) {
	res, err := u.rdb.Get(ctx, getKey(userID, "exists")).Bool()
	if err != nil {
		return false, err
	}

	return res, nil
}

func (u *UserCache) SetIsActivated(ctx context.Context, userID string, val bool) error {
	_, err := u.rdb.
		Set(ctx, getKey("activated", userID), val, u.ttl).
		Result()
	return err
}

func (u *UserCache) GetIsActivated(ctx context.Context, userID string) (bool, error) {
	res, err := u.rdb.Get(ctx, getKey(userID, "activated")).Bool()
	if err != nil {
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
