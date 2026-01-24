//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors" // Needed for simulating internal errors
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/analytics"
	"github.com/Unhyphenated/shrinks-backend/internal/auth"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/encoding"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/service"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

type MockConfig struct {
	SaveLinkFn func(ctx context.Context, longURL string, userID *uint64) (string, error)
}

func newMockStore(cfg MockConfig) *storage.MockStore {
	// Default implementation for methods not being tested
	defaultGetFn := func(ctx context.Context, shortURL string) (*model.Link, error) { return nil, nil }

	return &storage.MockStore{
		SaveLinkFn:      cfg.SaveLinkFn,
		GetLinkByCodeFn: defaultGetFn,
		CloseFn:         func() {},
	}
}

func newMockCache() *cache.MockCache {
	return &cache.MockCache{
		GetFn: func(ctx context.Context, key string) (*model.Link, error) {
			return nil, nil // Default: cache miss (empty string)
		},
		SetFn: func(ctx context.Context, key string, val *model.Link, expiration time.Duration) error {
			return nil // Default: no-op
		},
		CloseFn: func() {},
	}
}

func newMockAnalytics() *analytics.MockAnalytics {
	return &analytics.MockAnalytics{
		RecordEventFn: func(ctx context.Context, event *model.AnalyticsEvent) error {
			return nil
		},
		RetrieveAnalyticsFn: func(ctx context.Context, linkID uint64, period string) (*model.AnalyticsSummary, error) {
			return nil, nil
		},
	}
}

func newMockLinkService() *service.MockLinkService {
	return &service.MockLinkService{
		// No default implementations - let each test configure what it needs
	}
}

// =================================================================
// 1. E2E Test: Success Case (HTTP 201 Created)
// =================================================================

func TestHandlerShorten_Success(t *testing.T) {
	const expectedLongURL = "https://www.google.com/test-success"
	const mockID uint64 = 100
	expectedShortURL := encoding.Encode(mockID)

	cfg := MockConfig{
		SaveLinkFn: func(ctx context.Context, longURL string, userID *uint64) (string, error) {
			if longURL != expectedLongURL {
				t.Fatalf("Mock received wrong URL: got %s", longURL)
			}
			// Simulates the DB return and subsequent encoding logic
			return expectedShortURL, nil
		},
	}
	mockStore := newMockStore(cfg)
	mockCache := newMockCache()
	mockAnalytics := newMockAnalytics()

	svc := service.NewLinkService(mockStore, mockCache, mockAnalytics)
	handler := handlerShorten(svc)

	reqBody := model.CreateLinkRequest{URL: expectedLongURL}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/shorten", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Status mismatch. Got %v, want %v. Body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp model.CreateLinkResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.ShortCode != expectedShortURL {
		t.Errorf("Short code mismatch: Got %s, want %s", resp.ShortCode, expectedShortURL)
	}
}

// =================================================================
// 2. E2E Test: Internal Server Error (HTTP 500)
// =================================================================
// This tests the response when the service/database fails unexpectedly.

