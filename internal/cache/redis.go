package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(redisURL string) (*RedisClient, error) {
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

	return &RedisClient{Client: client}, nil
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found in Redis")
	} else if err != nil {
		return "", fmt.Errorf("failed to get value from Redis: %w", err)
	}

	return val, nil
}

func (c *RedisClient) Set(ctx context.Context, key string, val string, expiration time.Duration) error {
	err := c.Client.Set(ctx, key, val, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}
	return nil
}

func (c *RedisClient) Close() {
	c.Client.Close()
}