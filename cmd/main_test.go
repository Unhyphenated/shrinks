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
	"github.com/Unhyphenated/shrinks-backend/internal/encoding"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/service"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
)


type MockConfig struct {
    SaveLinkFn func(ctx context.Context, longURL string, userID *uint64) (string, error)
}

func newMockStore(cfg MockConfig) *storage.MockStore {
    // Default implementation for methods not being tested
    defaultGetFn := func(ctx context.Context, shortURL string) (*model.Link, error) { return nil, nil }

    return &storage.MockStore{
        SaveLinkFn:         cfg.SaveLinkFn,
        GetLinkByCodeFn:    defaultGetFn,
        CloseFn:            func() {},
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