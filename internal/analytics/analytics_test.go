package analytics

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"github.com/joho/godotenv"
)

var testStore *storage.PostgresStore
var testService *AnalyticsService

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		os.Exit(0)
	}

	var err error
	testStore, err = storage.NewPostgresStore(dbURL)
	if err != nil {
		panic("Failed to connect to DB: " + err.Error())
	}

	testService = NewAnalyticsService(testStore)

	exitCode := m.Run()

	testStore.Close()
	os.Exit(exitCode)
}

// Helper: create test user
func createTestUser(t *testing.T, email string) uint64 {
	ctx := context.Background()
	_, _ = testStore.Pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
	userID, err := testStore.CreateUser(ctx, email, "hash123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	return userID
}

// Helper: create test link
func createTestLink(t *testing.T, userID *uint64) *model.Link {
	ctx := context.Background()
	shortCode, err := testStore.SaveLink(ctx, "https://example.com/test", userID)
	if err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}
	link, _ := testStore.GetLinkByCode(ctx, shortCode)
	return link
}

// Helper: cleanup
func cleanup(email string) {
	ctx := context.Background()
	_, _ = testStore.Pool.Exec(ctx, `
		DELETE FROM analytics WHERE link_id IN (
			SELECT l.id FROM links l JOIN users u ON l.user_id = u.id WHERE u.email = $1
		)`, email)
	_, _ = testStore.Pool.Exec(ctx, `
		DELETE FROM links WHERE user_id IN (SELECT id FROM users WHERE email = $1)`, email)
	_, _ = testStore.Pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email = $1)", email)
	_, _ = testStore.Pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
}

