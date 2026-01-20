package cache

import (
	"context"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
)

type MockCache struct {
	GetFn   func(ctx context.Context, key string) (*model.Link, error)
	SetFn   func(ctx context.Context, key string, link *model.Link, expiration time.Duration) error
	CloseFn func()
}

// Ensure MockCache implements cache.Cache interface
var _ Cache = (*MockCache)(nil)

func (m *MockCache) Get(ctx context.Context, key string) (*model.Link, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, key)
	}
	return nil, nil // Default: cache miss
}

func (m *MockCache) Set(ctx context.Context, key string, link *model.Link, expiration time.Duration) error {
	if m.SetFn != nil {
		return m.SetFn(ctx, key, link, expiration)
	}
	return nil // Default: no-op
}

func (m *MockCache) Close() {
	if m.CloseFn != nil {
		m.CloseFn()
	}
}
