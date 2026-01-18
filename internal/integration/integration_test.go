package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/analytics"
	"github.com/Unhyphenated/shrinks-backend/internal/auth"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/service"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"github.com/joho/godotenv"
)

var (
	testStore     *storage.PostgresStore
	testCache     *cache.RedisCache
	testAuth      *auth.AuthService
	testAnalytics *analytics.AnalyticsService
	testLink      *service.LinkService
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if dbURL == "" || redisURL == "" {
		os.Exit(0) // Skip if not configured
	}

	var err error

	testStore, err = storage.NewPostgresStore(dbURL)
	if err != nil {
		panic("Failed to connect to DB: " + err.Error())
	}

	testCache, err = cache.NewRedisCache(redisURL)
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	testAuth = auth.NewAuthService(testStore)
	testAnalytics = analytics.NewAnalyticsService(testStore)
	testLink = service.NewLinkService(testStore, testCache, testAnalytics)

	os.Setenv("JWT_SECRET", "integration-test-secret")

	exitCode := m.Run()

	testStore.Close()
	testCache.Close()
	os.Exit(exitCode)
}

// Helper: cleanup user and related data
func cleanup(t *testing.T, email string) {
	ctx := context.Background()
	_, _ = testStore.Pool.Exec(ctx, `
		DELETE FROM analytics WHERE link_id IN (
			SELECT l.id FROM links l JOIN users u ON l.user_id = u.id WHERE u.email = $1
		)`, email)
	_, _ = testStore.Pool.Exec(ctx, `
		DELETE FROM links WHERE user_id IN (SELECT id FROM users WHERE email = $1)`, email)
	_, _ = testStore.Pool.Exec(ctx, `
		DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email = $1)`, email)
	_, _ = testStore.Pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
}

// Helper: flush redis keys matching pattern
func flushTestKeys(pattern string) {
	ctx := context.Background()
	keys, _ := testCache.Client.Keys(ctx, pattern).Result()
	for _, key := range keys {
		testCache.Client.Del(ctx, key)
	}
}

