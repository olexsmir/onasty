package rdb

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type DB struct{ *redis.Client }

func Connect(ctx context.Context, addr, password string, db int) (*DB, error) {
	client := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &DB{Client: client}, nil
}
