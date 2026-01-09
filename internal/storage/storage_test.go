package storage

import (
	"context"
	"log"
	"os"
	"testing"

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
	shortCode, err := testStore.SaveLink(ctx, longURL, nil) // nil = anonymous link
	if err != nil {
		t.Fatalf("SaveLink failed: %v", err)
	}
	if len(shortCode) == 0 {
		t.Fatal("SaveLink returned empty short code")
	}

	// 2. GetLinkByCode (Integration Test Part 2: Read)
	link, err := testStore.GetLinkByCode(ctx, shortCode)
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

	_, err = testStore.Pool.Exec(ctx, "DELETE FROM links WHERE short_code = $1", shortCode)
	if err != nil {
		t.Logf("Cleanup failed: %v", err) // Log but don't fail the test
	}
}
