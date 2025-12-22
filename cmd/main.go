package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

	if dbURL == "" {
        log.Fatal("DATABASE_URL environment variable is not set. Cannot connect to Postgres.") 
    }

	store, err := storage.NewPostgresStore(dbURL)
	if err != nil {
        log.Fatalf("Failed to initialize database store: %v", err)
    }

	defer store.Close()

	linkService := service.NewLinkService(store)

	// Simple HTTP server setup
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten",   handlerShorten(linkService))
	mux.HandleFunc("GET /{shortURL}", handlerRedirect(linkService))

	fmt.Println("Server starting on :8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Application is ready to serve requests.")
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

		shortURL, err := svc.Shorten(r.Context(), req.URL)

		if err != nil {
			   util.WriteError(w, http.StatusInternalServerError, "Failed to shorten URL")
			return
		}

		resp := model.CreateLinkResponse{
			ShortURL: shortURL, 
			LongURL: req.URL,
		}

		util.WriteJSON(w, http.StatusCreated, resp)
	}
}

func handlerRedirect(svc *service.LinkService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := r.URL.Path[1:]

		if shortURL == "" {
			util.WriteError(w, http.StatusBadRequest, "Short URL code is required")
			return
		}

		longURL, err := svc.Redirect(r.Context(), shortURL)
		if err != nil {
			util.WriteError(w, http.StatusNotFound, "Link not found")
			return
		}
		
		http.Redirect(w, r, longURL, http.StatusFound)
	}
}