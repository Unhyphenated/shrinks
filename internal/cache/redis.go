package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (*model.Link, error)
	Set(ctx context.Context, key string, val *model.Link, expiration time.Duration) error
	Close()
}

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}
	client := redis.NewClient(opts)

	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	fmt.Println("Successfully initialized Redis Client!")

	return &RedisCache{Client: client}, nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (*model.Link, error) {
	res := c.Client.HGetAll(ctx, key)
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("failed to get value from Redis: %w", err)
	}

	if len(res.Val()) == 0 {
        return nil, nil
    }

	var link model.Link
    if err := res.Scan(&link); err != nil {
        return nil, fmt.Errorf("failed to scan redis hash: %w", err)
    }

    return &link, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, link *model.Link, expiration time.Duration) error {
	pipe := c.Client.Pipeline()
	pipe.HSet(ctx, key, link)
	pipe.Expire(ctx, key, expiration)

	_, err := pipe.Exec(ctx)
    if err != nil {
        return fmt.Errorf("failed to set cache with expiration: %w", err)
    }
	return nil
}

func (c *RedisCache) Close() {
	c.Client.Close()
}