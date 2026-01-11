package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Unhyphenated/shrinks-backend/internal/auth"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/service"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"github.com/Unhyphenated/shrinks-backend/internal/util"
	"github.com/Unhyphenated/shrinks-backend/internal/analytics"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Create PostgresStore
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Cannot connect to Postgres.")
	}

	store, err := storage.NewPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database store: %v", err)
	}

	defer store.Close()

	cache, err := cache.NewRedisCache(redisURL)
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err)
	}
	defer cache.Close()

	analyticsService := analytics.NewAnalyticsService(store)
	linkService := service.NewLinkService(store, cache, analyticsService)
	authService := auth.NewAuthService(store)

	// Simple HTTP server setup
	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/links/shorten", auth.OptionalAuth(handlerShorten(linkService)))
	mux.HandleFunc("GET /api/v1/links/{shortCode}", handlerRedirect(linkService))

	mux.HandleFunc("POST /api/v1/auth/register", handlerRegister(authService))
	mux.HandleFunc("POST /api/v1/auth/login", handlerLogin(authService))

	mux.HandleFunc("POST /api/v1/auth/refresh", handlerRefresh(authService))

	mux.Handle("GET /api/v1/analytics/{shortCode}", auth.RequireAuth(handlerLinkAnalytics(analyticsService, linkService))) // auth guard

	fmt.Println("Server starting on :8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Application is ready to serve requests.")
}

func handlerRegister(svc *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.RegisterRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.WriteError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if req.Email == "" {
			util.WriteError(w, http.StatusBadRequest, "Email is required")
			return
		}

		if req.Password == "" {
			util.WriteError(w, http.StatusBadRequest, "Password is required")
			return
		}

		registerResp, err := svc.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidEmail):
				util.WriteError(w, http.StatusBadRequest, "Invalid email format")
			case errors.Is(err, auth.ErrPasswordTooShort):
				util.WriteError(w, http.StatusBadRequest, "Password must be at least 8 characters")
			case errors.Is(err, auth.ErrPasswordTooLong):
				util.WriteError(w, http.StatusBadRequest, "Password exceeds 72 characters")
			case errors.Is(err, auth.ErrUserAlreadyExists):
				util.WriteError(w, http.StatusConflict, "User already exists")
			default:
				log.Printf("Registration error: %v", err)
				util.WriteError(w, http.StatusInternalServerError, "Failed to register user")
			}
			return
		}

		util.WriteJSON(w, http.StatusCreated, registerResp)
	}
}

func handlerLogin(svc *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.WriteError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if req.Email == "" {
			util.WriteError(w, http.StatusBadRequest, "Email is required")
			return
		}

		if req.Password == "" {
			util.WriteError(w, http.StatusBadRequest, "Password is required")
			return
		}

		authResp, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidEmail):
				util.WriteError(w, http.StatusBadRequest, "Invalid email format")
			case errors.Is(err, auth.ErrInvalidCredentials):
				util.WriteError(w, http.StatusUnauthorized, "Invalid email or password")
			default:
				log.Printf("Login error: %v", err)
				util.WriteError(w, http.StatusInternalServerError, "Failed to login")
			}
			return
		}

		util.WriteJSON(w, http.StatusOK, authResp)
	}
}

func handlerRefresh(svc *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.RefreshTokenRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.WriteError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if req.RefreshToken == "" {
			util.WriteError(w, http.StatusBadRequest, "Refresh token is required")
			return
		}

		refreshResp, err := svc.RefreshAccessToken(r.Context(), req.RefreshToken)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidRefreshToken):
				util.WriteError(w, http.StatusUnauthorized, "Invalid refresh token")
			case errors.Is(err, auth.ErrRefreshTokenExpired):
				util.WriteError(w, http.StatusUnauthorized, "Refresh token has expired")
			default:
				log.Printf("Refresh token error: %v", err)
				util.WriteError(w, http.StatusInternalServerError, "Failed to refresh access token")
			}
			return
		}

		util.WriteJSON(w, http.StatusOK, refreshResp)
	}
}

func handlerShorten(svc *service.LinkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.CreateLinkRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.WriteError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if req.URL == "" {
			util.WriteError(w, http.StatusBadRequest, "URL is required")
			return
		}

		var userID *uint64
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if ok {
			userID = &claims.UserID
		}

		shortCode, err := svc.Shorten(r.Context(), req.URL, userID)

		if err != nil {
			util.WriteError(w, http.StatusInternalServerError, "Failed to shorten URL")
			return
		}

		resp := model.CreateLinkResponse{
			ShortCode: shortCode,
			LongURL:   req.URL,
		}

		util.WriteJSON(w, http.StatusCreated, resp)
	}
}

func handlerRedirect(svc *service.LinkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ua := util.ParseUserAgent(r.Header.Get("User-Agent"))
		ip := util.GetIP(r.Header.Get("X-Forwarded-For"), r.RemoteAddr)
		
		event := &model.AnalyticsEvent{
			IPAddress:  util.AnonymizeIP(ip),
			DeviceType: ua.DeviceType,
			Browser:    ua.Browser,
			OS:         ua.OS,
			UserAgent:  r.Header.Get("User-Agent"),
    	}

		path := r.URL.Path
		shortCode := strings.TrimPrefix(path, "/api/v1/links/")

		if shortCode == "" {
			util.WriteError(w, http.StatusBadRequest, "Short URL code is required")
			return
		}

		longURL, err := svc.Redirect(r.Context(), shortCode, event)
		if err != nil {
			util.WriteError(w, http.StatusNotFound, "Link not found")
			return
		}

		http.Redirect(w, r, longURL, http.StatusFound)
	}
}

// handlerListLinks - GET /api/v1/links
// Returns all links for authenticated user with basic stats
// func handlerListLinks(linkService *service.LinkService, analyticsService *service.AnalyticsService) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// TODO: get userID from JWT token (RequireAuth middleware)
// 		// TODO: call analyticsService.GetUserLinksWithStats(userID)
// 		// TODO: return JSON response with links array
// 	}
// }

// handlerLinkAnalytics - GET /api/v1/links/{shortCode}/analytics?period=30d
// Returns detailed analytics for specific link
func handlerLinkAnalytics(analyticsService *analytics.AnalyticsService, linkService *service.LinkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: extract shortCode from URL path
		path := r.URL.Path
		shortCode := strings.TrimPrefix(path, "/api/v1/analytics/")

		// TODO: get userID from JWT token (RequireAuth middleware)
		var userID *uint64
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if ok {
			userID = &claims.UserID
		}

		// TODO: get link by shortCode
		link, err := linkService.Store.GetLinkByCode(r.Context(), shortCode)
		if err != nil || link == nil {
			util.WriteError(w, http.StatusNotFound, "Link not found")
			return
		}

		// TODO: check link ownership (return 403 if user doesn't own link)
		if link.UserID == nil || userID == nil || *link.UserID != *userID {
			util.WriteError(w, http.StatusForbidden, "Access denied")
			return
		}
		// TODO: parse period query param (default to "30d", validate "24h"/"7d"/"30d") DO THIS LATER
		// TODO: call analyticsService.RetrieveAnalytics(linkID, period)
		// TODO: return JSON response with AnalyticsSummary
	}
}
