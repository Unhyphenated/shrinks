package storage

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/joho/godotenv"
)

var testStore *PostgresStore

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Printf("Warning: .env file not found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set for testing")
	}

	testStore, err = NewPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize test store: %v", err)
	}

	exitCode := m.Run()

	testStore.Close()
	os.Exit(exitCode)
}

// ===== EXISTING TEST (keep) =====

// Test #25: SaveLink and GetLinkByCode
func TestStorage_SaveAndGetLink(t *testing.T) {
	ctx := context.Background()
	longURL := "https://example.com/test-url-12345"

	shortCode, err := testStore.SaveLink(ctx, longURL, nil)
	if err != nil {
		t.Fatalf("SaveLink failed: %v", err)
	}
	if len(shortCode) == 0 {
		t.Fatal("SaveLink returned empty short code")
	}

	link, err := testStore.GetLinkByCode(ctx, shortCode)
	if err != nil {
		t.Fatalf("GetLinkByCode failed: %v", err)
	}
	if link == nil {
		t.Fatal("GetLinkByCode returned nil link")
	}
	if link.LongURL != longURL {
		t.Errorf("LongURL = %s, want %s", link.LongURL, longURL)
	}

	// Cleanup
	_, _ = testStore.Pool.Exec(ctx, "DELETE FROM links WHERE short_code = $1", shortCode)
}

// ===== NEW STORAGE TESTS =====

// Helper: create a test user and return ID
func createTestUser(t *testing.T, email string) uint64 {
	ctx := context.Background()
	// Clean up first
	_, _ = testStore.Pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)

	userID, err := testStore.CreateUser(ctx, email, "hashedpassword123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// Helper: create a test link and return it
func createTestLink(t *testing.T, userID *uint64) *model.Link {
	ctx := context.Background()
	shortCode, err := testStore.SaveLink(ctx, "https://example.com/test", userID)
	if err != nil {
		t.Fatalf("Failed to create test link: %v", err)
	}
	link, _ := testStore.GetLinkByCode(ctx, shortCode)
	return link
}

// Helper: cleanup user and their data
func cleanupUser(email string) {
	ctx := context.Background()
	_, _ = testStore.Pool.Exec(ctx, `
		DELETE FROM analytics WHERE link_id IN (
			SELECT l.id FROM links l
			JOIN users u ON l.user_id = u.id
			WHERE u.email = $1
		)`, email)
	_, _ = testStore.Pool.Exec(ctx, `
		DELETE FROM links WHERE user_id IN (
			SELECT id FROM users WHERE email = $1
		)`, email)
	_, _ = testStore.Pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email = $1)", email)
	_, _ = testStore.Pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
}