func TestHandlerShorten_InternalServerError(t *testing.T) {
	const testURL = "https://fail.com"

	mockDBError := errors.New("simulated DB connection failed")
	cfg := MockConfig{
		SaveLinkFn: func(ctx context.Context, longURL string, userID *uint64) (string, error) {
			return "", mockDBError
		},
	}
	mockStore := newMockStore(cfg)
	mockCache := newMockCache()
	mockAnalytics := newMockAnalytics()

	svc := service.NewLinkService(mockStore, mockCache, mockAnalytics)
	handler := handlerShorten(svc)

	reqBody := model.CreateLinkRequest{URL: testURL}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/shorten", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Status mismatch. Got %v, want %v. Body: %s", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

// =================================================================
// 3. E2E Test: Bad Request / Validation Failure (HTTP 400)
// =================================================================

func TestHandlerShorten_BadRequest(t *testing.T) {
	cfg := MockConfig{
		SaveLinkFn: func(ctx context.Context, longURL string, userID *uint64) (string, error) {
			return "", nil // Won't be called due to validation error
		},
	}
	mockStore := newMockStore(cfg)
	mockCache := newMockCache()
	mockAnalytics := newMockAnalytics()
	svc := service.NewLinkService(mockStore, mockCache, mockAnalytics)
	handler := handlerShorten(svc)

	invalidBody := `{"not_a_url_field": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/shorten", bytes.NewReader([]byte(invalidBody)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Status mismatch. Got %v, want %v. Body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestHandlerRedirect_Success(t *testing.T) {
	mockStore := newMockStore(MockConfig{})
	mockStore.GetLinkByCodeFn = func(ctx context.Context, code string) (*model.Link, error) {
		return &model.Link{
			ID:        1,
			ShortCode: "abc123",
			LongURL:   "https://example.com/destination",
		}, nil
	}

	mockCache := newMockCache()
	mockAnalytics := newMockAnalytics()

	svc := service.NewLinkService(mockStore, mockCache, mockAnalytics)
	handler := handlerRedirect(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/abc123", nil)
	req.SetPathValue("shortCode", "abc123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusFound)
	}

	location := rr.Header().Get("Location")
	if location != "https://example.com/destination" {
		t.Errorf("Location = %s, want https://example.com/destination", location)
	}
}

func TestHandlerRedirect_NotFound(t *testing.T) {
	mockStore := newMockStore(MockConfig{})
	mockStore.GetLinkByCodeFn = func(ctx context.Context, code string) (*model.Link, error) {
		return nil, nil
	}

	mockCache := newMockCache()
	mockAnalytics := newMockAnalytics()

	svc := service.NewLinkService(mockStore, mockCache, mockAnalytics)
	handler := handlerRedirect(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/nonexistent", nil)
	req.SetPathValue("shortCode", "nonexistent")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHandlerHealth_Success(t *testing.T) {
	handler := handlerHealth()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// ===== REGISTER ENDPOINT =====

// Test: Register success
func TestHandlerRegister_Success(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		RegisterFn: func(ctx context.Context, email, password string) (model.RegisterResponse, error) {
			return model.RegisterResponse{UserID: 123}, nil
		},
	}

	handler := handlerRegister(mockAuthService)

	body := `{"email": "test@example.com", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var resp model.RegisterResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if resp.UserID != 123 {
		t.Errorf("UserID = %d, want 123", resp.UserID)
	}
}

// Test: Register with invalid email
func TestHandlerRegister_InvalidEmail(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		RegisterFn: func(ctx context.Context, email, password string) (model.RegisterResponse, error) {
			return model.RegisterResponse{}, auth.ErrInvalidEmail
		},
	}

	handler := handlerRegister(mockAuthService)

	body := `{"email": "not-an-email", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// Test: Register with duplicate email
func TestHandlerRegister_DuplicateEmail(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		RegisterFn: func(ctx context.Context, email, password string) (model.RegisterResponse, error) {
			return model.RegisterResponse{}, auth.ErrUserAlreadyExists
		},
	}

	handler := handlerRegister(mockAuthService)

	body := `{"email": "existing@example.com", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusConflict)
	}
}

// Test: Register with short password
func TestHandlerRegister_ShortPassword(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		RegisterFn: func(ctx context.Context, email, password string) (model.RegisterResponse, error) {
			return model.RegisterResponse{}, auth.ErrPasswordTooShort
		},
	}

	handler := handlerRegister(mockAuthService)

	body := `{"email": "test@example.com", "password": "short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// ===== LOGIN ENDPOINT =====

// Test: Login success
func TestHandlerLogin_Success(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		LoginFn: func(ctx context.Context, email, password string) (model.AuthResponse, error) {
			return model.AuthResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				User:         model.User{ID: 1, Email: email},
			}, nil
		},
	}

	handler := handlerLogin(mockAuthService)

	body := `{"email": "test@example.com", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp model.AuthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if resp.AccessToken != "access-token" {
		t.Errorf("AccessToken = %s, want access-token", resp.AccessToken)
	}
}

