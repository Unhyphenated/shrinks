package storage

import (
	"context"
	"os"
	"testing"
	"log"

	"github.com/joho/godotenv"
)

var testStore *PostgresStore

// TestMain is a special function that runs once before any test in this package.
func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Setup: Connect to the test database (using the DATABASE_URL from .env)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set for testing. Cannot run integration tests.")
	}
	
	testStore, err = NewPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize test store: %v", err)
	}
	
	exitCode := m.Run() 
	
	// Teardown: Close the connection pool after all tests are done
	testStore.Close()
	os.Exit(exitCode)
}

func TestStorage_SaveAndGetLink(t *testing.T) {
	ctx := context.Background()
	longURL := "https://example.com/test-url-12345"
	
	// 1. SaveLink (Integration Test Part 1: Write)
	shortURL, err := testStore.SaveLink(ctx, longURL)
	if err != nil {
		t.Fatalf("SaveLink failed: %v", err)
	}
	if len(shortURL) == 0 {
		t.Fatal("SaveLink returned empty short code")
	}

	// 2. GetLinkByCode (Integration Test Part 2: Read)
	link, err := testStore.GetLinkByCode(ctx, shortURL)
	if err != nil {
		t.Fatalf("GetLinkByCode failed: %v", err)
	}
	if link == nil {
		t.Fatal("GetLinkByCode returned nil link")
	}
	
	// Verification
	if link.LongURL != longURL {
		t.Errorf("Expected URL %s, Got %s", longURL, link.LongURL)
	}
	
	_, err = testStore.Pool.Exec(ctx, "DELETE FROM links WHERE short_code = $1", shortURL)
	if err != nil {
		t.Logf("Cleanup failed: %v", err) // Log but don't fail the test
	}
}

// internal/storage/storage_test.go

func TestStorage_UpdateClickCount(t *testing.T) {
	ctx := context.Background()
	longURL := "https://www.testing-my-click-tracker.com/page-a"

	var shortURL string

	// 1. Setup: Insert a new link to get an ID (clicks starts at 0)
	var err error
	shortURL, err = testStore.SaveLink(ctx, longURL)
	if err != nil {
		t.Fatalf("Setup failed: SaveLink: %v", err)
	}
	
	// 2. Schedule Cleanup: CRITICAL STEP
	// This function ensures the row is deleted, even if the test fails at step 3.
	t.Cleanup(func() {
		_, err := testStore.Pool.Exec(ctx, "DELETE FROM links WHERE short_code = $1", shortURL)
		if err != nil {
			// Log cleanup failure, but don't fail the test itself
			t.Logf("Cleanup failed for shortURL %s: %v", shortURL, err)
		}
	})

	// 3. Read Initial State: Get the link to retrieve the initial ID and click count (0)
	link, err := testStore.GetLinkByCode(ctx, shortURL)
	if err != nil || link == nil {
		t.Fatalf("Setup failed: GetLinkByCode: %v", err)
	}
    
    if link.Clicks != 0 {
        t.Fatalf("Initial clicks expected 0, got %d", link.Clicks)
    }

	err = testStore.UpdateClickCount(ctx, link.ID)
	if err != nil {
		t.Fatalf("UpdateClickCount failed: %v", err)
	}

	updatedLink, err := testStore.GetLinkByCode(ctx, shortURL)
	if err != nil || updatedLink == nil {
		t.Fatalf("Verification failed: GetLinkByCode after update: %v", err)
	}

	// CRITICAL CHECK: The count must have incremented by exactly 1 (0 -> 1)
	expectedClicks := link.Clicks + 1 
	if updatedLink.Clicks != expectedClicks {
		t.Errorf("Click count mismatch: Expected %d, Got %d", expectedClicks, updatedLink.Clicks)
	}
}