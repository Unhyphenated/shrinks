package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"github.com/joho/godotenv"
)

var testStore *storage.PostgresStore

// TestMain sets up the test database connection
func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		// Don't fail if .env doesn't exist, might be using environment variables
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Skip tests if DATABASE_URL is not set
		os.Exit(0)
	}

	var err2 error
	testStore, err2 = storage.NewPostgresStore(dbURL)
	if err2 != nil {
		panic("Failed to initialize test store: " + err2.Error())
	}

	exitCode := m.Run()

	testStore.Close()
	os.Exit(exitCode)
}

// cleanupUsers removes test users to avoid conflicts
func cleanupUsers(t *testing.T, emails ...string) {
	ctx := context.Background()
	for _, email := range emails {
		// First delete refresh tokens for the user
		user, err := testStore.GetUserByEmail(ctx, email)
		if err == nil && user != nil {
			_ = testStore.DeleteUserRefreshTokens(ctx, user.ID)
		}
		// Then delete the user
		_, err = testStore.Pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
		if err != nil {
			t.Logf("Cleanup failed for %s: %v", email, err)
		}
	}
}

func TestAuthService_Register_Success(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	service := NewAuthService(testStore)
	email := "register-test@example.com"
	password := "password123"

	// Cleanup before and after
	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	resp, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if resp.UserID == 0 {
		t.Error("Expected non-zero user ID")
	}

	// Verify user was created
	user, err := testStore.GetUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if user == nil {
		t.Fatal("User was not created")
	}
	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	service := NewAuthService(testStore)
	email := "duplicate-test@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Register first time
	_, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Try to register again
	_, err = service.Register(context.Background(), email, password)
	if err == nil {
		t.Fatal("Expected error for duplicate email")
	}
	if err != ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	service := NewAuthService(testStore)

	tests := []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"invalid format", "not-an-email"},
		{"missing @", "testexample.com"},
		{"missing domain", "test@"},
		{"missing local part", "@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Register(context.Background(), tt.email, "password123")
			if err == nil {
				t.Error("Expected error for invalid email")
			}
			if err != ErrInvalidEmail {
				t.Errorf("Expected ErrInvalidEmail, got %v", err)
			}
		})
	}
}

func TestAuthService_Register_PasswordValidation(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	service := NewAuthService(testStore)
	email := "password-test@example.com"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"too_short", "short", ErrPasswordTooShort},
		{"too_long", string(make([]byte, 73)), ErrPasswordTooLong},
		{"valid_length", "password123", nil},
		{"exactly_8_chars", "12345678", nil},
		{"exactly_72_chars", string(make([]byte, 72)), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEmail := tt.name + "@example.com"
			cleanupUsers(t, testEmail)
			defer cleanupUsers(t, testEmail)

			_, err := service.Register(context.Background(), testEmail, tt.password)
			if tt.wantErr != nil {
				if err == nil {
					t.Error("Expected error")
				} else if err != tt.wantErr {
					t.Errorf("Expected %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	service := NewAuthService(testStore)
	email := "login-test@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Register user first
	_, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Login
	resp, err := service.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
	if resp.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, resp.User.Email)
	}

	// Verify access token is valid
	claims, err := ValidateToken(resp.AccessToken)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s in token, got %s", email, claims.Email)
	}
	if claims.UserID != resp.User.ID {
		t.Errorf("Expected UserID %d in token, got %d", resp.User.ID, claims.UserID)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	service := NewAuthService(testStore)
	email := "login-invalid@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Register user
	_, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"wrong password", email, "wrongpassword"},
		{"non-existent email", "nonexistent@example.com", password},
		{"empty email", "", password},
		{"invalid email format", "not-an-email", password},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Login(context.Background(), tt.email, tt.password)
			if err == nil {
				t.Error("Expected error for invalid credentials")
			}
			if err != ErrInvalidCredentials && err != ErrInvalidEmail {
				t.Errorf("Expected ErrInvalidCredentials or ErrInvalidEmail, got %v", err)
			}
		})
	}
}

func TestAuthService_EndToEnd_RegisterLoginValidate(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	service := NewAuthService(testStore)
	email := "e2e-test@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Step 1: Register
	registerResp, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}
	if registerResp.UserID == 0 {
		t.Fatal("Expected non-zero user ID")
	}

	// Step 2: Login
	loginResp, err := service.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loginResp.AccessToken == "" {
		t.Fatal("Expected non-empty access token")
	}
	if loginResp.RefreshToken == "" {
		t.Fatal("Expected non-empty refresh token")
	}

	// Step 3: Validate access token
	claims, err := ValidateToken(loginResp.AccessToken)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	// Step 4: Verify claims match
	if claims.UserID != registerResp.UserID {
		t.Errorf("Expected UserID %d, got %d", registerResp.UserID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	// Step 5: Verify user info matches
	if loginResp.User.ID != registerResp.UserID {
		t.Errorf("Expected UserID %d, got %d", registerResp.UserID, loginResp.User.ID)
	}
	if loginResp.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, loginResp.User.Email)
	}
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	service := NewAuthService(testStore)

	_, err := service.Login(context.Background(), "nonexistent@example.com", "password123")
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_RefreshAccessToken_Success(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	service := NewAuthService(testStore)
	email := "refresh-test@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Register and login to get refresh token
	_, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	loginResp, err := service.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if loginResp.RefreshToken == "" {
		t.Fatal("Expected non-empty refresh token")
	}

	// Refresh access token
	refreshResp, err := service.RefreshAccessToken(context.Background(), loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}

	if refreshResp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	// Verify new access token is valid
	claims, err := ValidateToken(refreshResp.AccessToken)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s in token, got %s", email, claims.Email)
	}
	if claims.UserID != loginResp.User.ID {
		t.Errorf("Expected UserID %d in token, got %d", loginResp.User.ID, claims.UserID)
	}
}

