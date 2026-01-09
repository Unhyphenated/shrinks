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

	linkService := service.NewLinkService(store, cache)
	authService := auth.NewAuthService(store)

	// Simple HTTP server setup
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/shorten", handlerShorten(linkService))
	mux.HandleFunc("GET /api/v1/{shortCode}", handlerRedirect(linkService))

	mux.HandleFunc("POST /api/v1/register", handlerRegister(authService))
	mux.HandleFunc("POST /api/v1/login", handlerLogin(authService))

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

		shortCode, err := svc.Shorten(r.Context(), req.URL)

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
		path := r.URL.Path
		shortCode := strings.TrimPrefix(path, "/api/v1")

		if shortCode == "" {
			util.WriteError(w, http.StatusBadRequest, "Short URL code is required")
			return
		}

		longURL, err := svc.Redirect(r.Context(), shortCode)
		if err != nil {
			util.WriteError(w, http.StatusNotFound, "Link not found")
			return
		}

		http.Redirect(w, r, longURL, http.StatusFound)
	}
}
