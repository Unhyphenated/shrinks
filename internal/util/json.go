package util

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
    
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Log the failure, but since we already wrote headers, we can't send a clean error response.
		log.Printf("Error writing JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
    errorPayload := map[string]string{"error": message}
	
	writeJSON(w, status, errorPayload)
}