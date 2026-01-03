package util

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
		
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
    errorPayload := map[string]string{"error": message}

	WriteJSON(w, status, errorPayload)
}