func TestAuthService_RefreshAccessToken_InvalidToken(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	service := NewAuthService(testStore)

	// Try to refresh with invalid token
	_, err := service.RefreshAccessToken(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid refresh token")
	}
	if err != ErrInvalidRefreshToken {
		t.Errorf("Expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthService_RefreshAccessToken_ExpiredToken(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	service := NewAuthService(testStore)
	email := "expired-test@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Register and login
	_, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	loginResp, err := service.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Manually expire the refresh token in database
	ctx := context.Background()
	tokenHash := HashRefreshToken(loginResp.RefreshToken)
	storedToken, err := testStore.GetRefreshToken(ctx, tokenHash)
	if err != nil || storedToken == nil {
		t.Fatalf("Failed to get refresh token: %v", err)
	}

	// Update expiration to past
	_, err = testStore.Pool.Exec(ctx,
		"UPDATE refresh_tokens SET expires_at = $1 WHERE token_hash = $2",
		storedToken.ExpiresAt.Add(-8*24*time.Hour), tokenHash)
	if err != nil {
		t.Fatalf("Failed to expire token: %v", err)
	}

	// Try to refresh with expired token
	_, err = service.RefreshAccessToken(context.Background(), loginResp.RefreshToken)
	if err == nil {
		t.Fatal("Expected error for expired refresh token")
	}
	if err != ErrRefreshTokenExpired {
		t.Errorf("Expected ErrRefreshTokenExpired, got %v", err)
	}
}

func TestAuthService_ConcurrentRegistrations(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	service := NewAuthService(testStore)
	email := "concurrent-test@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Try to register the same email concurrently
	done := make(chan bool, 2)
	errors := make(chan error, 2)

	go func() {
		_, err := service.Register(context.Background(), email, password)
		errors <- err
		done <- true
	}()

	go func() {
		_, err := service.Register(context.Background(), email, password)
		errors <- err
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// At least one should succeed, at least one should fail
	err1 := <-errors
	err2 := <-errors

	successCount := 0
	if err1 == nil {
		successCount++
	}
	if err2 == nil {
		successCount++
	}

	if successCount != 1 {
		t.Errorf("Expected exactly one success, got %d successes", successCount)
	}

	// At least one should be ErrUserAlreadyExists
	if err1 != ErrUserAlreadyExists && err2 != ErrUserAlreadyExists {
		t.Error("Expected at least one ErrUserAlreadyExists error")
	}
}

func TestAuthService_Logout_DeletesRefreshToken(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	service := NewAuthService(testStore)
	email := "logout-test@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Register and login
	_, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	loginResp, err := service.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Logout
	err = service.Logout(context.Background(), loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Verify token is deleted from DB
	tokenHash := HashRefreshToken(loginResp.RefreshToken)
	storedToken, err := testStore.GetRefreshToken(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("GetRefreshToken failed: %v", err)
	}

	if storedToken != nil {
		t.Error("Refresh token still exists after logout")
	}
}

// Test #55: Logout with invalid token returns error
func TestAuthService_Logout_InvalidToken(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	service := NewAuthService(testStore)

	err := service.Logout(context.Background(), "invalid-token-that-doesnt-exist")

	if err == nil {
		t.Error("Logout should return error for invalid token")
	}
}

// Test #56: Cannot refresh after logout
func TestAuthService_Logout_CannotReuseToken(t *testing.T) {
	if testStore == nil {
		t.Skip("DATABASE_URL not set")
	}

	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	service := NewAuthService(testStore)
	email := "logout-reuse-test@example.com"
	password := "password123"

	cleanupUsers(t, email)
	defer cleanupUsers(t, email)

	// Register and login
	_, err := service.Register(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	loginResp, err := service.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Logout
	err = service.Logout(context.Background(), loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Try to refresh with the logged-out token
	_, err = service.RefreshAccessToken(context.Background(), loginResp.RefreshToken)
	if err == nil {
		t.Error("RefreshAccessToken should fail after logout")
	}

	if err != ErrInvalidRefreshToken {
		t.Errorf("Expected ErrInvalidRefreshToken, got %v", err)
	}
}