//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/analytics"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

// ===== MOCKS =====

func newMockStore() *storage.MockStore {
	return &storage.MockStore{
		SaveLinkFn:      func(ctx context.Context, longURL string, userID *uint64) (string, error) { return "", nil },
		GetLinkByCodeFn: func(ctx context.Context, code string) (*model.Link, error) { return nil, nil },
		CloseFn:         func() {},
	}
}

func newMockCache() *cache.MockCache {
	return &cache.MockCache{
		GetFn:   func(ctx context.Context, key string) (*model.Link, error) { return nil, nil },
		SetFn:   func(ctx context.Context, key string, link *model.Link, exp time.Duration) error { return nil },
		CloseFn: func() {},
	}
}

func newMockAnalytics() *analytics.MockAnalytics {
	return &analytics.MockAnalytics{
		RecordEventFn: func(ctx context.Context, event *model.AnalyticsEvent) error { return nil },
	}
}

// ===== TESTS =====

// Test #65: Shorten returns short code
func TestShorten_Success(t *testing.T) {
	mockStore := newMockStore()
	mockStore.SaveLinkFn = func(ctx context.Context, longURL string, userID *uint64) (string, error) {
		return "abc123", nil
	}

	svc := NewLinkService(mockStore, newMockCache(), newMockAnalytics())

	code, err := svc.Shorten(context.Background(), "https://example.com", nil)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if code != "abc123" {
		t.Errorf("Code = %s, want abc123", code)
	}
}

// Test #66: Shorten with userID associates link to user
func TestShorten_WithUserID(t *testing.T) {
	var capturedUserID *uint64

	mockStore := newMockStore()
	mockStore.SaveLinkFn = func(ctx context.Context, longURL string, userID *uint64) (string, error) {
		capturedUserID = userID
		return "xyz789", nil
	}

	svc := NewLinkService(mockStore, newMockCache(), newMockAnalytics())

	userID := uint64(42)
	_, err := svc.Shorten(context.Background(), "https://example.com", &userID)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if capturedUserID == nil {
		t.Fatal("UserID was not passed to store")
	}

	if *capturedUserID != 42 {
		t.Errorf("UserID = %d, want 42", *capturedUserID)
	}
}

// Test #67: Redirect uses cache when available
func TestRedirect_CacheHit(t *testing.T) {
	dbCalled := false

	mockStore := newMockStore()
	mockStore.GetLinkByCodeFn = func(ctx context.Context, code string) (*model.Link, error) {
		dbCalled = true
		return nil, nil
	}

	mockCache := newMockCache()
	mockCache.GetFn = func(ctx context.Context, key string) (*model.Link, error) {
		return &model.Link{
			ID:        1,
			ShortCode: "cached",
			LongURL:   "https://cached.example.com",
		}, nil
	}

	svc := NewLinkService(mockStore, mockCache, newMockAnalytics())

	url, err := svc.Redirect(context.Background(), "cached", nil)
	if err != nil {
		t.Fatalf("Redirect failed: %v", err)
	}

	if url != "https://cached.example.com" {
		t.Errorf("URL = %s, want https://cached.example.com", url)
	}

	// Give goroutine time to potentially call DB
	time.Sleep(10 * time.Millisecond)

	if dbCalled {
		t.Error("DB was called despite cache hit")
	}
}

// Test #68: Redirect falls back to DB on cache miss
func TestRedirect_CacheMiss_DBHit(t *testing.T) {
	cacheSetCalled := false

	mockStore := newMockStore()
	mockStore.GetLinkByCodeFn = func(ctx context.Context, code string) (*model.Link, error) {
		return &model.Link{
			ID:        2,
			ShortCode: "fromdb",
			LongURL:   "https://fromdb.example.com",
		}, nil
	}

	mockCache := newMockCache()
	mockCache.GetFn = func(ctx context.Context, key string) (*model.Link, error) {
		return nil, nil // Cache miss
	}
	mockCache.SetFn = func(ctx context.Context, key string, link *model.Link, exp time.Duration) error {
		cacheSetCalled = true
		return nil
	}

	svc := NewLinkService(mockStore, mockCache, newMockAnalytics())

	url, err := svc.Redirect(context.Background(), "fromdb", nil)
	if err != nil {
		t.Fatalf("Redirect failed: %v", err)
	}

	if url != "https://fromdb.example.com" {
		t.Errorf("URL = %s, want https://fromdb.example.com", url)
	}

	// Give goroutine time to call cache.Set
	time.Sleep(50 * time.Millisecond)

	if !cacheSetCalled {
		t.Error("Cache.Set was not called after DB hit")
	}
}

// Test #69: Redirect returns error for non-existent link
func TestRedirect_NotFound(t *testing.T) {
	mockStore := newMockStore()
	mockStore.GetLinkByCodeFn = func(ctx context.Context, code string) (*model.Link, error) {
		return nil, nil // Not found
	}

	mockCache := newMockCache()
	mockCache.GetFn = func(ctx context.Context, key string) (*model.Link, error) {
		return nil, nil // Cache miss
	}

	svc := NewLinkService(mockStore, mockCache, newMockAnalytics())

	_, err := svc.Redirect(context.Background(), "nonexistent", nil)

	if err == nil {
		t.Error("Redirect should return error for non-existent link")
	}

	if !errors.Is(err, ErrLinkNotFound) {
		t.Errorf("Error = %v, want ErrLinkNotFound", err)
	}
}

// Test #70: Redirect records analytics event
func TestRedirect_RecordsAnalytics(t *testing.T) {
	var capturedEvent *model.AnalyticsEvent

	mockStore := newMockStore()
	mockStore.GetLinkByCodeFn = func(ctx context.Context, code string) (*model.Link, error) {
		return &model.Link{
			ID:        99,
			ShortCode: "tracked",
			LongURL:   "https://tracked.example.com",
		}, nil
	}

	mockCache := newMockCache()

	mockAnalytics := newMockAnalytics()
	mockAnalytics.RecordEventFn = func(ctx context.Context, event *model.AnalyticsEvent) error {
		capturedEvent = event
		return nil
	}

	svc := NewLinkService(mockStore, mockCache, mockAnalytics)

	event := &model.AnalyticsEvent{
		IPAddress:  "1.2.3.0",
		DeviceType: "Desktop",
		Browser:    "Chrome",
		OS:         "Windows",
	}

	_, err := svc.Redirect(context.Background(), "tracked", event)
	if err != nil {
		t.Fatalf("Redirect failed: %v", err)
	}

	// Give goroutine time to record
	time.Sleep(50 * time.Millisecond)

	if capturedEvent == nil {
		t.Fatal("Analytics event was not recorded")
	}

	if capturedEvent.LinkID != 99 {
		t.Errorf("LinkID = %d, want 99", capturedEvent.LinkID)
	}

	if capturedEvent.IPAddress != "1.2.3.0" {
		t.Errorf("IPAddress = %s, want 1.2.3.0", capturedEvent.IPAddress)
	}
}
