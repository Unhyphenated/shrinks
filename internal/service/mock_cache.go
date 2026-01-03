package service

import (
	"context"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/cache"
)

type MockCache struct {
	GetFn   func(ctx context.Context, key string) (string, error)
	SetFn   func(ctx context.Context, key string, val string, expiration time.Duration) error
	CloseFn func()
}

// Ensure MockCache implements cache.Cache interface
var _ cache.Cache = (*MockCache)(nil)

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, key)
	}
	return "", nil // Default: cache miss
}

func (m *MockCache) Set(ctx context.Context, key string, val string, expiration time.Duration) error {
	if m.SetFn != nil {
		return m.SetFn(ctx, key, val, expiration)
	}
	return nil // Default: no-op
}

func (m *MockCache) Close() {
	if m.CloseFn != nil {
		m.CloseFn()
	}
}
