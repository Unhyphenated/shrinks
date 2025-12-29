package cache

import (
	"context"
	"fmt"
	"time"
	"encoding/json"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Client *redis.Client
}

func NewRedisCache(redisURL string) (*Cache, error) {
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

	return &Cache{Client: client}, nil
}

func (c *Cache) Get(ctx context.Context, key string) (*model.Link, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return &model.Link{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get value from Redis: %w", err)
	}

	return &model.Link{
		LongURL: val,
		ShortURL: key,
		Clicks: 0,
		CreatedAt: time.Now(),
	}, nil
}

func (c *Cache) Set(ctx context.Context, key string, val *model.Link, expiration time.Duration) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("Failed to marshal link to JSON: %w", err)
	}

	err = c.Client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}
	return nil
}

func (c *Cache) Close() {
	c.Client.Close()
}