// Test #87: Full flow - Create link, click it, verify analytics
func TestFullFlow_CreateClickAnalytics(t *testing.T) {
	if testStore == nil {
		t.Skip("DB not configured")
	}

	ctx := context.Background()
	email := "integration-analytics@example.com"
	defer cleanup(t, email)

	// 1. Register user
	registerResp, err := testAuth.Register(ctx, email, "password123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	userID := registerResp.UserID

	// 2. Create a short link
	shortCode, err := testLink.Shorten(ctx, "https://example.com/integration-test", &userID)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if shortCode == "" {
		t.Fatal("Shorten returned empty code")
	}

	// 3. Simulate clicks with different devices
	clicks := []struct {
		ip      string
		device  string
		browser string
		os      string
	}{
		{"1.1.1.0", "Desktop", "Chrome", "Windows"},
		{"1.1.1.0", "Desktop", "Chrome", "Windows"}, // Same IP
		{"2.2.2.0", "Mobile", "Safari", "iOS"},
		{"3.3.3.0", "Desktop", "Firefox", "macOS"},
	}

	for _, click := range clicks {
		event := &model.AnalyticsEvent{
			IPAddress:  click.ip,
			DeviceType: click.device,
			Browser:    click.browser,
			OS:         click.os,
		}

		_, err := testLink.Redirect(ctx, shortCode, event)
		if err != nil {
			t.Fatalf("Redirect failed: %v", err)
		}
	}

	// Give async analytics time to record
	time.Sleep(100 * time.Millisecond)

	// 4. Get link to find ID
	link, err := testStore.GetLinkByCode(ctx, shortCode)
	if err != nil || link == nil {
		t.Fatalf("GetLinkByCode failed: %v", err)
	}

	// 5. Retrieve analytics
	summary, err := testAnalytics.RetrieveAnalytics(ctx, link.ID, "7d")
	if err != nil {
		t.Fatalf("RetrieveAnalytics failed: %v", err)
	}

	// 6. Verify analytics
	if summary.TotalClicks != 4 {
		t.Errorf("TotalClicks = %d, want 4", summary.TotalClicks)
	}

	if summary.UniqueVisitors != 3 {
		t.Errorf("UniqueVisitors = %d, want 3", summary.UniqueVisitors)
	}

	// Check device breakdown
	deviceMap := make(map[string]int)
	for _, d := range summary.ClicksByDevice {
		deviceMap[d.Device] = d.Clicks
	}

	if deviceMap["Desktop"] != 3 {
		t.Errorf("Desktop clicks = %d, want 3", deviceMap["Desktop"])
	}
	if deviceMap["Mobile"] != 1 {
		t.Errorf("Mobile clicks = %d, want 1", deviceMap["Mobile"])
	}
}

// Test #88: Full flow - Register, login, logout, refresh fails
func TestFullFlow_RegisterLoginLogout(t *testing.T) {
	if testStore == nil {
		t.Skip("DB not configured")
	}

	ctx := context.Background()
	email := "integration-auth@example.com"
	password := "password123"
	defer cleanup(t, email)

	// 1. Register
	registerResp, err := testAuth.Register(ctx, email, password)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if registerResp.UserID == 0 {
		t.Fatal("Register returned zero UserID")
	}

	// 2. Login
	loginResp, err := testAuth.Login(ctx, email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if loginResp.AccessToken == "" {
		t.Fatal("Login returned empty AccessToken")
	}
	if loginResp.RefreshToken == "" {
		t.Fatal("Login returned empty RefreshToken")
	}

	// 3. Verify token works
	_, err = testAuth.RefreshAccessToken(ctx, loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshAccessToken failed before logout: %v", err)
	}

	// 4. Logout
	err = testAuth.Logout(ctx, loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// 5. Verify refresh now fails
	_, err = testAuth.RefreshAccessToken(ctx, loginResp.RefreshToken)
	if err == nil {
		t.Error("RefreshAccessToken should fail after logout")
	}

	if err != auth.ErrInvalidRefreshToken {
		t.Errorf("Expected ErrInvalidRefreshToken, got %v", err)
	}
}

// Test #89: Full flow - Create, list, delete links
func TestFullFlow_CreateListDeleteLinks(t *testing.T) {
	if testStore == nil {
		t.Skip("DB not configured")
	}

	ctx := context.Background()
	email := "integration-links@example.com"
	defer cleanup(t, email)

	// 1. Register user
	registerResp, err := testAuth.Register(ctx, email, "password123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	userID := registerResp.UserID

	// 2. Create 3 links
	codes := make([]string, 3)
	for i := 0; i < 3; i++ {
		code, err := testLink.Shorten(ctx, "https://example.com/link"+string(rune('0'+i)), &userID)
		if err != nil {
			t.Fatalf("Shorten failed: %v", err)
		}
		codes[i] = code
	}

	// 3. List links
	links, total, err := testStore.GetUserLinks(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("GetUserLinks failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Total = %d, want 3", total)
	}
	if len(links) != 3 {
		t.Errorf("Got %d links, want 3", len(links))
	}

	// 4. Delete first link
	err = testStore.DeleteLink(ctx, codes[0], userID)
	if err != nil {
		t.Fatalf("DeleteLink failed: %v", err)
	}

	// 5. Verify deleted
	links, total, err = testStore.GetUserLinks(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("GetUserLinks failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Total after delete = %d, want 2", total)
	}

	// 6. Verify redirect fails for deleted link
	_, err = testLink.Redirect(ctx, codes[0], nil)
	if err == nil {
		t.Error("Redirect should fail for deleted link")
	}

	// 7. Other links still work
	url, err := testLink.Redirect(ctx, codes[1], nil)
	if err != nil {
		t.Fatalf("Redirect failed for existing link: %v", err)
	}
	if url == "" {
		t.Error("Redirect returned empty URL")
	}
}