// Test: Login with wrong password
func TestHandlerLogin_WrongPassword(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		LoginFn: func(ctx context.Context, email, password string) (model.AuthResponse, error) {
			return model.AuthResponse{}, auth.ErrInvalidCredentials
		},
	}

	handler := handlerLogin(mockAuthService)

	body := `{"email": "test@example.com", "password": "wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// Test: Login with non-existent user
func TestHandlerLogin_NonExistentUser(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		LoginFn: func(ctx context.Context, email, password string) (model.AuthResponse, error) {
			return model.AuthResponse{}, auth.ErrInvalidCredentials
		},
	}

	handler := handlerLogin(mockAuthService)

	body := `{"email": "nobody@example.com", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// ===== REFRESH ENDPOINT =====

// Test: Refresh success
func TestHandlerRefresh_Success(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		RefreshAccessTokenFn: func(ctx context.Context, refreshToken string) (model.RefreshTokenResponse, error) {
			return model.RefreshTokenResponse{AccessToken: "new-access-token"}, nil
		},
	}

	handler := handlerRefresh(mockAuthService)

	body := `{"refresh_token": "valid-refresh-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp model.RefreshTokenResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if resp.AccessToken != "new-access-token" {
		t.Errorf("AccessToken = %s, want new-access-token", resp.AccessToken)
	}
}

// Test: Refresh with invalid token
func TestHandlerRefresh_InvalidToken(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		RefreshAccessTokenFn: func(ctx context.Context, refreshToken string) (model.RefreshTokenResponse, error) {
			return model.RefreshTokenResponse{}, auth.ErrInvalidRefreshToken
		},
	}

	handler := handlerRefresh(mockAuthService)

	body := `{"refresh_token": "invalid-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// Test: Refresh with expired token
func TestHandlerRefresh_ExpiredToken(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		RefreshAccessTokenFn: func(ctx context.Context, refreshToken string) (model.RefreshTokenResponse, error) {
			return model.RefreshTokenResponse{}, auth.ErrRefreshTokenExpired
		},
	}

	handler := handlerRefresh(mockAuthService)

	body := `{"refresh_token": "expired-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// Test #83: Delete link returns 403 for non-owner
func TestHandlerDeleteLink_Forbidden(t *testing.T) {
	mockLinkService := &service.MockLinkService{}

	// 2. THIS IS THE KEY: Define the function on the SERVICE mock
	mockLinkService.DeleteLinkFn = func(ctx context.Context, shortCode string, uid uint64) error {
		// Return the error that triggers a 403 in your handler
		return service.ErrNotOwner
	}

	// 3. Pass this mock to the handler
	handler := handlerDeleteLink(mockLinkService)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/links/abc123", nil)
	req.SetPathValue("shortCode", "abc123")
	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 999})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// Test #84: Delete link returns 404 for non-existent
func TestHandlerDeleteLink_NotFound(t *testing.T) {
	mockLinkService := &service.MockLinkService{
		DeleteLinkFn: func(ctx context.Context, shortCode string, uid uint64) error {
			return service.ErrLinkNotFound
		},
	}

	handler := handlerDeleteLink(mockLinkService)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/links/nonexistent", nil)
	req.SetPathValue("shortCode", "nonexistent")
	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 1})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

// Test #76: Analytics endpoint returns JSON
func TestHandlerAnalytics_Success(t *testing.T) {
	userID := uint64(1)
	mockLinkService := &service.MockLinkService{
		GetLinkByCodeFn: func(ctx context.Context, code string) (*model.Link, error) {
			return &model.Link{
				ID:        100,
				UserID:    &userID,
				ShortCode: "abc123",
				LongURL:   "https://example.com",
			}, nil
		},
	}

	mockAnalytics := &analytics.MockAnalytics{
		RetrieveAnalyticsFn: func(ctx context.Context, linkID uint64, period string) (*model.AnalyticsSummary, error) {
			return &model.AnalyticsSummary{
				LinkID:         linkID,
				TotalClicks:    42,
				UniqueVisitors: 30,
			}, nil
		},
	}

	handler := handlerLinkAnalytics(mockAnalytics, mockLinkService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/abc123/analytics", nil)
	req.SetPathValue("shortCode", "abc123")

	// Add auth context
	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: userID})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp model.AnalyticsSummary
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.TotalClicks != 42 {
		t.Errorf("TotalClicks = %d, want 42", resp.TotalClicks)
	}
}

// Test #77: Analytics returns 401 without auth
func TestHandlerAnalytics_Unauthorized(t *testing.T) {
	mockLinkService := newMockLinkService()
	mockAnalytics := newMockAnalytics()

	handler := handlerLinkAnalytics(mockAnalytics, mockLinkService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/abc123/analytics", nil)
	req.SetPathValue("shortCode", "abc123")
	// No auth context

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// Test #78: Analytics returns 403 for non-owned link
func TestHandlerAnalytics_Forbidden(t *testing.T) {
	ownerID := uint64(1)
	otherUserID := uint64(999)
	mockLinkService := &service.MockLinkService{
		GetLinkByCodeFn: func(ctx context.Context, code string) (*model.Link, error) {
			return &model.Link{
				ID:        100,
				UserID:    &ownerID, // Owned by user 1
				ShortCode: "abc123",
			}, nil
		},
	}

	mockAnalytics := newMockAnalytics()

	handler := handlerLinkAnalytics(mockAnalytics, mockLinkService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/abc123/analytics", nil)
	req.SetPathValue("shortCode", "abc123")

	// Auth as different user
	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: otherUserID})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

// Test #79: Analytics returns 404 for non-existent link
func TestHandlerAnalytics_NotFound(t *testing.T) {
	mockLinkService := newMockLinkService()
	mockStore := newMockStore(MockConfig{})
	mockStore.GetLinkByCodeFn = func(ctx context.Context, code string) (*model.Link, error) {
		return nil, nil
	}

	mockAnalytics := newMockAnalytics()

	handler := handlerLinkAnalytics(mockAnalytics, mockLinkService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/nonexistent/analytics", nil)
	req.SetPathValue("shortCode", "nonexistent")

	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 1})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

// Test #80: List links returns paginated results
func TestHandlerListLinks_Success(t *testing.T) {
	userID := uint64(1)
	mockLinkService := &service.MockLinkService{
		GetUserLinksFn: func(ctx context.Context, uid uint64, limit, offset int) ([]model.Link, int, error) {
			return []model.Link{
				{ID: 1, ShortCode: "link1", LongURL: "https://example.com/1"},
				{ID: 2, ShortCode: "link2", LongURL: "https://example.com/2"},
			}, 5, nil // 5 total
		},
	}

	handler := handlerListLinks(mockLinkService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links?limit=2&offset=0", nil)
	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: userID})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	links := resp["links"].([]interface{})
	if len(links) != 2 {
		t.Errorf("Got %d links, want 2", len(links))
	}

	total := int(resp["total"].(float64))
	if total != 5 {
		t.Errorf("Total = %d, want 5", total)
	}
}

// Test #81: List links returns 401 without auth
func TestHandlerListLinks_Unauthorized(t *testing.T) {
	mockLinkService := newMockLinkService()

	handler := handlerListLinks(mockLinkService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links", nil)
	// No auth

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// Test #82: Delete link returns 204
func TestHandlerDeleteLink_Success(t *testing.T) {
	userID := uint64(1)
	deleteCalled := false

	mockLinkService := &service.MockLinkService{
		DeleteLinkFn: func(ctx context.Context, shortCode string, uid uint64) error {
			deleteCalled = true
			return nil // Success
		},
	}

	handler := handlerDeleteLink(mockLinkService)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/links/abc123", nil)
	req.SetPathValue("shortCode", "abc123")
	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: userID})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusNoContent)
	}

	if !deleteCalled {
		t.Error("DeleteLink was not called")
	}
}

// Test #85: Logout returns 200
func TestHandlerLogout_Success(t *testing.T) {
	logoutCalled := false

	mockAuthService := &auth.MockAuthService{
		LogoutFn: func(ctx context.Context, refreshToken string) error {
			logoutCalled = true
			return nil
		},
	}

	handler := handlerLogout(mockAuthService)

	body := `{"refresh_token": "valid-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	if !logoutCalled {
		t.Error("Logout was not called")
	}
}

// Test #86: Logout returns 401 for invalid token
func TestHandlerLogout_InvalidToken(t *testing.T) {
	mockAuthService := &auth.MockAuthService{
		LogoutFn: func(ctx context.Context, refreshToken string) error {
			return auth.ErrInvalidRefreshToken
		},
	}

	handler := handlerLogout(mockAuthService)

	body := `{"refresh_token": "invalid-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// Test #87: GetGlobalStats returns correct stats
func TestHandlerGetGlobalStats_Success(t *testing.T) {
	mockLinkService := &service.MockLinkService{
		GetGlobalStatsFn: func(ctx context.Context) (*model.GlobalStatsResponse, error) {
			return &model.GlobalStatsResponse{
				TotalLinks:    1500,
				TotalRequests: 45000,
			}, nil
		},
	}

	handler := handlerGetGlobalStats(mockLinkService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/stats", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp model.GlobalStatsResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.TotalLinks != 1500 {
		t.Errorf("TotalLinks = %d, want 1500", resp.TotalLinks)
	}

	if resp.TotalRequests != 45000 {
		t.Errorf("TotalRequests = %d, want 45000", resp.TotalRequests)
	}
}

// Test #88: GetGlobalStats returns 500 on service error
func TestHandlerGetGlobalStats_ServiceError(t *testing.T) {
	mockLinkService := &service.MockLinkService{
		GetGlobalStatsFn: func(ctx context.Context) (*model.GlobalStatsResponse, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := handlerGetGlobalStats(mockLinkService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/stats", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}