// Test #57: RecordEvent saves to DB
func TestRecordEvent_Success(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	email := "record-event-test@example.com"
	defer func() {
		cleanup(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	event := &model.AnalyticsEvent{
		LinkID:     link.ID,
		IPAddress:  "192.168.1.0",
		UserAgent:  "TestAgent",
		DeviceType: "Desktop",
		OS:         "Linux",
		Browser:    "Firefox",
		ClickedAt:  time.Now(),
	}

	err := testService.RecordEvent(ctx, event)
	if err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	// Verify saved
	events, _ := testStore.GetAnalyticsEvents(ctx, link.ID, time.Now().Add(-time.Hour))
	if len(events) != 1 {
		t.Errorf("Got %d events, want 1", len(events))
	}
}

// Test #58: RetrieveAnalytics aggregates total clicks
func TestRetrieveAnalytics_AggregatesCorrectly(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	email := "aggregate-test@example.com"
	defer func() {
		cleanup(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	now := time.Now()

	// Create 5 events
	for i := 0; i < 5; i++ {
		_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{
			LinkID:    link.ID,
			IPAddress: "1.1.1.0",
			ClickedAt: now,
		})
	}

	summary, err := testService.RetrieveAnalytics(ctx, link.ID, "7d")
	if err != nil {
		t.Fatalf("RetrieveAnalytics failed: %v", err)
	}

	if summary.TotalClicks != 5 {
		t.Errorf("TotalClicks = %d, want 5", summary.TotalClicks)
	}
}

// Test #59: RetrieveAnalytics groups by device
func TestRetrieveAnalytics_GroupsByDevice(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	email := "device-group-test@example.com"
	defer func() {
		cleanup(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	now := time.Now()

	// 3 Desktop, 2 Mobile
	devices := []string{"Desktop", "Desktop", "Desktop", "Mobile", "Mobile"}
	for i, device := range devices {
		_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{
			LinkID:     link.ID,
			IPAddress:  "1.1.1." + string(rune('0'+i)),
			DeviceType: device,
			ClickedAt:  now,
		})
	}

	summary, err := testService.RetrieveAnalytics(ctx, link.ID, "7d")
	if err != nil {
		t.Fatalf("RetrieveAnalytics failed: %v", err)
	}

	// Check device breakdown
	deviceMap := make(map[string]int)
	for _, d := range summary.ClicksByDevice {
		deviceMap[d.Device] = d.Clicks
	}

	if deviceMap["Desktop"] != 3 {
		t.Errorf("Desktop clicks = %d, want 3", deviceMap["Desktop"])
	}
	if deviceMap["Mobile"] != 2 {
		t.Errorf("Mobile clicks = %d, want 2", deviceMap["Mobile"])
	}
}

// Test #60: RetrieveAnalytics groups by date
func TestRetrieveAnalytics_GroupsByDate(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	email := "date-group-test@example.com"
	defer func() {
		cleanup(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// 2 today, 3 yesterday
	for i := 0; i < 2; i++ {
		_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{
			LinkID:    link.ID,
			IPAddress: "1.1.1.0",
			ClickedAt: now,
		})
	}
	for i := 0; i < 3; i++ {
		_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{
			LinkID:    link.ID,
			IPAddress: "2.2.2.0",
			ClickedAt: yesterday,
		})
	}

	summary, err := testService.RetrieveAnalytics(ctx, link.ID, "7d")
	if err != nil {
		t.Fatalf("RetrieveAnalytics failed: %v", err)
	}

	if len(summary.ClicksByDate) < 2 {
		t.Errorf("Got %d date buckets, want at least 2", len(summary.ClicksByDate))
	}
}

// Test #61: RetrieveAnalytics groups by browser
func TestRetrieveAnalytics_GroupsByBrowser(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	email := "browser-group-test@example.com"
	defer func() {
		cleanup(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	now := time.Now()

	browsers := []string{"Chrome", "Chrome", "Firefox", "Safari"}
	for i, browser := range browsers {
		_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{
			LinkID:    link.ID,
			IPAddress: "1.1.1." + string(rune('0'+i)),
			Browser:   browser,
			ClickedAt: now,
		})
	}

	summary, err := testService.RetrieveAnalytics(ctx, link.ID, "7d")
	if err != nil {
		t.Fatalf("RetrieveAnalytics failed: %v", err)
	}

	browserMap := make(map[string]int)
	for _, b := range summary.ClicksByBrowser {
		browserMap[b.Browser] = b.Clicks
	}

	if browserMap["Chrome"] != 2 {
		t.Errorf("Chrome clicks = %d, want 2", browserMap["Chrome"])
	}
}

// Test #62: RetrieveAnalytics groups by OS
func TestRetrieveAnalytics_GroupsByOS(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	email := "os-group-test@example.com"
	defer func() {
		cleanup(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	now := time.Now()

	osList := []string{"Windows", "Windows", "macOS", "Linux"}
	for i, os := range osList {
		_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{
			LinkID:    link.ID,
			IPAddress: "1.1.1." + string(rune('0'+i)),
			OS:        os,
			ClickedAt: now,
		})
	}

	summary, err := testService.RetrieveAnalytics(ctx, link.ID, "7d")
	if err != nil {
		t.Fatalf("RetrieveAnalytics failed: %v", err)
	}

	osMap := make(map[string]int)
	for _, o := range summary.ClicksByOS {
		osMap[o.OS] = o.Clicks
	}

	if osMap["Windows"] != 2 {
		t.Errorf("Windows clicks = %d, want 2", osMap["Windows"])
	}
}

// Test #63: RetrieveAnalytics counts unique visitors by IP
func TestRetrieveAnalytics_UniqueVisitors(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	email := "unique-visitors-test@example.com"
	defer func() {
		cleanup(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	now := time.Now()

	// 5 clicks from 3 unique IPs
	ips := []string{"1.1.1.0", "1.1.1.0", "2.2.2.0", "2.2.2.0", "3.3.3.0"}
	for _, ip := range ips {
		_ = testStore.SaveAnalyticsEvent(ctx, &model.AnalyticsEvent{
			LinkID:    link.ID,
			IPAddress: ip,
			ClickedAt: now,
		})
	}

	summary, err := testService.RetrieveAnalytics(ctx, link.ID, "7d")
	if err != nil {
		t.Fatalf("RetrieveAnalytics failed: %v", err)
	}

	if summary.TotalClicks != 5 {
		t.Errorf("TotalClicks = %d, want 5", summary.TotalClicks)
	}

	if summary.UniqueVisitors != 3 {
		t.Errorf("UniqueVisitors = %d, want 3", summary.UniqueVisitors)
	}
}

// Test #64: RetrieveAnalytics returns zeros for no events
func TestRetrieveAnalytics_EmptyForNoEvents(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	email := "empty-analytics-test@example.com"
	defer func() {
		cleanup(email)
	}()

	userID := createTestUser(t, email)
	link := createTestLink(t, &userID)

	// No events

	summary, err := testService.RetrieveAnalytics(ctx, link.ID, "7d")
	if err != nil {
		t.Fatalf("RetrieveAnalytics failed: %v", err)
	}

	if summary.TotalClicks != 0 {
		t.Errorf("TotalClicks = %d, want 0", summary.TotalClicks)
	}

	if summary.UniqueVisitors != 0 {
		t.Errorf("UniqueVisitors = %d, want 0", summary.UniqueVisitors)
	}
}
