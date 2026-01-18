package cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/joho/godotenv"
)

var testCache *RedisCache

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../.env")

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Skip tests if REDIS_URL not set
		os.Exit(0)
	}

	var err error
	testCache, err = NewRedisCache(redisURL)
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	exitCode := m.Run()

	testCache.Close()
	os.Exit(exitCode)
}

// Helper to clean up test keys
func cleanup(key string) {
	testCache.Client.Del(context.Background(), key)
}

// Test #22: Cache miss returns nil, not empty struct
func TestCache_GetMiss_ReturnsNil(t *testing.T) {
	if testCache == nil {
		t.Skip("REDIS_URL not set")
	}

	ctx := context.Background()
	key := "test:nonexistent:12345"

	// Ensure key doesn't exist
	cleanup(key)

	link, err := testCache.Get(ctx, key)

	// Should not error
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	// Should return nil for cache miss
	if link != nil {
		t.Errorf("Get returned %+v, want nil for cache miss", link)
	}
}

// Test #23: Set then Get returns same data
func TestCache_SetAndGet_RoundTrip(t *testing.T) {
	if testCache == nil {
		t.Skip("REDIS_URL not set")
	}

	ctx := context.Background()
	key := "test:roundtrip:12345"
	defer cleanup(key)

	original := &model.Link{
		ID:        12345,
		ShortCode: "abc123",
		LongURL:   "https://example.com/very/long/url",
		CreatedAt: time.Now().Truncate(time.Second), // Truncate for comparison
	}

	// Set
	err := testCache.Set(ctx, key, original, 1*time.Minute)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	// Get
	retrieved, err := testCache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Get returned nil, want link")
	}

	// Verify fields
	if retrieved.ID != original.ID {
		t.Errorf("ID = %d, want %d", retrieved.ID, original.ID)
	}
	if retrieved.ShortCode != original.ShortCode {
		t.Errorf("ShortCode = %s, want %s", retrieved.ShortCode, original.ShortCode)
	}
	if retrieved.LongURL != original.LongURL {
		t.Errorf("LongURL = %s, want %s", retrieved.LongURL, original.LongURL)
	}
}

// Test #24: Expiration works
func TestCache_Expiration(t *testing.T) {
	if testCache == nil {
		t.Skip("REDIS_URL not set")
	}

	ctx := context.Background()
	key := "test:expiration:12345"
	defer cleanup(key)

	link := &model.Link{
		ID:        99999,
		ShortCode: "expire",
		LongURL:   "https://example.com/expires",
	}

	// Set with 1 second TTL
	err := testCache.Set(ctx, key, link, 1*time.Second)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	// Should exist immediately
	retrieved, err := testCache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Get returned nil immediately after Set")
	}

	// Wait for expiration
	time.Sleep(1500 * time.Millisecond)

	// Should be gone
	retrieved, err = testCache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get returned error after expiration: %v", err)
	}
	if retrieved != nil {
		t.Errorf("Get returned %+v after expiration, want nil", retrieved)
	}
}

// Test with nil UserID (nullable field)
func TestCache_NullableUserID(t *testing.T) {
	if testCache == nil {
		t.Skip("REDIS_URL not set")
	}

	ctx := context.Background()
	key := "test:nullable:12345"
	defer cleanup(key)

	// Link without user (anonymous)
	original := &model.Link{
		ID:        11111,
		UserID:    nil,
		ShortCode: "anon",
		LongURL:   "https://example.com/anonymous",
	}

	err := testCache.Set(ctx, key, original, 1*time.Minute)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	retrieved, err := testCache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if retrieved.UserID != nil {
		t.Errorf("UserID = %v, want nil", retrieved.UserID)
	}
}

// Test with non-nil UserID
func TestCache_WithUserID(t *testing.T) {
	if testCache == nil {
		t.Skip("REDIS_URL not set")
	}

	ctx := context.Background()
	key := "test:withuser:12345"
	defer cleanup(key)

	userID := uint64(42)
	original := &model.Link{
		ID:        22222,
		UserID:    &userID,
		ShortCode: "owned",
		LongURL:   "https://example.com/owned",
	}

	err := testCache.Set(ctx, key, original, 1*time.Minute)
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	retrieved, err := testCache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if retrieved.UserID == nil {
		t.Fatal("UserID = nil, want non-nil")
	}
	if *retrieved.UserID != userID {
		t.Errorf("UserID = %d, want %d", *retrieved.UserID, userID)
	}
}