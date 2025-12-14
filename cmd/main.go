package main

import (
    "fmt"
    "log"
	"os"
    "net/http"

	"github.com/joho/godotenv"
	"github.com/Unhyphenated/shrinks-backend/internal/storage"
)

func main() {
	err := godotenv.Load() 
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
        log.Fatal("DATABASE_URL environment variable is not set. Cannot connect to Postgres.") 
    }

	store, err := storage.NewPostgresStore(dbURL)
	if err != nil {
        log.Fatalf("Failed to initialize database store: %v", err)
    }

	defer store.Close()
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", func(w http.ResponseWriter, r* http.Request) {
		w.Write([]byte("Shorten URL endpoint TODO"))
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r* http.Request) {
		w.Write([]byte("Redirect endpoint TODO"))
	})

	fmt.Println("Server starting on :8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Application is ready to serve requests.")
}
