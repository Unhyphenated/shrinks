package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Close()
}

func NewRedisClient(redisURL string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
		Password: "",
		DB: 0,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return rdb, nil
}