// Test #26: GetAnalyticsEvents returns events in date range
func TestGetAnalyticsEvents_ReturnsEventsInDateRange(t *testing.T) {
	ctx := context.Background()
	email := "analytics-range-test@example.com"
	defer func() {
		cleanupUser(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	// Create events at different times
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	events := []*model.AnalyticsEvent{
		{LinkID: link.ID, IPAddress: "1.1.1.0", ClickedAt: now},
		{LinkID: link.ID, IPAddress: "2.2.2.0", ClickedAt: yesterday},
		{LinkID: link.ID, IPAddress: "3.3.3.0", ClickedAt: lastWeek},
	}

	for _, e := range events {
		if err := testStore.SaveAnalyticsEvent(ctx, e); err != nil {
			t.Fatalf("Failed to save event: %v", err)
		}
	}

	// Query last 3 days (should get 2 events)
	startDate := now.Add(-3 * 24 * time.Hour)

	results, err := testStore.GetAnalyticsEvents(ctx, link.ID, startDate)
	if err != nil {
		t.Fatalf("GetAnalyticsEvents failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Got %d events, want 2", len(results))
	}
}

// Test #27: GetAnalyticsEvents returns empty for no clicks
func TestGetAnalyticsEvents_EmptyForNoClicks(t *testing.T) {
	ctx := context.Background()
	email := "analytics-empty-test@example.com"
	defer func() {
		cleanupUser(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	// No events created

	results, err := testStore.GetAnalyticsEvents(ctx, link.ID, time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("GetAnalyticsEvents failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Got %d events, want 0", len(results))
	}
}

// Test #28: GetAnalyticsEvents filters other links
func TestGetAnalyticsEvents_FiltersOtherLinks(t *testing.T) {
	ctx := context.Background()
	email := "analytics-filter-test@example.com"
	defer func() {
		cleanupUser(email)
	}()

	userID := createTestUser(t, email)
	link1 := createTestLink(t, &userID)
	link2 := createTestLink(t, &userID)

	now := time.Now()

	// Events for link1
	_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{LinkID: link1.ID, IPAddress: "1.1.1.0", ClickedAt: now})
	_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{LinkID: link1.ID, IPAddress: "2.2.2.0", ClickedAt: now})

	// Events for link2
	_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{LinkID: link2.ID, IPAddress: "3.3.3.0", ClickedAt: now})

	// Query link1 only
	results, err := testStore.GetAnalyticsEvents(ctx, link1.ID, now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("GetAnalyticsEvents failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Got %d events for link1, want 2", len(results))
	}

	for _, e := range results {
		if e.LinkID != link1.ID {
			t.Errorf("Event has LinkID %d, want %d", e.LinkID, link1.ID)
		}
	}
}

// Test #29: DeleteLink succeeds for owner
func TestDeleteLink_Success(t *testing.T) {
	ctx := context.Background()
	email := "delete-success-test@example.com"
	defer func() {
		cleanupUser(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	err := testStore.DeleteLink(ctx, link.ShortCode, userID)
	if err != nil {
		t.Fatalf("DeleteLink failed: %v", err)
	}

	// Verify deleted
	deleted, _ := testStore.GetLinkByCode(ctx, link.ShortCode)
	if deleted != nil {
		t.Error("Link still exists after deletion")
	}
}

// Test #30: DeleteLink fails for non-owner
func TestDeleteLink_NotOwner(t *testing.T) {
	ctx := context.Background()
	email1 := "delete-owner-test@example.com"
	email2 := "delete-other-test@example.com"
	defer func() {
		cleanupUser(email1)
	}()
	defer func() {
		cleanupUser(email2)
	}()

	userID1 := createTestUser(t, email1)
	userID2 := createTestUser(t, email2)

	link := createTestLink(t, &userID1)

	// User2 tries to delete User1's link
	err := testStore.DeleteLink(ctx, link.ShortCode, userID2)
	if err == nil {
		t.Error("DeleteLink should fail for non-owner")
	}

	// Verify still exists
	existing, _ := testStore.GetLinkByCode(ctx, link.ShortCode)
	if existing == nil {
		t.Error("Link was deleted by non-owner")
	}
}

// Test #31: DeleteLink returns error for non-existent link
func TestDeleteLink_NotFound(t *testing.T) {
	ctx := context.Background()
	email := "delete-notfound-test@example.com"
	defer func() {
		cleanupUser(email)
	}()

	userID := createTestUser(t, email)

	err := testStore.DeleteLink(ctx, "nonexistent123", userID)
	if err == nil {
		t.Error("DeleteLink should return error for non-existent link")
	}
}

// Test #32: GetUserLinks returns paginated results
func TestGetUserLinks_ReturnsPaginated(t *testing.T) {
	ctx := context.Background()
	email := "paginated-test@example.com"
	defer func() {
		cleanupUser(email)
	}()

	userID := createTestUser(t, email)

	// Create 5 links
	for i := 0; i < 5; i++ {
		_, err := testStore.SaveLink(ctx, "https://example.com/page"+string(rune('0'+i)), &userID)
		if err != nil {
			t.Fatalf("Failed to create link: %v", err)
		}
	}

	// Get first page (limit 2)
	links, total, err := testStore.GetUserLinks(ctx, userID, 2, 0)
	if err != nil {
		t.Fatalf("GetUserLinks failed: %v", err)
	}

	if len(links) != 2 {
		t.Errorf("Got %d links, want 2", len(links))
	}

	if total != 5 {
		t.Errorf("Total = %d, want 5", total)
	}

	// Get second page
	links2, _, err := testStore.GetUserLinks(ctx, userID, 2, 2)
	if err != nil {
		t.Fatalf("GetUserLinks page 2 failed: %v", err)
	}

	if len(links2) != 2 {
		t.Errorf("Page 2 got %d links, want 2", len(links2))
	}

	// Verify different links
	if links[0].ID == links2[0].ID {
		t.Error("Page 1 and Page 2 returned same links")
	}
}

// Test #33: GetUserLinks returns empty for user with no links
func TestGetUserLinks_Empty(t *testing.T) {
	ctx := context.Background()
	email := "empty-links-test@example.com"
	defer func() {
		cleanupUser(email)
	}()

	userID := createTestUser(t, email)

	links, total, err := testStore.GetUserLinks(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("GetUserLinks failed: %v", err)
	}

	if len(links) != 0 {
		t.Errorf("Got %d links, want 0", len(links))
	}

	if total != 0 {
		t.Errorf("Total = %d, want 0", total)
	}
}

// Test #34: SaveAnalyticsEvent success
func TestSaveAnalyticsEvent_Success(t *testing.T) {
	ctx := context.Background()
	email := "save-event-test@example.com"
	defer func() {
		cleanupUser(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	event := &model.AnalyticsEvent{
		LinkID:     link.ID,
		IPAddress:  "192.168.1.0",
		UserAgent:  "Mozilla/5.0 Test",
		DeviceType: "Desktop",
		OS:         "Windows",
		Browser:    "Chrome 120",
		ClickedAt:  time.Now(),
	}

	err := testStore.SaveAnalyticsEvent(ctx, event)
	if err != nil {
		t.Fatalf("SaveAnalyticsEvent failed: %v", err)
	}

	// Verify saved
	events, _ := testStore.GetAnalyticsEvents(ctx, link.ID, time.Now().Add(-time.Hour))
	if len(events) != 1 {
		t.Fatalf("Got %d events, want 1", len(events))
	}

	if events[0].Browser != "Chrome 120" {
		t.Errorf("Browser = %s, want Chrome 120", events[0].Browser)
	}
}
