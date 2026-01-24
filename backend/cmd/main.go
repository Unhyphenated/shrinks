package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Unhyphenated/shrinks-backend/internal/analytics"
	"github.com/Unhyphenated/shrinks-backend/internal/auth"
	"github.com/Unhyphenated/shrinks-backend/internal/cache"
	"github.com/Unhyphenated/shrinks-backend/internal/model"
	"github.com/Unhyphenated/shrinks-backend/internal/service"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
	"github.com/Unhyphenated/shrinks-backend/internal/util"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Create PostgresStore & RedisCache
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Cannot connect to Postgres.")
	}

	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is not set. Cannot connect to Redis.")
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

	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/links/shorten", auth.OptionalAuth(handlerShorten(linkService)))
	mux.HandleFunc("GET /api/v1/links/{shortCode}", handlerRedirect(linkService))
	mux.Handle("GET /api/v1/links/{shortCode}/analytics", auth.RequireAuth(handlerLinkAnalytics(analyticsService, linkService)))
	mux.Handle("GET /api/v1/links", auth.RequireAuth(handlerListLinks(linkService)))
	mux.Handle("DELETE /api/v1/links/{shortCode}", auth.RequireAuth(handlerDeleteLink(linkService)))
	mux.HandleFunc("GET /api/v1/links/stats", handlerGetGlobalStats(linkService))

	mux.HandleFunc("POST /api/v1/auth/register", handlerRegister(authService))
	mux.HandleFunc("POST /api/v1/auth/login", handlerLogin(authService))
	mux.HandleFunc("POST /api/v1/auth/refresh", handlerRefresh(authService))
	mux.HandleFunc("POST /api/v1/auth/logout", handlerLogout(authService))
	mux.Handle("GET /api/v1/auth/me", auth.RequireAuth(handlerMe()))

	mux.HandleFunc("GET /health", handlerHealth())

	fmt.Println("Server starting on :8080")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handlerCORSMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handlerRegister(svc auth.AuthProvider) http.HandlerFunc {
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

func handlerLogin(svc auth.AuthProvider) http.HandlerFunc {
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

		isProduction := os.Getenv("ENV") == "production"
		sameSite := http.SameSiteLaxMode
		if isProduction {
			sameSite = http.SameSiteStrictMode
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    authResp.RefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   isProduction,
			SameSite: sameSite,
			MaxAge:   3600 * 24 * 7, // 7 days
		})

		util.WriteJSON(w, http.StatusOK, authResp)
	}
}

func handlerLogout(svc auth.AuthProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req model.LogoutRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.WriteError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if req.RefreshToken == "" {
			util.WriteError(w, http.StatusBadRequest, "Refresh token is required")
			return
		}

		err := svc.Logout(r.Context(), req.RefreshToken)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidRefreshToken):
				util.WriteError(w, http.StatusUnauthorized, "Invalid refresh token")
			case errors.Is(err, auth.ErrRefreshTokenExpired):
				util.WriteError(w, http.StatusUnauthorized, "Refresh token has expired")
			default:
				log.Printf("Logout error: %v", err)
				util.WriteError(w, http.StatusInternalServerError, "Failed to logout")
			}
			return
		}
		util.WriteJSON(w, http.StatusOK, "Logged out successfully")
	}
}

func handlerRefresh(svc auth.AuthProvider) http.HandlerFunc {
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

func handlerShorten(svc service.LinkProvider) http.HandlerFunc {
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
			switch {
			case errors.Is(err, service.ErrInvalidURL):
				util.WriteError(w, http.StatusBadRequest, "Invalid URL")
			case errors.Is(err, service.ErrURLScheme):
				util.WriteError(w, http.StatusBadRequest, "Invalid URL scheme")
			case errors.Is(err, service.ErrURLHost):
				util.WriteError(w, http.StatusBadRequest, "Invalid URL host")
			default:
				util.WriteError(w, http.StatusInternalServerError, "Failed to shorten URL")
			}
			return
		}

		resp := model.CreateLinkResponse{
			ShortCode: shortCode,
			LongURL:   req.URL,
		}

		util.WriteJSON(w, http.StatusCreated, resp)
	}
}

func handlerRedirect(svc service.LinkProvider) http.HandlerFunc {
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

		shortCode := r.PathValue("shortCode")

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

func handlerLinkAnalytics(analyticsService analytics.AnalyticsProvider, linkService service.LinkProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			util.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		shortCode := r.PathValue("shortCode")
		period := r.URL.Query().Get("period")
		if period == "" {
			period = "30d"
		}

		link, err := linkService.GetLinkByCode(r.Context(), shortCode)
		if err != nil {
			util.WriteError(w, http.StatusInternalServerError, "Failed to get link")
			return
		}

		if link == nil {
			util.WriteError(w, http.StatusNotFound, "Link not found")
			return
		}

		// Check ownership
		if link.UserID == nil || *link.UserID != claims.UserID {
			util.WriteError(w, http.StatusForbidden, "Not authorized to view analytics for this link")
			return
		}

		analyticsSummary, err := analyticsService.RetrieveAnalytics(r.Context(), link.ID, period)
		if err != nil {
			util.WriteError(w, http.StatusInternalServerError, "Failed to retrieve analytics")
			return
		}
		util.WriteJSON(w, http.StatusOK, analyticsSummary)
	}
}
func handlerListLinks(linkService service.LinkProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			util.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		links, total, err := linkService.GetUserLinks(r.Context(), claims.UserID, 10, 0)
		if err != nil {
			util.WriteError(w, http.StatusInternalServerError, "Failed to list links")
			return
		}
		util.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"links": links,
			"total": total,
		})
	}
}

func handlerDeleteLink(linkService service.LinkProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaimsFromContext(r.Context())
		if !ok {
			util.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		shortCode := r.PathValue("shortCode")
		err := linkService.DeleteLink(r.Context(), shortCode, claims.UserID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrNotOwner):
				util.WriteError(w, http.StatusForbidden, "Not authorized to delete this link")
			case errors.Is(err, service.ErrLinkNotFound):
				util.WriteError(w, http.StatusNotFound, "Link not found")
			case errors.Is(err, storage.ErrNotOwner):
				util.WriteError(w, http.StatusForbidden, "Not authorized to delete this link")
			case errors.Is(err, storage.ErrLinkNotFound):
				util.WriteError(w, http.StatusNotFound, "Link not found")
			default:
				log.Printf("Delete link error: %v", err)
				util.WriteError(w, http.StatusInternalServerError, "Failed to delete link")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handlerCORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := getAllowedOrigins()

		for _, allowed := range allowedOrigins {
			if origin == allowed || allowed == "*" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getAllowedOrigins() []string {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		return []string{"http://localhost:3000, http://localhost:5173"}
	}
	return strings.Split(origins, ",")
}

func handlerGetGlobalStats(linkService service.LinkProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := linkService.GetGlobalStats(r.Context())
		if err != nil {
			log.Printf("Get global stats error: %v", err)
			util.WriteError(w, http.StatusInternalServerError, "Failed to retrieve global stats")
			return
		}
		util.WriteJSON(w, http.StatusOK, stats)
	}
}

func handlerHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, "OK")
	}
}

func handlerMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, _ := auth.GetClaimsFromContext(r.Context())

		// Return the user info based on the ID in the token claims
		util.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"id":    claims.UserID,
			"email": claims.Email,
		})
	}
}
